package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/stevecallear/pram"
	"github.com/stevecallear/pram/proto/testpb"
)

func main() {
	pram.SetLogger(log.Default())

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	snsClient := newLocalStackSNS(cfg)
	sqsClient := newLocalStackSQS(cfg)

	reg := pram.NewRegistry(snsClient, sqsClient, pram.WithPrefixNaming("local", "example"))
	pub := pram.NewPublisher(snsClient, pram.WithTopicRegistry(reg))
	sub := pram.NewSubscriber(sqsClient, pram.WithQueueRegistry(reg), pram.WithErrorHandler(func(err error) {
		pram.Logf("subscriber: %v", err)
	}))

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-c
		log.Println("shutting down")
		cancel()
	}()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		defer wg.Done()

		err := sub.Subscribe(ctx, new(handler))
		if err != nil {
			pram.Logf("subscribe: %v", err)
		}
	}()

	go func() {
		defer wg.Done()

		t := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				msg := &testpb.Message{Value: uuid.NewString()}
				err := pub.Publish(ctx, msg)
				if err != nil {
					pram.Logf("publish: %v", err)
				}
			}
		}
	}()

	wg.Wait()
	log.Println("done")
}

func newLocalStackSNS(cfg aws.Config) *sns.Client {
	return sns.NewFromConfig(cfg, func(opts *sns.Options) {
		opts.EndpointResolver = sns.EndpointResolverFromURL("http://localhost:4566")
	})
}

func newLocalStackSQS(cfg aws.Config) *sqs.Client {
	return sqs.NewFromConfig(cfg, func(opts *sqs.Options) {
		opts.EndpointResolver = sqs.EndpointResolverFromURL("http://localhost:4566")
	})
}

type handler struct{}

func (h *handler) Message() proto.Message {
	return new(testpb.Message)
}

func (h *handler) Handle(ctx context.Context, m proto.Message, md pram.Metadata) error {
	tm := m.(*testpb.Message)
	log.Println("handled:", tm.Value)
	return nil
}
