package consul

import (
	"context"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"log"
	"strconv"
)

type Registry struct {
	client *consul.Client
}

func NewRegistry(host string, port int) (*Registry, error) {
	config := consul.DefaultConfig()
	config.Address = host + ":" + strconv.Itoa(port)

	client, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Registry{client}, nil
}

func (registry *Registry) Register(
	ctx context.Context,
	instanceID string,
	address string,
	hostPort int,
	serviceName string,
) error {

	return registry.client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		ID:      instanceID,
		Address: address,
		Port:    hostPort,
		Name:    serviceName,
		Check: &consul.AgentServiceCheck{
			CheckID:                        instanceID,
			TLSSkipVerify:                  true,
			TTL:                            "5s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "10s",
		},
	})
}

func (registry *Registry) Deregister(
	ctx context.Context,
	instanceID string,
) error {

	log.Printf("Deregistering service %s", instanceID)
	return registry.client.Agent().CheckDeregister(instanceID)
}

func (registry *Registry) HealthCheck(instanceID string) error {
	return registry.client.Agent().UpdateTTL(instanceID, "online", consul.HealthPassing)
}

func (registry *Registry) Discover(ctx context.Context, serviceName string) ([]string, error) {
	entries, _, err := registry.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}

	var instances []string
	for _, entry := range entries {
		instances = append(instances, fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port))
	}

	return instances, nil
}
