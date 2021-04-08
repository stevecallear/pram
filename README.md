# pram
[![Build Status](https://travis-ci.org/stevecallear/pram.svg?branch=master)](https://travis-ci.org/stevecallear/pram)
[![codecov](https://codecov.io/gh/stevecallear/pram/branch/master/graph/badge.svg)](https://codecov.io/gh/stevecallear/pram)
[![Go Report Card](https://goreportcard.com/badge/github.com/stevecallear/pram)](https://goreportcard.com/report/github.com/stevecallear/pram)

`pram` is a lightweight messaging framework using AWS SNS/SQS and Google Protobuf with convention based infrastructure creation.

## Publisher
`Publisher` publishes messages to the appropriate topic. The topic ARN is resolved using the `PublisherOptions.TopicARNFn` function. A `Registry` instance can be used to resolve/create infrastucture by convention.

Published messages are wrapped with the `prampb.Message` type, then encoded to a base64 representation of the marshalled byte slice prior to being sent to SNS.

```
r := pram.NewRegistry(snsClient, sqsClient, pram.WithPrefixNaming("dev", "service"))
p := pram.NewPublisher(snsClient, pram.WithTopicRegistry(r))

err := p.Publish(context.Background(), &testpb.Message{Value: "value"})
if err != nil {
    log.Fatalln(err)
}
```

### Metadata
Message metadata can be modified at the point of publish, for example to add a correlation ID.

```
err := p.Publish(context.Background(), m, pram.WithCorrelationID(correlationID))
```

## Subscriber
`Subscriber` receives messages published to the appropriate queue. The queue URL is resolved using the `SubscriberOptions.QueueURLFn` function. A `Registry` instance can be used to resolve/create infrastructure by convention.

### Handler
Each message subscription requires an implementation of `pram.Handler` to generate empty messages of the appropriate type and handle received messages. A one-to-one mapping between message types and handlers is assumed, with the message instance from `Message` guaranteed to be the input to `Handle`.

```
type handler struct {}

func (h *handler) Message() proto.Message {
    return new(testpb.Message)
}

func (h *handler) Handle(ctx context.Context, m proto.Message, md pram.Metadata) error {
	tm := m.(*testpb.Message)
	// handle the message
	return nil
}
```

### Subscribe
A message subscription can be created using `Subscribe`. Each received message will spawn a new goroutine to execute the supplied handler.

By default message receive and handling errors are discarded. This behaviour can be changed using `pram.WithErrorHandler`.

```
r := pram.NewRegistry(snsClient, sqsClient, pram.WithPrefixNaming("dev", "service"))
s := pram.NewSubscriber(sqsClient, pram.WithQueueRegistry(r))

// Subscribe will block until the supplied context is cancelled
err := s.Subscribe(context.Background(), new(handler))
```

While each call to `Subscribe` is blocking, a single subscriber can handle multiple message types by using goroutines.

```
r := pram.NewRegistry(snsClient, sqsClient, pram.WithPrefixNaming("dev", "service"))
s := pram.NewSubscriber(sqsClient, pram.WithQueueRegistry(r))

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

wg := new(sync.WaitGroup)
for _, h := range []pram.Handler{new(handlerA), new(handlerB)} {
    wg.Add(1)

    go func(h pram.Handler) {
        defer wg.Done()

        err := s.Subscribe(ctx, h)
        if err != nil {
            log.Println(err)
        }
    }(h)
}

wg.Wait()
```

## Registry
`Registry` is responsible for creating SNS/SQS infrastructure by convention. The adopted naming convention defines how messages will be routed.

By default each published/subscribed message will result in a single SNS topic and associated SQS queue. This effectively results in competing consumers for all message types, which will likely be inappropriate for all but the simplest implementations.

To support typical routing patterns, the registry should be configured to generate queues for each subscribing service. This convention can be applied using `pram.WithPrefixNaming`.

### Example
Service 'a' publishes a message to the `dev-package-Message` SNS topic. All instances of service 'a' will publish to the same topic.

```
r := pram.NewRegistry(snsc, sqsc, pram.WithPrefixNaming("dev", "a"))
p := pram.NewPublisher(snsc, pram.WithTopicRegistry(r))

p.Publish(ctx, new(package.Message))
```

Service 'b' subscribes to the `dev-package-Message` topic using the `dev-b-package-Message` SQS queue. All instances of service 'b' will act as competing consumers.

```
r := pram.NewRegistry(snsc, sqsc, pram.WithPrefixNaming("dev", "b"))
s := pram.NewSubscriber(snsc, sqsc, pram.WithQueueRegistry(r))

s.Subscribte(ctx, new(handler))
```

Service 'c' subscribes to the same topic, but uses the `dev-c-package-Message` queue. All instances of service 'c' will act as competing consumers, but are served by a separate queue from that used by service 'b'.

```
r := pram.NewRegistry(snsc, sqsc, pram.WithPrefixNaming("dev", "c"))
s := pram.NewSubscriber(snsc, sqsc, pram.WithQueueRegistry(r))

s.Subscribte(ctx, new(handler))
```

## Logging
Info level logs, such as infrastructure creation and message publish/receive can be output by providing a `pram.Logger` implementation to `pram.SetLogger`. This can be used to understand the underlying AWS SDK calls being made. For example, the following configuration uses a standard library logger.

```
l := log.New(os.Stdout, "", log.Ldate|log.Ltime)
pram.SetLogger(l)
```

In the case of message publishing, registry and handler errors are returned immediately to the calling code, so can be handled in the usual manner. For message subscriptions, however, the `Subscribe` function will only return an error if the required queue cannot be resolved. Handler and AWS SDK errors will not be returned. To log these errors, an handler should be supplied when creating the subscriber.

```
s := pram.NewSubscriber(sqsClient, pram.WithQueueRegistry(reg), pram.WithErrorHandler(func(err error) {
    pram.Logf("subscriber: %v", err)
}))
```