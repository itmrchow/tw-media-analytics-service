package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/iterator"
)

var _ Queue = &GcpPubSub{}

type GcpPubSub struct {
	tracer trace.Tracer
	logger *zerolog.Logger
	client *pubsub.Client
}

func NewGcpPubSub(ctx context.Context, logger *zerolog.Logger) *GcpPubSub {
	// Tracer
	tracer := otel.Tracer("domain/queue")
	ctx, span := tracer.Start(ctx, "domain/queue/NewGcpPubSub: New Gcp PubSub")

	// Logger
	logger.Info().Ctx(ctx).Msg("NewGcpPubSub: start")
	defer func() {
		span.End()
		logger.Info().Ctx(ctx).Msg("NewGcpPubSub: end")
	}()

	// ProjectID
	projectID := viper.GetString("GCP_PROJECT_ID")

	// Create client
	client, err := pubsub.NewClientWithConfig(ctx, projectID, &pubsub.ClientConfig{
		EnableOpenTelemetryTracing: true,
	})
	if err != nil {
		logger.Fatal().Ctx(ctx).Err(err).Msg("Failed to create client")
	}

	// Create object
	g := &GcpPubSub{
		tracer: tracer,
		logger: logger,
		client: client,
	}

	return g
}

func (g *GcpPubSub) InitTopic(ctx context.Context) error {
	// Trace
	ctx, span := g.tracer.Start(ctx, "domain/queue/InitTopic: Init Topic")
	defer func() {
		span.End()
		g.logger.Info().Ctx(ctx).Msg("InitTopic: end")
	}()

	g.logger.Info().Ctx(ctx).Msg("InitTopic: start")

	// Get topics
	expectedTopics := GetTopics()

	// Get exist topics
	existingTopics := make(map[string]struct{})
	it := g.client.Topics(ctx)
	for {
		topic, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list topics: %w", err)
		}
		existingTopics[topic.ID()] = struct{}{}
	}

	for _, expectedTopic := range expectedTopics {
		topicName := fmt.Sprintf("%s_%s", string(expectedTopic), viper.GetString("ENV"))

		// Check if topic exists
		if _, exists := existingTopics[topicName]; exists {
			g.logger.Debug().Ctx(ctx).Str("topic", topicName).Msg("Topic already exists")
			continue
		}

		// Create Topic
		topic, err := g.client.CreateTopic(ctx, topicName)
		if err != nil {
			g.logger.Error().Ctx(ctx).Err(err).Msg("Failed to create topic")
			return err
		}
		g.logger.Info().Msgf("Topic %s created", topic.ID())
	}

	return nil
}

func (g *GcpPubSub) CloseClient() error {
	return g.client.Close()
}

func (g *GcpPubSub) Publish(ctx context.Context, topicID QueueTopic, message any) error {
	// Trace
	ctx, span := g.tracer.Start(ctx, "domain/queue/gcp_pub_sub/Publish: Publish Message")
	defer func() {
		g.logger.Info().Ctx(ctx).Msg("Publish: end")
		span.End()
	}()

	g.logger.Info().Ctx(ctx).Msg("Publish: start")

	data, err := json.Marshal(message)
	if err != nil {
		g.logger.Error().Err(err).Msg("Failed to marshal message")
		return err
	}

	topic := g.client.Topic(string(topicID) + "_" + viper.GetString("ENV"))
	_, err = topic.Publish(ctx, &pubsub.Message{
		Data: data,
	}).Get(ctx)
	if err != nil {
		g.logger.Error().Err(err).Msg("Failed to publish message")
		return err
	}
	return nil
}

func (g *GcpPubSub) Consume(
	ctx context.Context,
	topic QueueTopic,
	subID string,
	handler func(ctx context.Context, msg []byte) error,
) error {
	// Trace
	ctx, span := g.tracer.Start(ctx, "domain/queue/Consume: Consume Setting Subscription")
	defer func() {
		g.logger.Info().Ctx(ctx).Msg("Consume: end")
		span.End()
	}()

	g.logger.Info().Ctx(ctx).Msg("Consume: start")

	if subID == "" {
		subID = fmt.Sprintf("%s_%s_sub", string(topic), viper.GetString("ENV"))
	}

	sub := g.client.Subscription(subID)
	exists, err := sub.Exists(ctx)
	if err != nil {
		g.logger.Error().Err(err).Msg("Failed to check if subscription exists")
		return err
	}

	if !exists {
		topic := fmt.Sprintf("%s_%s", string(topic), viper.GetString("ENV"))
		sub, err = g.client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
			Topic:       g.client.Topic(string(topic)),
			AckDeadline: 20 * time.Second,
		})
		if err != nil {
			g.logger.Error().Err(err).Msg("Failed to create subscription")
			return err
		}
	}

	g.logger.Info().Msgf("Consuming messages from %s", sub.ID())

	sub.ReceiveSettings = pubsub.ReceiveSettings{
		NumGoroutines:          3,
		MaxOutstandingMessages: 15,
	}

	// Create goroutine to receive messages
	go func() {
		if err = sub.Receive(ctx, func(msgCtx context.Context, msg *pubsub.Message) {
			if err = handler(msgCtx, msg.Data); err != nil {
				g.logger.Error().Err(err).Msg("Failed to handle message")
				msg.Nack()
				return
			}
			msg.Ack()
		}); err != nil {
			g.logger.Error().Err(err).Msg("Failed to receive message")
			return
		}
	}()

	return nil
}
