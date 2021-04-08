package pram_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/golang/mock/gomock"
	"google.golang.org/protobuf/proto"

	"github.com/stevecallear/pram"
	"github.com/stevecallear/pram/internal/assert"
	"github.com/stevecallear/pram/mocks"
	"github.com/stevecallear/pram/proto/testpb"
)

func TestSubscriber_SubscribeAsync(t *testing.T) {
	msg := &testpb.Message{Value: "value"}

	tests := []struct {
		name     string
		setup    func(*mocks.MockSQSMockRecorder)
		queueFn  func(context.Context, proto.Message) (string, error)
		handleFn func(context.Context, proto.Message, pram.Metadata) error
		err      bool
	}{
		{
			name:  "should return an error if the queue cannot be resolved",
			setup: func(*mocks.MockSQSMockRecorder) {},
			err:   true,
		},
		{
			name: "should return receive errors",
			setup: func(m *mocks.MockSQSMockRecorder) {
				m.ReceiveMessage(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			},
			queueFn: func(context.Context, proto.Message) (string, error) {
				return "queue", nil
			},
			err: true,
		},
		{
			name: "should send decode errors",
			setup: func(m *mocks.MockSQSMockRecorder) {
				m.ReceiveMessage(gomock.Any(), gomock.Any()).Return(&sqs.ReceiveMessageOutput{
					Messages: []types.Message{
						{
							MessageId:     aws.String("messageid"),
							Body:          aws.String("{\"Message\":\"\"}"),
							ReceiptHandle: aws.String("receipthandle"),
						},
					},
				}, nil).Times(1)
			},
			queueFn: func(context.Context, proto.Message) (string, error) {
				return "queue", nil
			},
			err: true,
		},
		{
			name: "should send handle errors",
			setup: func(m *mocks.MockSQSMockRecorder) {
				m.ReceiveMessage(gomock.Any(), gomock.Any()).Return(newReceiveMessageOutput(msg), nil).Times(1)
			},
			queueFn: func(context.Context, proto.Message) (string, error) {
				return "queue", nil
			},
			handleFn: func(context.Context, proto.Message, pram.Metadata) error {
				return errors.New("error")
			},
			err: true,
		},
		{
			name: "should send delete errors",
			setup: func(m *mocks.MockSQSMockRecorder) {
				m.ReceiveMessage(gomock.Any(), gomock.Any()).Return(newReceiveMessageOutput(msg), nil).Times(1)

				m.DeleteMessage(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			},
			queueFn: func(context.Context, proto.Message) (string, error) {
				return "queue", nil
			},
			handleFn: func(context.Context, proto.Message, pram.Metadata) error {
				return nil
			},
			err: true,
		},
		{
			name: "should handle messages",
			setup: func(m *mocks.MockSQSMockRecorder) {
				m.ReceiveMessage(gomock.Any(), gomock.Any()).Return(newReceiveMessageOutput(msg), nil).Times(1)

				m.DeleteMessage(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			},
			queueFn: func(context.Context, proto.Message) (string, error) {
				return "queue", nil
			},
			handleFn: func(context.Context, proto.Message, pram.Metadata) error {
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqsc := mocks.NewMockSQS(ctrl)
			tt.setup(sqsc.EXPECT())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var err error
			sut := pram.NewSubscriber(sqsc, func(o *pram.SubscriberOptions) {
				if tt.queueFn != nil {
					o.QueueURLFn = tt.queueFn
				}

				o.ErrorFn = func(e error) {
					err = e
					cancel()
				}

				o.ReceiveInterval = 10 * time.Millisecond
				o.WaitTimeSeconds = 0
			})

			serr := sut.Subscribe(ctx, newHandler(tt.handleFn, cancel))
			if err == nil {
				err = serr
			}

			assert.ErrorExists(t, err, tt.err)
		})
	}
}

func TestWithQueueRegistry(t *testing.T) {
	t.Run("should update the options", func(t *testing.T) {
		r := pram.NewRegistry(nil, nil)

		o := pram.SubscriberOptions{}
		pram.WithQueueRegistry(r)(&o)

		exp := reflect.ValueOf(r.QueueURL).Pointer()
		act := reflect.ValueOf(o.QueueURLFn).Pointer()

		if act != exp {
			t.Errorf("got %v, expected %v", act, exp)
		}
	})
}

func TestWithErrorHandler(t *testing.T) {
	t.Run("should update the options", func(t *testing.T) {
		fn := func(error) {}

		o := pram.SubscriberOptions{}
		pram.WithErrorHandler(fn)(&o)

		exp := reflect.ValueOf(fn).Pointer()
		act := reflect.ValueOf(o.ErrorFn).Pointer()

		if act != exp {
			t.Errorf("got %v, expected %v", act, exp)
		}
	})
}

type handler struct {
	handleFn func(context.Context, proto.Message, pram.Metadata) error
	cancel   context.CancelFunc
}

func newHandler(handleFn func(context.Context, proto.Message, pram.Metadata) error, cancel context.CancelFunc) *handler {
	return &handler{
		handleFn: handleFn,
		cancel:   cancel,
	}
}

func (h *handler) Message() proto.Message {
	return new(testpb.Message)
}

func (h *handler) Handle(ctx context.Context, m proto.Message, md pram.Metadata) error {
	defer h.cancel()
	return h.handleFn(ctx, m, md)
}

func newReceiveMessageOutput(m proto.Message) *sqs.ReceiveMessageOutput {
	enc, err := pram.Marshal(m)
	if err != nil {
		panic(err)
	}

	bm := map[string]string{
		"Message": base64.StdEncoding.EncodeToString(enc),
	}

	bb, err := json.Marshal(bm)
	if err != nil {
		panic(err)
	}

	return &sqs.ReceiveMessageOutput{
		Messages: []types.Message{
			{
				MessageId:     aws.String("messageid"),
				Body:          aws.String(string(bb)),
				ReceiptHandle: aws.String("receipthandle"),
			},
		},
	}
}
