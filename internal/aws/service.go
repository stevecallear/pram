package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
)

type (
	// SNS represents an sns client interface
	SNS interface {
		CreateTopic(ctx context.Context, params *sns.CreateTopicInput, optFns ...func(*sns.Options)) (*sns.CreateTopicOutput, error)
		SetTopicAttributes(ctx context.Context, params *sns.SetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.SetTopicAttributesOutput, error)
		Subscribe(ctx context.Context, params *sns.SubscribeInput, optFns ...func(*sns.Options)) (*sns.SubscribeOutput, error)
	}

	// SQS represents an sqs client interface
	SQS interface {
		CreateQueue(ctx context.Context, params *sqs.CreateQueueInput, optFns ...func(*sqs.Options)) (*sqs.CreateQueueOutput, error)
		GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
		SetQueueAttributes(ctx context.Context, params *sqs.SetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.SetQueueAttributesOutput, error)
	}

	// Service represents an sqs/sns queue service
	Service struct {
		snsc  SNS
		sqsc  SQS
		logFn func(string, ...interface{})
	}

	// EnsureTopicRequest represents an ensure topic request
	EnsureTopicRequest struct {
		TopicName string
	}

	// EnsureTopicResponse represents an ensure topic response
	EnsureTopicResponse struct {
		TopicARN string
	}

	// EnsureSubscriptionRequest represents an ensure subscription request
	EnsureSubscriptionRequest struct {
		TopicARN        string
		QueueName       string
		ErrorQueueName  string
		MaxReceiveCount int
	}

	// EnsureSubscriptionResponse represents an ensure subscription response
	EnsureSubscriptionResponse struct {
		QueueURL string
	}
)

// NewService returns a new queue service
func NewService(snsc SNS, sqsc SQS, logFn func(string, ...interface{})) *Service {
	return &Service{
		snsc:  snsc,
		sqsc:  sqsc,
		logFn: logFn,
	}
}

// EnsureTopic ensures that the specified topic exists
func (s *Service) EnsureTopic(ctx context.Context, req EnsureTopicRequest) (EnsureTopicResponse, error) {
	res, err := s.snsc.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: awssdk.String(req.TopicName),
	})
	if err != nil {
		return EnsureTopicResponse{}, err
	}

	ap, err := SNSAccessPolicy(*res.TopicArn)
	if err != nil {
		return EnsureTopicResponse{}, err
	}

	_, err = s.snsc.SetTopicAttributes(ctx, &sns.SetTopicAttributesInput{
		TopicArn:       res.TopicArn,
		AttributeName:  awssdk.String("Policy"),
		AttributeValue: awssdk.String(ap),
	})
	if err != nil {
		return EnsureTopicResponse{}, err
	}

	s.log("created topic %s", *res.TopicArn)

	return EnsureTopicResponse{
		TopicARN: *res.TopicArn,
	}, nil
}

// EnsureSubscription ensures that the specified topic subscription, queue and error queue exist
func (s *Service) EnsureSubscription(ctx context.Context, req EnsureSubscriptionRequest) (EnsureSubscriptionResponse, error) {
	_, eqa, err := s.createQueue(ctx, req.ErrorQueueName)
	if err != nil {
		return EnsureSubscriptionResponse{}, err
	}

	mqu, mqa, err := s.createQueue(ctx, req.QueueName)
	if err != nil {
		return EnsureSubscriptionResponse{}, err
	}

	ap, err := SQSAccessPolicy(req.TopicARN, mqa)
	if err != nil {
		return EnsureSubscriptionResponse{}, err
	}

	rp, err := SQSRedrivePolicy(eqa, req.MaxReceiveCount)
	if err != nil {
		return EnsureSubscriptionResponse{}, err
	}

	_, err = s.sqsc.SetQueueAttributes(ctx, &sqs.SetQueueAttributesInput{
		QueueUrl: awssdk.String(mqu),
		Attributes: map[string]string{
			"Policy":        ap,
			"RedrivePolicy": rp,
		},
	})
	if err != nil {
		return EnsureSubscriptionResponse{}, err
	}

	sr, err := s.snsc.Subscribe(ctx, &sns.SubscribeInput{
		Protocol: awssdk.String("sqs"),
		TopicArn: awssdk.String(req.TopicARN),
		Endpoint: awssdk.String(mqa),
	})
	if err != nil {
		return EnsureSubscriptionResponse{}, err
	}

	s.log("created subscription %s", *sr.SubscriptionArn)

	return EnsureSubscriptionResponse{
		QueueURL: mqu,
	}, nil
}

func (s *Service) createQueue(ctx context.Context, queueName string) (string, string, error) {

	cqr, err := s.sqsc.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: awssdk.String(queueName),
	})
	if err != nil {
		return "", "", err
	}

	qar, err := s.sqsc.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       cqr.QueueUrl,
		AttributeNames: []types.QueueAttributeName{"QueueArn"},
	})
	if err != nil {
		return "", "", err
	}

	s.log("created queue %s", *cqr.QueueUrl)

	return *cqr.QueueUrl, qar.Attributes["QueueArn"], nil
}

func (s *Service) log(format string, a ...interface{}) {
	if s.logFn != nil {
		s.logFn(format, a...)
	}
}
