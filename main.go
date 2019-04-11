package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/consul/api"
)

func processWatcher(pid int, checkName string, frequency time.Duration, client *api.Client) {
	channel := time.Tick(frequency)
	for range channel {
		_, err := os.FindProcess(pid)
		if err != nil {
			log.Println("[ConsulWrapper] Process is not running")
			err := client.Agent().FailTTL(checkName, "Process is not running")
			if err != nil {
				log.Println("[ConsulWrapper] Failed to send FAIL health check", err)
			}
			break
		}
		log.Println("[ConsulWrapper] Process is running")
		err = client.Agent().PassTTL(checkName, "Process is running")
		if err != nil {
			log.Println("[ConsulWrapper] Failed to send PASSING health check", err)
		}
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

func registerConsulService(name string, frequency time.Duration, client *api.Client) {
	log.Printf("[ConsulWrapper] Registering '%s' in consul\n", name)
	err := client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:   name,
		Name: name,
		Check: &api.AgentServiceCheck{
			CheckID:                        name,
			Name:                           name,
			DeregisterCriticalServiceAfter: fmt.Sprintf("%d", int64(2.1*frequency.Seconds())) + "s",
			TTL:                            fmt.Sprintf("%d", int64(3.1*frequency.Seconds())) + "s",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func deregisterConsulService(name string, client *api.Client) {
	log.Printf("[ConsulWrapper] Deregistering '%s' from consul\n", name)
	err := client.Agent().ServiceDeregister(name)
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\t%s [options] -service <ServiceName> <Command> [Args]\n\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	tokenPrt := flag.String("token", "", "Consul token used for registration")
	servicePtr := flag.String("service", "", "Consul Service Name")
	checkFrequencyPtr := flag.Duration("frequency", time.Duration(30), "Health Check Frequency (in seconds)")

	flag.Parse()
	if *servicePtr == "" {
		usage()
		os.Exit(1)
	}
	arguments := flag.Args()

	consulClient := getConsulClient("localhost:8500", *tokenPrt)
	registerConsulService(*servicePtr, *checkFrequencyPtr, consulClient)

	cmd := exec.Command(arguments[0], arguments[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Println("[ConsulWrapper] Failed to start: ", err)
		return
	}

	go processWatcher(cmd.Process.Pid, *servicePtr, *checkFrequencyPtr, consulClient)

	if err := cmd.Wait(); err != nil {
		exitCode := cmd.ProcessState.ExitCode()
		log.Println("[ConsulWrapper] Error: ", err.Error())
		log.Println("[ConsulWrapper] Process stopped running. Exit code: ", exitCode)
		deregisterConsulService(*servicePtr, consulClient)
		os.Exit(exitCode)
	}

	log.Println("[ConsulWrapper] Process exited. Exit code: ", cmd.ProcessState.ExitCode())
	deregisterConsulService(*servicePtr, consulClient)
}
