package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
)

func processWatcher(pid int, checkName string, frequency int, client *api.Client) {
	channel := time.Tick(time.Duration(frequency) * time.Second)
	for range channel {
		_, err := os.FindProcess(pid)
		if err != nil {
			fmt.Println("Process is not running")
			client.Agent().FailTTL(checkName, "Process is not running")
			break
		}
		fmt.Println("Process is running")
		client.Agent().PassTTL(checkName, "Process is running")
	}
}

func getConsulClient(httpAddr string, token string) *api.Client {
	consulClient, err := api.NewClient(&api.Config{
		Address: httpAddr,
		Token:   token,
	})
	if err != nil {
		log.Fatal(err)
	}
	return consulClient
}

func registerConsulService(name string, frequency int, client *api.Client) {
	fmt.Printf("Registering '%s' in consul\n", name)
	client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:   name,
		Name: name,
		Check: &api.AgentServiceCheck{
			CheckID:                        name,
			Name:                           name,
			DeregisterCriticalServiceAfter: strconv.Itoa(2*frequency) + "s",
			TTL:                            strconv.Itoa(3*frequency) + "s",
		},
	})
}

func deregisterConsulService(name string, client *api.Client) {
	fmt.Printf("Deregister '%s' from consul\n", name)
	client.Agent().ServiceDeregister(name)
}

func main() {
	tokenPrt := flag.String("token", "", "Consul token used for registration")
	servicePtr := flag.String("service", "", "Consul Service Name")
	checkFrequencyPtr := flag.Int("frequency", 30, "Health Check Frequency (in seconds)")
	commandPtr := flag.String("command", "", "Command to run")
	argumentsPtr := flag.String("args", "", "String with all arguments")
	flag.Parse()

	cmd := exec.Command(*commandPtr, *argumentsPtr)
	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start: ", err)
		return
	}

	consulClient := getConsulClient("localhost:8500", *tokenPrt)
	registerConsulService(*servicePtr, *checkFrequencyPtr, consulClient)
	go processWatcher(cmd.Process.Pid, *servicePtr, *checkFrequencyPtr, consulClient)

	if err := cmd.Wait(); err != nil {
		exitCode := cmd.ProcessState.ExitCode()
		fmt.Printf("Process stopped running. Error: '%s' Exit code: %d\n", err.Error(), exitCode)
		deregisterConsulService(*servicePtr, consulClient)
		os.Exit(exitCode)
	}

	fmt.Println("Process exited. Exit code: ", cmd.ProcessState.ExitCode())
	deregisterConsulService(*servicePtr, consulClient)
}
