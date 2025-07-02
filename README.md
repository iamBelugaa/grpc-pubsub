# gRPC PubSub System

A simple publish-subscribe messaging system built with gRPC and Go. Send
messages to topics and have multiple clients receive them in real-time.

## Getting Started

### 1. Install dependencies

```bash
make deps
```

### 2. Start the broker server

The broker must run first since it handles all message routing:

```bash
make run-broker
```

This starts the broker on `localhost:8080`.

### 3. Start a consumer

Open a new terminal and run:

```bash
make run-consumer
```

The consumer will connect to the broker and subscribe to the "kubernetes" topic.

### 4. Publish some messages

Open another terminal and run:

```bash
make run-publisher
```

The publisher will send several "New release: Kubernetes 1.30 is out" messages
to the "kubernetes" topic. You'll see the consumer receive these messages in
real-time.

## How it works

The system uses three gRPC methods defined in `pubsub.proto`:

1. **Publish**: Sends a message to a topic
2. **Subscribe**: Opens a stream to receive messages from a topic
3. **Unsubscribe**: Stops receiving messages from a topic

The broker keeps track of which consumers are subscribed to which topics. When a
publisher sends a message, the broker looks up all subscribers for that topic
and forwards the message to each one.

Each consumer gets a unique ID (UUID) so the broker can track subscriptions
properly.

## Protocol Buffer Definition

The gRPC service is defined in `pkg/protobuf/pubsub.proto` with three main
operations:

- `Publish(PublishRequest) returns (PublishResponse)` - Send a message
- `Subscribe(SubscribeRequest) returns (stream PayloadStream)` - Receive
  messages
- `Unsubscribe(UnsubscribeRequest) returns (UnsubscribeResponse)` - Stop
  receiving

Messages contain a topic name and payload bytes, so you can send any kind of
data.
