package pram_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/golang/mock/gomock"
	"google.golang.org/protobuf/proto"

	"github.com/stevecallear/pram"
	"github.com/stevecallear/pram/internal/assert"
	"github.com/stevecallear/pram/mocks"
	"github.com/stevecallear/pram/proto/testpb"
)

func TestPublisher_Publish(t *testing.T) {
	tests := []struct {
		name  string
		optFn func(*pram.PublisherOptions)
		setup func(*mocks.MockSNSMockRecorder)
		input proto.Message
		err   bool
	}{
		{
			name:  "should return an error if the topic cannot be resolved",
			setup: func(m *mocks.MockSNSMockRecorder) {},
			input: new(testpb.Message),
			err:   true,
		},
		{
			name: "should return publish errors",
			optFn: func(o *pram.PublisherOptions) {
				o.TopicARNFn = func(context.Context, proto.Message) (string, error) {
					return "topic", nil
				}
			},
			setup: func(m *mocks.MockSNSMockRecorder) {
				m.Publish(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			},
			input: new(testpb.Message),
			err:   true,
		},
		{
			name: "should publish the message",
			optFn: func(o *pram.PublisherOptions) {
				o.TopicARNFn = func(context.Context, proto.Message) (string, error) {
					return "topic", nil
				}
			},
			setup: func(m *mocks.MockSNSMockRecorder) {
				m.Publish(gomock.Any(), gomock.Any()).Return(&sns.PublishOutput{
					MessageId: aws.String("messageid"),
				}, nil).Times(1)
			},
			input: new(testpb.Message),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			snsc := mocks.NewMockSNS(ctrl)
			tt.setup(snsc.EXPECT())

			if tt.optFn == nil {
				tt.optFn = func(*pram.PublisherOptions) {}
			}

			sut := pram.NewPublisher(snsc, tt.optFn)

			err := sut.Publish(context.Background(), tt.input)
			assert.ErrorExists(t, err, tt.err)
		})
	}
}

func TestWithTopicRegistry(t *testing.T) {
	t.Run("should update the options", func(t *testing.T) {
		r := pram.NewRegistry(nil, nil)
		o := pram.PublisherOptions{}

		pram.WithTopicRegistry(r)(&o)

		exp := reflect.ValueOf(r.TopicARN).Pointer()
		act := reflect.ValueOf(o.TopicARNFn).Pointer()

		if act != exp {
			t.Errorf("got %v, expected %v", act, exp)
		}
	})
}
