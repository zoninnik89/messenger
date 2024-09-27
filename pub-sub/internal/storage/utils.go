package storage

import (
	"fmt"
	"github.com/zoninnik89/messenger/pub-sub/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

// Hashset implementation

type HashSet struct {
	Store map[*types.Client]struct{}
}

// NewHashSet creates a new empty HashSet.
func NewHashSet() *HashSet {
	return &HashSet{
		Store: make(map[*types.Client]struct{}),
	}
}

// Add inserts a value into the set.
func (s *HashSet) Add(value *types.Client) {
	s.Store[value] = struct{}{} // Empty struct takes no space.
}

// Remove deletes a value from the set.
func (s *HashSet) Remove(value *types.Client) {
	delete(s.Store, value)
}

// Contains checks if a value is in the set.
func (s *HashSet) Contains(value *types.Client) bool {
	_, exists := s.Store[value]
	return exists
}

func (s *HashSet) Size() int {
	return len(s.Store)
}

// Async Map

type AsyncMap struct {
	store sync.Map
}

func NewAsyncMap() *AsyncMap {
	return &AsyncMap{}
}

// Add a value to the slice at the given key
func (m *AsyncMap) Add(key string, value *types.Client) {
	// Use LoadOrStore to get or initialize the slice
	chat, _ := m.store.LoadOrStore(key, NewHashSet())

	// Use a mutex to protect appending to the slice
	mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	chat.(*HashSet).Add(value)
}

// Get the slice of values for the given key
func (m *AsyncMap) Get(key string) (*HashSet, error) {
	var op = "storage.Get"

	chat, ok := m.store.Load(key)
	if !ok {
		return &HashSet{}, fmt.Errorf("%s: %w", op, status.Error(codes.NotFound, "key not found"))
	}

	// Return a copy of the slice to avoid race conditions
	return chat.(*HashSet), nil
}

func (m *AsyncMap) Remove(key string, client *types.Client) error {
	var op = "storage.Remove"

	chat, ok := m.store.Load(key)
	if !ok {
		return fmt.Errorf("%s: %w", op, status.Error(codes.NotFound, "client not found"))
	}

	chat.(*HashSet).Remove(client)
	return nil
}
