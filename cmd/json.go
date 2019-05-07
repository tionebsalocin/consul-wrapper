package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"runtime"

	"github.com/hashicorp/consul/api"
	"github.com/spf13/cobra"
	"github.com/tionebsalocin/consul-wrapper/runner"
)

var config *api.AgentServiceRegistration
var configFile string

func init() {
	RootCmd.AddCommand(jsonCmd)
	jsonCmd.Flags().StringVarP(&configFile, "file", "j", "", "json config file")
}

var jsonCmd = &cobra.Command{
	Use:   "json -j <config file> 'command line'",
	Short: "Create a consul service based on config file",
	Long:  "Create a consul service based on config file",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := ioutil.ReadFile(configFile)
		config = &api.AgentServiceRegistration{}
		err := json.Unmarshal([]byte(file), &config)
		if err != nil {
			log.Fatal(err)
		}
		shell := "sh"
		shellFlag := "-c"
		if runtime.GOOS == "windows" {
			shell = "cmd"
			shellFlag = "/c"
		}
		runner.Run(runner.Config{
			ConsulToken:        token,
			CommandLine:        []string{shell, shellFlag, args[0]},
			SelfCheckFrequency: frequency,
			Definition:         config,
		})
	},
}
