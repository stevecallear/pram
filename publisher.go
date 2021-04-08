package pram

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"google.golang.org/protobuf/proto"
)

type (
	// Publisher represents a publisher
	Publisher struct {
		client     SNS
		topicARNFn func(context.Context, proto.Message) (string, error)
	}

	// PublisherOptions represents a set of publisher options
	PublisherOptions struct {
		TopicARNFn func(context.Context, proto.Message) (string, error)
	}
)

// NewPublisher returns a new publisher
func NewPublisher(client SNS, optFns ...func(*PublisherOptions)) *Publisher {
	o := PublisherOptions{
		TopicARNFn: func(context.Context, proto.Message) (string, error) {
			return "", errors.New("topic not found")
		},
	}

	for _, fn := range optFns {
		fn(&o)
	}

	return &Publisher{
		client:     client,
		topicARNFn: o.TopicARNFn,
	}
}

// Publish publishes the specified message
func (p *Publisher) Publish(ctx context.Context, m proto.Message, opts ...func(*Metadata)) error {
	b, err := Marshal(m, opts...)
	if err != nil {
		return err
	}

	arn, err := p.topicARNFn(ctx, m)
	if err != nil {
		return err
	}

	res, err := p.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(arn),
		Message:  aws.String(base64.StdEncoding.EncodeToString(b)),
	})
	if err != nil {
		return err
	}

	Logf("published %s to %s", *res.MessageId, arn)
	return nil
}

// WithTopicRegistry configures the subscriber to use the specified registry
// to resolve topics, creating them if they do not exist
func WithTopicRegistry(r *Registry) func(*PublisherOptions) {
	return func(o *PublisherOptions) {
		o.TopicARNFn = r.TopicARN
	}
}
