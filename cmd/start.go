// Copyright © 2016 Adam Kramer <akramer@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"
	"syscall"

	"github.com/akramer/lateral/client"
	"github.com/akramer/lateral/server"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

const MAGICENV = "LAT_MAGIC"

func realStart(cmd *cobra.Command, args []string) error {
	err := syscall.Setpgid(0, 0)
	if err != nil {
		glog.Errorln("Error setting process group ID")
		return err
	}
	os.Remove(Viper.GetString("socket"))
	l, err := server.NewUnixListener(Viper)
	defer l.Close()
	if err != nil {
		glog.Errorln("Error opening listening socket:", err)
		return err
	}
	server.Run(Viper, l)
	os.Remove(Viper.GetString("socket"))
	return nil
}

func forkMyself() error {
	os.Setenv(MAGICENV, Viper.GetString("socket"))
	attr := &syscall.ProcAttr{
		Dir:   "/",
		Env:   os.Environ(),
		Files: []uintptr{0, 1, 2}}
	_, err := syscall.ForkExec("/proc/self/exe", os.Args, attr)
	return err
}

func isRunning() bool {
	c, err := client.NewUnixConn(Viper)
	defer c.Close()
	if err == nil {
		return true
	}
	return false
}

func runStart(cmd *cobra.Command, args []string) {
	// If MAGICENV is set to the socket path, we can be (relatively) sure we're the child process.
	if Viper.GetBool("start.foreground") || os.Getenv(MAGICENV) == Viper.GetString("socket") {
		glog.Infoln("Not forking a child server")
		err := realStart(cmd, args)
		if err != nil {
			ExitCode = 1
			return
		}
	} else {
		if Viper.GetBool("new_server") && isRunning() {
			glog.Errorln("Server already running and new_server specified.")
			ExitCode = 1
			return
		}
		err := forkMyself()
		if err != nil {
			glog.Errorln("Error forking subprocess: ", err)
			ExitCode = 1
			return
		}
	}
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the lateral background server",
	Long: `Start the lateral background server. By default, this creates a new server
for every session. This essentially means each login shell will have its own
server.`,
	Run: runStart,
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP("new_server", "n", false, "Print an error and return a non-zero status if the server is already running")
	Viper.BindPFlag("start.new_server", startCmd.Flags().Lookup("new_server"))
	startCmd.Flags().BoolP("foreground", "f", false, "Do not fork off a background server: run in the foreground.")
	Viper.BindPFlag("start.foreground", startCmd.Flags().Lookup("foreground"))
	startCmd.Flags().IntP("concurrency", "c", 10, "Number of concurrent tasks to run")
	Viper.BindPFlag("start.concurrency", startCmd.Flags().Lookup("concurrency"))
}