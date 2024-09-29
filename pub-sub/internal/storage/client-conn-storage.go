package storage

import (
	"fmt"
	pb "github.com/zoninnik89/messenger/common/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

// Async Map

type ClientConnStorage struct {
	store sync.Map
}

func NewClientConnStorage() *ClientConnStorage {
	return &ClientConnStorage{}
}

// Add a value to the slice at the given key
func (m *ClientConnStorage) Add(key string, value chan *pb.Message) {

	m.store.Store(key, value)

}

// Get the slice of values for the given key
func (m *ClientConnStorage) Get(key string) (chan *pb.Message, error) {
	var op = "storage.Get"

	channel, ok := m.store.Load(key)
	if !ok {
		return nil, fmt.Errorf("%s: %w", op, status.Error(codes.NotFound, "key not found"))
	}

	// Return a copy of the slice to avoid race conditions
	return channel.(chan *pb.Message), nil
}

func (m *ClientConnStorage) Remove(key string) error {
	var op = "storage.Remove"

	_, ok := m.store.Load(key)
	if !ok {
		return fmt.Errorf("%s: %w", op, status.Error(codes.NotFound, "client not found"))
	}

	m.store.Delete(key)
	return nil
}
