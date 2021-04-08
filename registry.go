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
		service *aws.Service
		store   Store
		topic   TopicOptions
		queue   QueueOptions
	}

	// RegistryOptions represents a set of registry options
	RegistryOptions struct {
		Store Store
		Topic TopicOptions
		Queue QueueOptions
	}

	// TopicOptions represents a set of topic options
	TopicOptions struct {
		NameFn func(proto.Message) string
	}

	// QueueOptions represents a set of queue options
	QueueOptions struct {
		NameFn          func(proto.Message) string
		ErrorNameFn     func(proto.Message) string
		MaxReceiveCount int
	}
)

var defaultRegistryOptions = RegistryOptions{
	Topic: TopicOptions{
		NameFn: func(m proto.Message) string {
			return MessageName(m)
		},
	},
	Queue: QueueOptions{
		NameFn: func(m proto.Message) string {
			return MessageName(m)
		},
		ErrorNameFn: func(m proto.Message) string {
			return MessageName(m) + "_error"
		},
		MaxReceiveCount: 5,
	},
}

// NewRegistry returns a new registry
func NewRegistry(snsc SNS, sqsc SQS, optFns ...func(*RegistryOptions)) *Registry {
	o := defaultRegistryOptions
	for _, fn := range optFns {
		fn(&o)
	}

	if o.Store == nil {
		o.Store = new(store.InMemoryStore)
	}

	return &Registry{
		service: aws.NewService(snsc, sqsc, Logf),
		store:   o.Store,
		topic:   o.Topic,
		queue:   o.Queue,
	}
}

// TopicARN returns the topic arn for the specified message, or registers it if it does not exist
func (r *Registry) TopicARN(ctx context.Context, m proto.Message) (string, error) {
	tn := r.topic.NameFn(m)
	return r.store.GetOrSetTopicARN(ctx, tn, func() (string, error) {
		res, err := r.service.EnsureTopic(ctx, aws.EnsureTopicRequest{
			TopicName: tn,
		})
		if err != nil {
			return "", err
		}

		return res.TopicARN, nil
	})
}

// QueueURL returns the queue url for the specified message, or registers it if it does not exist
func (r *Registry) QueueURL(ctx context.Context, m proto.Message) (string, error) {
	tn := r.topic.NameFn(m)
	ta, err := r.store.GetOrSetTopicARN(ctx, tn, func() (string, error) {
		res, err := r.service.EnsureTopic(ctx, aws.EnsureTopicRequest{
			TopicName: tn,
		})
		if err != nil {
			return "", err
		}

		return res.TopicARN, nil
	})
	if err != nil {
		return "", err
	}

	qn := r.queue.NameFn(m)
	return r.store.GetOrSetQueueURL(ctx, qn, func() (string, error) {
		res, err := r.service.EnsureSubscription(ctx, aws.EnsureSubscriptionRequest{
			TopicARN:        ta,
			QueueName:       qn,
			ErrorQueueName:  r.queue.ErrorNameFn(m),
			MaxReceiveCount: r.queue.MaxReceiveCount,
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
		o.Topic.NameFn = func(m proto.Message) string {
			return fmt.Sprintf("%s-%s", stage, MessageName(m))
		}
		o.Queue.NameFn = func(m proto.Message) string {
			return fmt.Sprintf("%s-%s-%s", stage, service, MessageName(m))
		}
		o.Queue.ErrorNameFn = func(m proto.Message) string {
			return fmt.Sprintf("%s-%s-%s_error", stage, service, MessageName(m))
		}
	}
}
