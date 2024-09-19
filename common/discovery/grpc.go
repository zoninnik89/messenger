package discovery

import (
	"context"
	"log"
	"math/rand"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ServiceConnection(ctx context.Context, serviceName string, registry Registry) (*grpc.ClientConn, error) {
	addrs, err := registry.Discover(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	log.Printf("Discovered %d instances of %s", len(addrs), serviceName)

	// Randomly select an instance
	return grpc.NewClient(
		addrs[rand.Intn(len(addrs))],
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}
