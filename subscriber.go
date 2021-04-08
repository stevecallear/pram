package pram

import (
	"context"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/proto"
)

type (
	// Handler represents a message handler
	Handler interface {
		Message() proto.Message
		Handle(ctx context.Context, m proto.Message, md Metadata) error
	}

	// Subscriber represents a subscriber
	Subscriber struct {
		client                   SQS
		queueURLFn               func(context.Context, proto.Message) (string, error)
		errorFn                  func(error)
		maxNumberOfMessages      int
		receiveInterval          time.Duration
		waitTimeSeconds          int
		visibilityTimeoutSeconds int
	}

	// SubscriberOptions represents a set of subscriber options
	SubscriberOptions struct {
		QueueURLFn               func(context.Context, proto.Message) (string, error)
		ErrorFn                  func(error)
		MaxNumberOfMessages      int
		ReceiveInterval          time.Duration
		WaitTimeSeconds          int
		VisibilityTimeoutSeconds int
	}
)

// NewSubscriber returns a new subscriber
func NewSubscriber(client SQS, optFns ...func(*SubscriberOptions)) *Subscriber {
	opts := SubscriberOptions{
		QueueURLFn: func(context.Context, proto.Message) (string, error) {
			return "", errors.New("queue not found")
		},
		ErrorFn: func(error) {
			// discard errors by default
		},
		MaxNumberOfMessages:      10,
		ReceiveInterval:          time.Second,
		WaitTimeSeconds:          20,
		VisibilityTimeoutSeconds: 15,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	return &Subscriber{
		client:                   client,
		queueURLFn:               opts.QueueURLFn,
		errorFn:                  opts.ErrorFn,
		maxNumberOfMessages:      opts.MaxNumberOfMessages,
		waitTimeSeconds:          opts.WaitTimeSeconds,
		receiveInterval:          opts.ReceiveInterval,
		visibilityTimeoutSeconds: opts.VisibilityTimeoutSeconds,
	}
}

// Subscribe subscribes listens to messages for the specified handler
func (s *Subscriber) Subscribe(ctx context.Context, h Handler) error {
	q, err := s.queueURLFn(ctx, h.Message())
	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		defer wg.Done()

		rt := time.NewTicker(s.receiveInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-rt.C:
				msgs, err := s.receiveMessages(ctx, q)
				if err != nil {
					s.errorFn(err)
				}

				for _, msg := range msgs {
					wg.Add(1)
					go func(msg types.Message) {
						defer wg.Done()

						err := s.handleMessage(ctx, q, msg, h)
						if err != nil {
							s.errorFn(err)
						}
					}(msg)
				}
			}
		}
	}()

	wg.Wait()
	return nil
}

func (s *Subscriber) receiveMessages(ctx context.Context, queueURL string) ([]types.Message, error) {
	res, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: int32(s.maxNumberOfMessages),
		WaitTimeSeconds:     int32(s.waitTimeSeconds),
		VisibilityTimeout:   int32(s.visibilityTimeoutSeconds),
	})
	if err != nil {
		return nil, err
	}

	return res.Messages, nil
}

func (s *Subscriber) handleMessage(ctx context.Context, queueURL string, m types.Message, h Handler) error {
	Logf("received %s from %s", *m.MessageId, queueURL)

	em := gjson.Get(*m.Body, "Message").Str
	b, err := base64.StdEncoding.DecodeString(em)
	if err != nil {
		return err
	}

	dm, err := Unmarshal(b, h.Message())
	if err != nil {
		return err
	}

	err = h.Handle(ctx, dm.Payload, dm.Metadata)
	if err != nil {
		return err
	}

	_, err = s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: m.ReceiptHandle,
	})
	return err
}

// WithQueueRegistry configures the subscriber to use the specified registry
// to resolve queues, creating them if they do not exist
func WithQueueRegistry(r *Registry) func(*SubscriberOptions) {
	return func(o *SubscriberOptions) {
		o.QueueURLFn = r.QueueURL
	}
}

// WithErrorHandler configures the subscriber to use the specified error handler func
func WithErrorHandler(fn func(error)) func(*SubscriberOptions) {
	return func(o *SubscriberOptions) {
		o.ErrorFn = fn
	}
}
