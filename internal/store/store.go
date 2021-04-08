package store

import (
	"context"
	"sync"
)

// InMemoryStore represents an in-memory store
type InMemoryStore struct {
	items map[string]string
	mu    sync.RWMutex
}

// GetOrSetTopicARN returns the requested topic arn, or sets it if it does not exist
func (s *InMemoryStore) GetOrSetTopicARN(ctx context.Context, topicName string, fn func() (string, error)) (string, error) {
	return s.getOrSet("topic:"+topicName, fn)
}

// GetOrSetQueueURL returns the requested queue url, or sets it if it does not exist
func (s *InMemoryStore) GetOrSetQueueURL(ctx context.Context, queueName string, fn func() (string, error)) (string, error) {
	return s.getOrSet("queue:"+queueName, fn)
}

func (s *InMemoryStore) getOrSet(key string, fn func() (string, error)) (string, error) {
	v, ok := s.get(key)
	if ok {
		return v, nil
	}

	v, err := fn()
	if err != nil {
		return "", err
	}

	s.set(key, v)
	return v, nil
}

func (s *InMemoryStore) get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.items == nil {
		return "", false
	}

	v, ok := s.items[key]
	return v, ok
}

func (s *InMemoryStore) set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.items == nil {
		s.items = make(map[string]string)
	}

	s.items[key] = value
}
