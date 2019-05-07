package runner

import (
	"time"

	"github.com/hashicorp/consul/api"
)

type Config struct {
	ConsulToken        string
	CommandLine        []string
	Definition         *api.AgentServiceRegistration
	SelfCheckFrequency time.Duration
}
