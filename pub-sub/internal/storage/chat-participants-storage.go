package storage

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

// Hashset implementation

type HashSet struct {
	Store map[string]struct{}
}

// NewHashSet creates a new empty HashSet.
func NewHashSet() *HashSet {
	return &HashSet{
		Store: make(map[string]struct{}),
	}
}

// Add inserts a value into the set.
func (s *HashSet) Add(value string) {
	s.Store[value] = struct{}{} // Empty struct takes no space.
}

// Remove deletes a value from the set.
func (s *HashSet) Remove(value string) {
	delete(s.Store, value)
}

// Contains checks if a value is in the set.
func (s *HashSet) Contains(value string) bool {
	_, exists := s.Store[value]
	return exists
}

func (s *HashSet) Size() int {
	return len(s.Store)
}

// Async Map

type ChatParticipantsStorage struct {
	store sync.Map
}

func NewChatParticipantsStorage() *ChatParticipantsStorage {
	return &ChatParticipantsStorage{}
}

// Add a value to the slice at the given key
func (m *ChatParticipantsStorage) Add(chatID string, userID string) {
	// Use LoadOrStore to get or initialize the slice
	chat, _ := m.store.LoadOrStore(chatID, NewHashSet())

	// Use a mutex to protect appending to the slice
	mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	chat.(*HashSet).Add(userID)
}

// Get the slice of values for the given key
func (m *ChatParticipantsStorage) Get(chatID string) (*HashSet, error) {
	var op = "storage.Get"

	chat, ok := m.store.Load(chatID)
	if !ok {
		return &HashSet{}, fmt.Errorf("%s: %w", op, status.Error(codes.NotFound, "key not found"))
	}

	// Return a copy of the slice to avoid race conditions
	return chat.(*HashSet), nil
}

func (m *ChatParticipantsStorage) Remove(chatID string, userID string) error {
	var op = "storage.Remove"

	chat, ok := m.store.Load(chatID)
	if !ok {
		return fmt.Errorf("%s: %w", op, status.Error(codes.NotFound, "client not found"))
	}

	chat.(*HashSet).Remove(userID)
	return nil
}
