package discovery

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type Registry interface {
	Register(ctx context.Context, instanceID, host string, serviceName string, hostPort int) error
	Deregister(ctx context.Context, instanceID string) error
	Discover(ctx context.Context, serviceName string) ([]string, error)
	HealthCheck(instanceID string) error
}

func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
