package main

import (
	"sync"
)

// Hashset implementation

type HashSet struct {
	store map[*Client]struct{}
}

// NewHashSet creates a new empty HashSet.
func NewHashSet() *HashSet {
	return &HashSet{
		store: make(map[*Client]struct{}),
	}
}

// Add inserts a value into the set.
func (s *HashSet) Add(value *Client) {
	s.store[value] = struct{}{} // Empty struct takes no space.
}

// Remove deletes a value from the set.
func (s *HashSet) Remove(value *Client) {
	delete(s.store, value)
}

// Contains checks if a value is in the set.
func (s *HashSet) Contains(value *Client) bool {
	_, exists := s.store[value]
	return exists
}

func (s *HashSet) Size() int {
	return len(s.store)
}

// Async Map

type AsyncMap struct {
	store sync.Map
}

func NewAsyncMap() *AsyncMap {
	return &AsyncMap{}
}

// Add a value to the slice at the given key
func (m *AsyncMap) Add(key string, value *Client) {
	// Use LoadOrStore to get or initialize the slice
	chat, _ := m.store.LoadOrStore(key, NewHashSet())

	// Use a mutex to protect appending to the slice
	mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	chat.(*HashSet).Add(value)
}

// Get the slice of values for the given key
func (m *AsyncMap) Get(key string) *HashSet {
	chat, ok := m.store.Load(key)
	if !ok {
		return &HashSet{}
	}

	// Return a copy of the slice to avoid race conditions
	return chat.(*HashSet)
}

func (m *AsyncMap) Remove(key string, client *Client) {
	chat, ok := m.store.Load(key)
	if !ok {
		return
	}

	chat.(*HashSet).Remove(client)
}
