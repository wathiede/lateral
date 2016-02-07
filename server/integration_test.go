package server_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/akramer/lateral/client"
	"github.com/akramer/lateral/server"
	"github.com/spf13/viper"
)

var tempDir string

func init() {
	var err error
	// TODO: Delete this tempdir after the test finishes.
	tempDir, err = ioutil.TempDir("", "lateralTest")
	if err != nil {
		panic(err)
	}
}

func makeTestViper() *viper.Viper {
	v := viper.New()
	v.Set("socket", tempDir+"/socket")
	return v
}

func TestRunConnection(t *testing.T) {
	v := makeTestViper()
	l, err := server.NewUnixListener(v)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go server.Run(v, l)
	c, err := client.NewUnixConn(v)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	req := &server.Request{
		Type: server.REQUEST_GETPID,
	}
	err = client.SendRequest(c, req)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.ReceiveResponse(c)
	if err != nil {
		t.Fatal("got error", err)
	} else if resp.Type == server.RESPONSE_ERR {
		t.Fatal("got error", resp.Message)
	}
	if resp.Getpid.Pid != os.Getpid() {
		t.Error("Pid didn't match")
	}

	req.Fds = []int{0, 1, 2}
	req.HasFds = true
	req.Type = server.REQUEST_RUN
	req.Run = &server.RequestRun{
		Args: []string{"/bin/echo", "foo"},
		Env:  os.Environ(),
	}

	err = client.SendRequest(c, req)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = client.ReceiveResponse(c)
	if err != nil {
		t.Fatal("got error", err)
	} else if resp.Type == server.RESPONSE_ERR {
		t.Fatal("got error", resp.Message)
	}
	if resp.Getpid.Pid != os.Getpid() {
		t.Error("Pid didn't match")
	}
}
