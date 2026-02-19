package eventbus

import (
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
)

// Publish marshals the event to JSON and publishes it to the given topic.
func Publish(pub message.Publisher, topic string, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := message.NewMessage(uuid.NewString(), payload)

	if err := pub.Publish(topic, msg); err != nil {
		return fmt.Errorf("publish to %s: %w", topic, err)
	}

	return nil
}
