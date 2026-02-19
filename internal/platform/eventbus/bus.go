package eventbus

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

// EventBus wraps Watermill's in-process GoChannel pub/sub.
type EventBus struct {
	publisher  message.Publisher
	subscriber message.Subscriber
	router     *message.Router
}

// New creates a new in-process event bus.
func New() (*EventBus, error) {
	logger := watermill.NewSlogLogger(slog.Default())

	ch := gochannel.NewGoChannel(gochannel.Config{
		OutputChannelBuffer: 256,
	}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	return &EventBus{
		publisher:  ch,
		subscriber: ch,
		router:     router,
	}, nil
}

// Publisher returns the Watermill publisher.
func (b *EventBus) Publisher() message.Publisher { return b.publisher }

// Subscriber returns the Watermill subscriber.
func (b *EventBus) Subscriber() message.Subscriber { return b.subscriber }

// Router returns the Watermill message router for adding handlers.
func (b *EventBus) Router() *message.Router { return b.router }
