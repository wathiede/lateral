// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
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
	"github.com/akramer/lateral/client"
	"github.com/akramer/lateral/server"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// killCmd represents the kill command
var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kill the server with fire",
	Long: `Send a SIGKILL to the server's process group.
This should kill the server, and any subprocesses that have not changed their process group.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.NewUnixConn(Viper)
		defer c.Close()
		if err != nil {
			glog.Errorln("Error connecting to server:", err)
			ExitCode = 1
			return
		}
		req := &server.Request{
			Type: server.REQUEST_KILL,
		}
		err = client.SendRequest(c, req)
		if err != nil {
			glog.Errorln("Error sending request", err)
			ExitCode = 1
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(killCmd)

}
