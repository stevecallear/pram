package pram_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/golang/mock/gomock"
	"google.golang.org/protobuf/proto"

	"github.com/stevecallear/pram"
	"github.com/stevecallear/pram/internal/assert"
	"github.com/stevecallear/pram/internal/store"
	"github.com/stevecallear/pram/mocks"
	"github.com/stevecallear/pram/proto/testpb"
)

var (
	messageName = pram.MessageName(new(testpb.Message))
	topicARN    = "arn:aws:sns:eu-west-1:111122223333:" + messageName
	queueURL    = "https://sqs.eu-west-1.amazonaws.com/111122223333/" + messageName
	queueARN    = "arn:aws:sqs:eu-west-1:111122223333:" + messageName
)

func TestRegistry_TopicARN(t *testing.T) {
	tests := []struct {
		name  string
		setup func(pram.Store, *mocks.MockSNSMockRecorder)
		input proto.Message
		exp   string
		err   bool
	}{
		{
			name: "should return an error if the topic cannot be ensured",
			setup: func(_ pram.Store, c *mocks.MockSNSMockRecorder) {
				c.CreateTopic(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			},
			input: new(testpb.Message),
			err:   true,
		},
		{
			name: "should ensure the topic",
			setup: func(_ pram.Store, c *mocks.MockSNSMockRecorder) {
				c.CreateTopic(gomock.Any(), gomock.Any()).Return(newCreateTopicOutput(), nil).Times(1)
				c.SetTopicAttributes(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			input: new(testpb.Message),
			exp:   topicARN,
		},
		{
			name: "should not ensure the topic more than once",
			setup: func(s pram.Store, _ *mocks.MockSNSMockRecorder) {
				s.GetOrSetTopicARN(context.Background(), messageName, func() (string, error) {
					return topicARN, nil
				})
			},
			input: new(testpb.Message),
			exp:   topicARN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			snsc := mocks.NewMockSNS(ctrl)
			store := new(store.InMemoryStore)

			sut := pram.NewRegistry(snsc, nil, pram.WithStore(store))

			tt.setup(store, snsc.EXPECT())

			act, err := sut.TopicARN(context.Background(), tt.input)
			assert.ErrorExists(t, err, tt.err)

			if act != tt.exp {
				t.Errorf("got %s, expected %s", act, tt.exp)
			}
		})
	}
}

func TestRegistry_QueueURL(t *testing.T) {
	tests := []struct {
		name  string
		setup func(pram.Store, *mocks.MockSNSMockRecorder, *mocks.MockSQSMockRecorder)
		input proto.Message
		exp   string
		err   bool
	}{
		{
			name: "should return an error if the topic cannot be ensured",
			setup: func(_ pram.Store, nc *mocks.MockSNSMockRecorder, _ *mocks.MockSQSMockRecorder) {
				nc.CreateTopic(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			},
			input: new(testpb.Message),
			err:   true,
		},
		{
			name: "should return an error if the subscription cannot be ensured",
			setup: func(_ pram.Store, nc *mocks.MockSNSMockRecorder, qc *mocks.MockSQSMockRecorder) {
				gomock.InOrder(
					nc.CreateTopic(gomock.Any(), gomock.Any()).Return(newCreateTopicOutput(), nil).Times(1),
					nc.SetTopicAttributes(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1),

					qc.CreateQueue(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1),
				)
			},
			input: new(testpb.Message),
			err:   true,
		},
		{
			name: "should return the queue url",
			setup: func(_ pram.Store, nc *mocks.MockSNSMockRecorder, qc *mocks.MockSQSMockRecorder) {
				gomock.InOrder(
					nc.CreateTopic(gomock.Any(), gomock.Any()).Return(newCreateTopicOutput(), nil).Times(1),
					nc.SetTopicAttributes(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1),

					qc.CreateQueue(gomock.Any(), gomock.Any()).Return(newCreateQueueOutput(true), nil).Times(1),
					qc.GetQueueAttributes(gomock.Any(), gomock.Any()).Return(newGetQueueAttributesOutput(true), nil).Times(1),

					qc.CreateQueue(gomock.Any(), gomock.Any()).Return(newCreateQueueOutput(false), nil).Times(1),
					qc.GetQueueAttributes(gomock.Any(), gomock.Any()).Return(newGetQueueAttributesOutput(false), nil).Times(1),

					qc.SetQueueAttributes(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1),

					nc.Subscribe(gomock.Any(), gomock.Any()).Return(newSubscribeOutput(), nil).Times(1),
				)
			},
			input: new(testpb.Message),
			exp:   queueURL,
		},
		{
			name: "should not ensure the topic or queue more than once",
			setup: func(s pram.Store, _ *mocks.MockSNSMockRecorder, _ *mocks.MockSQSMockRecorder) {
				s.GetOrSetTopicARN(context.Background(), messageName, func() (string, error) {
					return topicARN, nil
				})
				s.GetOrSetQueueURL(context.Background(), messageName, func() (string, error) {
					return queueURL, nil
				})
			},
			input: new(testpb.Message),
			exp:   queueURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			snsc := mocks.NewMockSNS(ctrl)
			sqsc := mocks.NewMockSQS(ctrl)
			store := new(store.InMemoryStore)

			tt.setup(store, snsc.EXPECT(), sqsc.EXPECT())

			sut := pram.NewRegistry(snsc, sqsc, pram.WithStore(store))

			act, err := sut.QueueURL(context.Background(), tt.input)
			assert.ErrorExists(t, err, tt.err)

			if act != tt.exp {
				t.Errorf("got %s, expected %s", act, tt.exp)
			}
		})
	}
}

func TestWithPrefixNaming(t *testing.T) {
	t.Run("should configure the options", func(t *testing.T) {
		o := pram.RegistryOptions{}
		pram.WithPrefixNaming("stage", "service")(&o)

		exp := "stage-pram-test-Message"
		act := o.TopicNameFn(new(testpb.Message))

		if act != exp {
			t.Errorf("got %s, expected %s", act, exp)
		}

		exp = "stage-service-pram-test-Message"
		act = o.QueueNameFn(new(testpb.Message))

		if act != exp {
			t.Errorf("got %s, expected %s", act, exp)
		}

		exp = "stage-service-pram-test-Message_error"
		act = o.ErrorQueueNameFn(new(testpb.Message))

		if act != exp {
			t.Errorf("got %s, expected %s", act, exp)
		}
	})
}

func newCreateTopicOutput() *sns.CreateTopicOutput {
	return &sns.CreateTopicOutput{
		TopicArn: aws.String(topicARN),
	}
}

func newCreateQueueOutput(errorQueue bool) *sqs.CreateQueueOutput {
	url := queueURL
	if errorQueue {
		url = url + "_error"
	}

	return &sqs.CreateQueueOutput{
		QueueUrl: aws.String(url),
	}
}

func newGetQueueAttributesOutput(errorQueue bool) *sqs.GetQueueAttributesOutput {
	arn := queueARN
	if errorQueue {
		arn = arn + "_error"
	}

	return &sqs.GetQueueAttributesOutput{
		Attributes: map[string]string{
			"QueueArn": arn,
		},
	}
}

func newSubscribeOutput() *sns.SubscribeOutput {
	return &sns.SubscribeOutput{
		SubscriptionArn: aws.String("arn"),
	}
}
