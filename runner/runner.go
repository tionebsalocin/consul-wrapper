package runner

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/consul/api"
)

const selfCheckName string = "SelfCheck"

func processWatcher(pid int, config Config, client *api.Client) {
	channel := time.Tick(config.SelfCheckFrequency)
	for range channel {
		_, err := os.FindProcess(pid)
		if err != nil {
			log.Println("[ConsulWrapper] Process is not running")
			err := client.Agent().FailTTL(config.Definition.Name+"-"+selfCheckName, "Process is not running")
			if err != nil {
				log.Println("[ConsulWrapper] Failed to send FAIL health check", err)
			}
			break
		}
		log.Println("[ConsulWrapper] Process is running")
		err = client.Agent().PassTTL(config.Definition.Name+"-"+selfCheckName, "Process is running")
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

func registerConsulService(config Config, client *api.Client) {
	log.Printf("[ConsulWrapper] Registering '%s' in consul\n", config.Definition.Name)
	err := client.Agent().ServiceRegister(config.Definition)
	if err != nil {
		log.Fatal(err)
	}
}

func deregisterConsulService(config Config, client *api.Client) {
	log.Printf("[ConsulWrapper] Deregistering '%s' from consul\n", config.Definition.Name)
	err := client.Agent().ServiceDeregister(config.Definition.Name)
	if err != nil {
		log.Fatal(err)
	}
}

func Run(config Config) {
	config.Definition.Checks = append(config.Definition.Checks, &api.AgentServiceCheck{
		CheckID:                        config.Definition.Name + "-" + selfCheckName,
		Name:                           selfCheckName,
		DeregisterCriticalServiceAfter: fmt.Sprintf("%d", int64(2.1*config.SelfCheckFrequency.Seconds())) + "s",
		TTL:                            fmt.Sprintf("%d", int64(3.1*config.SelfCheckFrequency.Seconds())) + "s",
	})
	consulClient := getConsulClient("localhost:8500", config.ConsulToken)
	registerConsulService(config, consulClient)

	cmd := exec.Command(config.CommandLine[0], config.CommandLine[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Println("[ConsulWrapper] Failed to start: ", err)
		return
	}

	go processWatcher(cmd.Process.Pid, config, consulClient)

	signalWatch(cmd)
	/*
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-signals
			syscall.Kill(cmd.Process.Pid, sig.(syscall.Signal))
			cmd.Process.Kill()
		}()
	*/
	if err := cmd.Wait(); err != nil {
		exitCode := cmd.ProcessState.ExitCode()
		log.Println("[ConsulWrapper] Error: ", err.Error())
		log.Println("[ConsulWrapper] Process stopped running. Exit code: ", exitCode)
		deregisterConsulService(config, consulClient)
		os.Exit(exitCode)
	}

	log.Println("[ConsulWrapper] Process exited. Exit code: ", cmd.ProcessState.ExitCode())
	deregisterConsulService(config, consulClient)
}
