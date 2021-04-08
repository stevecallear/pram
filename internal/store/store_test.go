package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stevecallear/pram/internal/assert"
	"github.com/stevecallear/pram/internal/store"
)

func TestInMemoryStore_GetOrSetTopicARN(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*store.InMemoryStore)
		key     string
		valueFn func() (string, error)
		exp     string
		err     bool
	}{
		{
			name: "should return the value if the key exists",
			setup: func(m *store.InMemoryStore) {
				m.GetOrSetTopicARN(context.Background(), "topic-name", func() (string, error) {
					return "expected", nil
				})
			},
			valueFn: func() (string, error) {
				return "not expected", nil
			},
			key: "topic-name",
			exp: "expected",
		},
		{
			name:  "should set the value if the key does not exist",
			setup: func(m *store.InMemoryStore) {},
			valueFn: func() (string, error) {
				return "expected", nil
			},
			key: "topic-name",
			exp: "expected",
		},
		{
			name:  "should return value fn errors",
			setup: func(m *store.InMemoryStore) {},
			valueFn: func() (string, error) {
				return "", errors.New("error")
			},
			key: "topic-name",
			err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := new(store.InMemoryStore)
			tt.setup(sut)

			act, err := sut.GetOrSetTopicARN(context.Background(), tt.key, tt.valueFn)
			assert.ErrorExists(t, err, tt.err)

			if act != tt.exp {
				t.Errorf("got %s, expected %s", act, tt.exp)
			}
		})
	}
}

func TestInMemoryStore_GetOrSetQueueURL(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*store.InMemoryStore)
		key     string
		valueFn func() (string, error)
		exp     string
		err     bool
	}{
		{
			name: "should return the value if the key exists",
			setup: func(m *store.InMemoryStore) {
				m.GetOrSetQueueURL(context.Background(), "queue-name", func() (string, error) {
					return "expected", nil
				})
			},
			valueFn: func() (string, error) {
				return "not expected", nil
			},
			key: "queue-name",
			exp: "expected",
		},
		{
			name:  "should set the value if the key does not exist",
			setup: func(m *store.InMemoryStore) {},
			valueFn: func() (string, error) {
				return "expected", nil
			},
			key: "queue-name",
			exp: "expected",
		},
		{
			name:  "should return value fn errors",
			setup: func(m *store.InMemoryStore) {},
			valueFn: func() (string, error) {
				return "", errors.New("error")
			},
			key: "queue-name",
			err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := new(store.InMemoryStore)
			tt.setup(sut)

			act, err := sut.GetOrSetQueueURL(context.Background(), tt.key, tt.valueFn)
			assert.ErrorExists(t, err, tt.err)

			if act != tt.exp {
				t.Errorf("got %s, expected %s", act, tt.exp)
			}
		})
	}
}
