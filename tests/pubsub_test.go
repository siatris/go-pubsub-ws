package test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/siatris/go-pubsub-ws/pkg/client"
	"github.com/siatris/go-pubsub-ws/pkg/pubsub"
	"github.com/siatris/go-pubsub-ws/pkg/websocket"

	"github.com/go-redis/redis/v8"
)

func waitForRedis(ctx context.Context, redisClient redis.UniversalClient) error {
	var status *redis.StatusCmd
	interval := time.NewTicker(1 * time.Second)
	timeout := time.NewTimer(20 * time.Second)
	for {
		status = redisClient.Ping(ctx)
		select {
		case <-interval.C:
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return status.Err()
		}
	}
}

func createPubSub(ctx context.Context) *pubsub.PubSubService {
	wss := websocket.New(
		websocket.WithAuthentication(func(token string, conn websocket.WSConn) (interface{}, error) {
			return "test", nil
		}),
	)

	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{"redis:6379"},
		Password: "",
		DB:       0,
	})

	err := waitForRedis(ctx, redisClient)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %+v", err)
	}

	ps := pubsub.New(
		pubsub.WithWSService(wss),
		pubsub.WithRedisProvider(redisClient),
	)

	ps.HandlePublication("test", func(pub *pubsub.Publication) error {
		return nil
	})

	ps.HandleSubscription("test", func(conn websocket.WSConn) (pubsub.Subscription, error) {
		subscription := ps.CreateSubscription(ctx, conn, "test")

		subscription.Use(func(msg websocket.WSMessage) websocket.WSMessage {
			return msg
		})

		return subscription, nil
	})

	return ps
}

func TestPubsub(t *testing.T) {
	ctx := context.Background()
	ps := createPubSub(ctx)
	go ps.Run(ctx)

	cli, err := client.Connect(
		client.WithHost("localhost:8080"),
		client.WithAuthToken("test"),
	)
	if err != nil {
		t.Fatalf("Failed to connect to server: %+v", err)
	}

	testMsgs := make(chan websocket.WSMessage, 0)
	unsubscribe, err := cli.Subscribe("test", func(msg websocket.WSMessage) {
		testMsgs <- msg
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to channel: %+v", err)
	}

	cli.Publish("test", "hello")

	defer unsubscribe()
	defer cli.Close()

	timeout := time.NewTimer(10 * time.Second)
	select {
	case msg := <-testMsgs:
		if msg.Type() != "test" || msg.Data().(string) != "hello" {
			t.Fatal("Invalid message received")
		}
	case <-timeout.C:

		t.Fatal("Failed to receive message in time")
	}
}
