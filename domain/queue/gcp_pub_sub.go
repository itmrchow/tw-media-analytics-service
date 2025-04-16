package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var _ Queue = &GcpPubSub{}

type GcpPubSub struct {
	log    *zerolog.Logger
	client *pubsub.Client
}

func NewGcpPubSub(ctx context.Context, logger *zerolog.Logger) *GcpPubSub {

	projectID := viper.GetString("GCP_PROJECT_ID")
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create client")
	}

	return &GcpPubSub{
		log:    logger,
		client: client,
	}
}

func (g *GcpPubSub) CreateTopic() error {

	topics := GetTopics()

	for _, topicStr := range topics {
		topicStr := fmt.Sprintf("%s_%s", string(topicStr), viper.GetString("ENV"))

		topic := g.client.Topic(topicStr)
		exists, err := topic.Exists(context.Background())
		if err != nil {
			g.log.Error().Err(err).Msg("Failed to check if topic exists")
			return err
		}

		if !exists {
			topic, err := g.client.CreateTopic(context.Background(), string(topicStr))
			if err != nil {
				g.log.Error().Err(err).Msg("Failed to create topic")
				return err
			}
			g.log.Info().Msgf("Topic %s created", topic.ID())
		}
	}

	return nil
}

func (g *GcpPubSub) CloseClient() error {
	return g.client.Close()
}

func (g *GcpPubSub) Publish(ctx context.Context, topicID QueueTopic, message any) error {
	data, err := json.Marshal(message)
	if err != nil {
		g.log.Error().Err(err).Msg("Failed to marshal message")
		return err
	}

	topic := g.client.Topic(string(topicID) + "_" + viper.GetString("ENV"))
	_, err = topic.Publish(ctx, &pubsub.Message{
		Data: data,
	}).Get(ctx)
	if err != nil {
		g.log.Error().Err(err).Msg("Failed to publish message")
		return err
	}
	return nil
}

func (g *GcpPubSub) Consume(ctx context.Context, topic QueueTopic, handler func(ctx context.Context, msg []byte) error) error {

	subID := fmt.Sprintf("%s_%s_sub", string(topic), viper.GetString("ENV"))

	sub := g.client.Subscription(subID)
	log.Info().Msgf("Consuming messages from %s", sub.ID())

	if err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		if err := handler(ctx, msg.Data); err != nil {
			g.log.Error().Err(err).Msg("Failed to handle message")
			msg.Nack()
			return
		}
		msg.Ack()
	}); err != nil {
		g.log.Error().Err(err).Msg("Failed to receive message")
		return err
	}

	return nil
}
