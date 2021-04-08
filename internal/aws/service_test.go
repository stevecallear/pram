package aws_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/golang/mock/gomock"

	"github.com/stevecallear/pram/internal/assert"
	"github.com/stevecallear/pram/internal/aws"
	"github.com/stevecallear/pram/internal/aws/mocks"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go -package=mocks

const (
	topicName = "stage-package-Message"
	topicARN  = "arn:aws:sns:eu-west-1:111122223333:" + topicName

	queueName = "stage-service-package-Message"
	queueURL  = "https://sqs.eu-west-1.amazonaws.com/111122223333/" + queueName
	queueARN  = "arn:aws:sqs:eu-west-1:111122223333:" + queueName

	errorQueueName = queueName + "_error"
	errorQueueURL  = "https://sqs.eu-west-1.amazonaws.com/111122223333/" + errorQueueName
	errorQueueARN  = "arn:aws:sqs:eu-west-1:111122223333:" + errorQueueName
)

func TestService_EnsureTopic(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*mocks.MockSNSMockRecorder)
		input aws.EnsureTopicRequest
		exp   aws.EnsureTopicResponse
		err   bool
	}{
		{
			name: "should return an error if the topic cannot be created",
			setup: func(m *mocks.MockSNSMockRecorder) {
				m.CreateTopic(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			},
			input: aws.EnsureTopicRequest{
				TopicName: topicName,
			},
			err: true,
		},
		{
			name: "should return an error if the policy cannot be generated",
			setup: func(m *mocks.MockSNSMockRecorder) {
				m.CreateTopic(gomock.Any(), gomock.Any()).Return(&sns.CreateTopicOutput{
					TopicArn: awssdk.String("invalid"),
				}, nil).Times(1)
			},
			input: aws.EnsureTopicRequest{
				TopicName: topicName,
			},
			err: true,
		},
		{
			name: "should return an error if the attributes cannot be set",
			setup: func(m *mocks.MockSNSMockRecorder) {
				m.CreateTopic(gomock.Any(), gomock.Any()).Return(&sns.CreateTopicOutput{
					TopicArn: awssdk.String(topicARN),
				}, nil).Times(1)

				m.SetTopicAttributes(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			},
			input: aws.EnsureTopicRequest{
				TopicName: topicName,
			},
			err: true,
		},
		{
			name: "should ensure the topic exists",
			setup: func(m *mocks.MockSNSMockRecorder) {
				m.CreateTopic(gomock.Any(), &sns.CreateTopicInput{
					Name: awssdk.String(topicName),
				}).Return(&sns.CreateTopicOutput{
					TopicArn: awssdk.String(topicARN),
				}, nil).Times(1)

				m.SetTopicAttributes(gomock.Any(), gomock.Any()).
					Return(new(sns.SetTopicAttributesOutput), nil).Times(1)
			},
			input: aws.EnsureTopicRequest{
				TopicName: topicName,
			},
			exp: aws.EnsureTopicResponse{
				TopicARN: topicARN,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			snsc := mocks.NewMockSNS(ctrl)
			tt.setup(snsc.EXPECT())

			sut := aws.NewService(snsc, nil, nil)
			act, err := sut.EnsureTopic(context.Background(), tt.input)

			assert.ErrorExists(t, err, tt.err)
			assert.DeepEqual(t, act, tt.exp)
		})
	}
}

func TestService_EnsureSubscription(t *testing.T) {
	input := aws.EnsureSubscriptionRequest{
		TopicARN:        topicARN,
		QueueName:       queueName,
		ErrorQueueName:  errorQueueName,
		MaxReceiveCount: 5,
	}

	tests := []struct {
		name  string
		setup func(*mocks.MockSNSMockRecorder, *mocks.MockSQSMockRecorder)
		input aws.EnsureSubscriptionRequest
		exp   aws.EnsureSubscriptionResponse
		err   bool
	}{
		{
			name: "should return an error if the error queue cannot be created",
			setup: func(snsc *mocks.MockSNSMockRecorder, sqsc *mocks.MockSQSMockRecorder) {
				sqsc.CreateQueue(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			},
			input: input,
			err:   true,
		},
		{
			name: "should return an error if the error queue attributes cannot be retrieved",
			setup: func(snsc *mocks.MockSNSMockRecorder, sqsc *mocks.MockSQSMockRecorder) {
				gomock.InOrder(
					sqsc.CreateQueue(gomock.Any(), gomock.Any()).Return(&sqs.CreateQueueOutput{
						QueueUrl: awssdk.String(errorQueueURL),
					}, nil).Times(1),

					sqsc.GetQueueAttributes(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1),
				)
			},
			input: input,
			err:   true,
		},
		{
			name: "should return an error if the queue cannot be created",
			setup: func(snsc *mocks.MockSNSMockRecorder, sqsc *mocks.MockSQSMockRecorder) {
				gomock.InOrder(
					sqsc.CreateQueue(gomock.Any(), gomock.Any()).Return(&sqs.CreateQueueOutput{
						QueueUrl: awssdk.String(errorQueueURL),
					}, nil).Times(1),

					sqsc.GetQueueAttributes(gomock.Any(), gomock.Any()).Return(&sqs.GetQueueAttributesOutput{
						Attributes: map[string]string{
							"QueueArn": errorQueueARN,
						},
					}, nil).Times(1),

					sqsc.CreateQueue(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1),
				)
			},
			input: input,
			err:   true,
		},
		{
			name: "should return an error if the attribute cannot be set",
			setup: func(snsc *mocks.MockSNSMockRecorder, sqsc *mocks.MockSQSMockRecorder) {
				gomock.InOrder(
					sqsc.CreateQueue(gomock.Any(), gomock.Any()).Return(&sqs.CreateQueueOutput{
						QueueUrl: awssdk.String(errorQueueURL),
					}, nil).Times(1),

					sqsc.GetQueueAttributes(gomock.Any(), gomock.Any()).Return(&sqs.GetQueueAttributesOutput{
						Attributes: map[string]string{
							"QueueArn": errorQueueARN,
						},
					}, nil).Times(1),

					sqsc.CreateQueue(gomock.Any(), gomock.Any()).Return(&sqs.CreateQueueOutput{
						QueueUrl: awssdk.String(queueURL),
					}, nil).Times(1),

					sqsc.GetQueueAttributes(gomock.Any(), gomock.Any()).Return(&sqs.GetQueueAttributesOutput{
						Attributes: map[string]string{
							"QueueArn": queueARN,
						},
					}, nil).Times(1),

					sqsc.SetQueueAttributes(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1),
				)
			},
			input: input,
			err:   true,
		},
		{
			name: "should return an error if the subscription cannot be created",
			setup: func(snsc *mocks.MockSNSMockRecorder, sqsc *mocks.MockSQSMockRecorder) {
				gomock.InOrder(
					sqsc.CreateQueue(gomock.Any(), gomock.Any()).Return(&sqs.CreateQueueOutput{
						QueueUrl: awssdk.String(errorQueueURL),
					}, nil).Times(1),

					sqsc.GetQueueAttributes(gomock.Any(), gomock.Any()).Return(&sqs.GetQueueAttributesOutput{
						Attributes: map[string]string{
							"QueueArn": errorQueueARN,
						},
					}, nil).Times(1),

					sqsc.CreateQueue(gomock.Any(), gomock.Any()).Return(&sqs.CreateQueueOutput{
						QueueUrl: awssdk.String(queueURL),
					}, nil).Times(1),

					sqsc.GetQueueAttributes(gomock.Any(), gomock.Any()).Return(&sqs.GetQueueAttributesOutput{
						Attributes: map[string]string{
							"QueueArn": queueARN,
						},
					}, nil).Times(1),

					sqsc.SetQueueAttributes(gomock.Any(), gomock.Any()).
						Return(new(sqs.SetQueueAttributesOutput), nil).Times(1),

					snsc.Subscribe(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1),
				)
			},
			input: input,
			err:   true,
		},
		{
			name: "should ensure the subscription exists",
			setup: func(snsc *mocks.MockSNSMockRecorder, sqsc *mocks.MockSQSMockRecorder) {
				gomock.InOrder(
					sqsc.CreateQueue(gomock.Any(), &sqs.CreateQueueInput{
						QueueName: awssdk.String(errorQueueName),
					}).Return(&sqs.CreateQueueOutput{
						QueueUrl: awssdk.String(errorQueueURL),
					}, nil).Times(1),

					sqsc.GetQueueAttributes(gomock.Any(), gomock.Any()).Return(&sqs.GetQueueAttributesOutput{
						Attributes: map[string]string{
							"QueueArn": errorQueueARN,
						},
					}, nil).Times(1),

					sqsc.CreateQueue(gomock.Any(), &sqs.CreateQueueInput{
						QueueName: awssdk.String(queueName),
					}).Return(&sqs.CreateQueueOutput{
						QueueUrl: awssdk.String(queueURL),
					}, nil).Times(1),

					sqsc.GetQueueAttributes(gomock.Any(), gomock.Any()).Return(&sqs.GetQueueAttributesOutput{
						Attributes: map[string]string{
							"QueueArn": queueARN,
						},
					}, nil).Times(1),

					sqsc.SetQueueAttributes(gomock.Any(), gomock.Any()).
						Return(new(sqs.SetQueueAttributesOutput), nil).Times(1),

					snsc.Subscribe(gomock.Any(), &sns.SubscribeInput{
						Protocol: awssdk.String("sqs"),
						TopicArn: awssdk.String(topicARN),
						Endpoint: awssdk.String(queueARN),
					}).Return(&sns.SubscribeOutput{
						SubscriptionArn: awssdk.String("arn"),
					}, nil).Times(1),
				)
			},
			input: input,
			exp: aws.EnsureSubscriptionResponse{
				QueueURL: queueURL,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			snsc := mocks.NewMockSNS(ctrl)
			sqsc := mocks.NewMockSQS(ctrl)
			tt.setup(snsc.EXPECT(), sqsc.EXPECT())

			sut := aws.NewService(snsc, sqsc, nil)
			act, err := sut.EnsureSubscription(context.Background(), tt.input)

			assert.ErrorExists(t, err, tt.err)
			assert.DeepEqual(t, act, tt.exp)
		})
	}
}
