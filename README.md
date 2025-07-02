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

## Available Commands

### Build and Run

- `make run-broker` - Build and start the message broker
- `make run-consumer` - Build and start a test consumer
- `make run-publisher` - Build and start a test publisher

### Build Only

- `make build-broker` - Build the broker binary
- `make build-consumer` - Build the consumer binary
- `make build-publisher` - Build the publisher binary

### Development

- `make deps` - Download and verify Go modules
- `make tidy` - Clean up Go modules
- `make fmt` - Format all Go code
- `make clean` - Remove build artifacts
- `make gen-pb` - Generate Go code from protobuf files

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

## Project Layout

```
grpc-pubsub/
├── cmd/pubsub/
│   ├── broker/main.go      # Broker server application
│   ├── consumer/main.go    # Test consumer application
│   └── publisher/main.go   # Test publisher application
├── internal/
│   ├── broker/             # Broker implementation
│   │   ├── broker.go       # Main broker logic
│   │   └── status.go       # Status enum helpers
│   ├── consumer/           # Consumer client
│   │   └── consumer.go     # Consumer implementation
│   ├── publisher/          # Publisher client
│   │   └── publisher.go    # Publisher implementation
│   └── server/             # gRPC server wrapper
│       └── server.go       # Server setup and lifecycle
├── pkg/
│   └── protobuf/
│       └── pubsub.proto    # gRPC service definition
├── Makefile                # Build commands
```

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
