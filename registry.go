package pram

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/stevecallear/pram/internal/aws"
	"github.com/stevecallear/pram/internal/store"
)

type (
	// Store represents a key value store
	Store interface {
		GetOrSetTopicARN(ctx context.Context, topicName string, fn func() (string, error)) (string, error)
		GetOrSetQueueURL(ctx context.Context, queueName string, fn func() (string, error)) (string, error)
	}

	// Registry represents an infrastructure registry
	Registry struct {
		service          *aws.Service
		store            Store
		topicNameFn      func(proto.Message) string
		queueNameFn      func(proto.Message) string
		errorQueueNameFn func(proto.Message) string
	}

	// RegistryOptions represents a set of registry options
	RegistryOptions struct {
		TopicNameFn      func(proto.Message) string
		QueueNameFn      func(proto.Message) string
		ErrorQueueNameFn func(proto.Message) string
		Store            Store
	}
)

// NewRegistry returns a new registry
func NewRegistry(snsc SNS, sqsc SQS, optFns ...func(*RegistryOptions)) *Registry {
	opts := RegistryOptions{
		TopicNameFn: func(m proto.Message) string {
			return MessageName(m)
		},
		QueueNameFn: func(m proto.Message) string {
			return MessageName(m)
		},
		ErrorQueueNameFn: func(m proto.Message) string {
			return MessageName(m) + "_error"
		},
	}
	for _, fn := range optFns {
		fn(&opts)
	}

	if opts.Store == nil {
		opts.Store = new(store.InMemoryStore)
	}

	return &Registry{
		service:          aws.NewService(snsc, sqsc, Logf),
		store:            opts.Store,
		topicNameFn:      opts.TopicNameFn,
		queueNameFn:      opts.QueueNameFn,
		errorQueueNameFn: opts.ErrorQueueNameFn,
	}
}

// TopicARN returns the topic arn for the specified message, or registers it if it does not exist
func (r *Registry) TopicARN(ctx context.Context, m proto.Message) (string, error) {
	return r.store.GetOrSetTopicARN(ctx, r.topicNameFn(m), func() (string, error) {
		res, err := r.service.EnsureTopic(ctx, aws.EnsureTopicRequest{
			TopicName: r.topicNameFn(m),
		})
		if err != nil {
			return "", err
		}

		return res.TopicARN, nil
	})
}

// QueueURL returns the queue url for the specified message, or registers it if it does not exist
func (r *Registry) QueueURL(ctx context.Context, m proto.Message) (string, error) {
	ta, err := r.store.GetOrSetTopicARN(ctx, r.topicNameFn(m), func() (string, error) {
		res, err := r.service.EnsureTopic(ctx, aws.EnsureTopicRequest{
			TopicName: r.topicNameFn(m),
		})
		if err != nil {
			return "", err
		}

		return res.TopicARN, nil
	})
	if err != nil {
		return "", err
	}

	return r.store.GetOrSetQueueURL(ctx, r.queueNameFn(m), func() (string, error) {
		res, err := r.service.EnsureSubscription(ctx, aws.EnsureSubscriptionRequest{
			TopicARN:        ta,
			QueueName:       r.queueNameFn(m),
			ErrorQueueName:  r.errorQueueNameFn(m),
			MaxReceiveCount: 5, // todo: config
		})
		if err != nil {
			return "", err
		}

		return res.QueueURL, nil
	})
}

// WithStore configures the registry to use the specified store
func WithStore(s Store) func(*RegistryOptions) {
	return func(o *RegistryOptions) {
		o.Store = s
	}
}

// WithPrefixNaming configures the registry to use prefix naming to support complex message routing
// It applies the following format, assuming a protobuf type name of package.Message:
//  topic: stage-package-Message
//  queue: stage-service-package-Message
//  error: stage-service-package-Message_error
func WithPrefixNaming(stage, service string) func(*RegistryOptions) {
	return func(o *RegistryOptions) {
		o.TopicNameFn = func(m proto.Message) string {
			return fmt.Sprintf("%s-%s", stage, MessageName(m))
		}
		o.QueueNameFn = func(m proto.Message) string {
			return fmt.Sprintf("%s-%s-%s", stage, service, MessageName(m))
		}
		o.ErrorQueueNameFn = func(m proto.Message) string {
			return fmt.Sprintf("%s-%s-%s_error", stage, service, MessageName(m))
		}
	}
}
