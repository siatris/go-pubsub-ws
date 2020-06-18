package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-pubsub-ws/pkg/pubsub"
	"go-pubsub-ws/pkg/websocket"

	"github.com/joho/godotenv"
)

func handleShutdown(cancel func()) {
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)
	<-signChan

	cancel()
}

type AuthClaims struct {
	UserName string
	IsAdmin  bool
}

func main() {
	godotenv.Load()

	ctx, done := context.WithCancel(context.Background())

	wss := websocket.New(
		websocket.WithAuthentication(func(token string, conn *websocket.WSConn) (interface{}, error) {
			// Authenticate user, can check database here.
			// Then return error or AuthClaims for the user
			return AuthClaims{
				IsAdmin:  true,
				UserName: "test",
			}, nil
		}),
	)
	go wss.Run(ctx, ":8080")

	var redisURL = "localhost:6379"
	if rURL := os.Getenv("REDIS_URL"); rURL != "" {
		redisURL = rURL
	}
	ps := pubsub.New(
		pubsub.WithWSService(wss),
		pubsub.WithRedisURL(redisURL),
	)

	ps.HandlePublication("notification", func(pub *pubsub.Publication) error {
		// If return error, publication wont be forwarded to subscribers
		var claims AuthClaims
		var ok bool
		if claims, ok = pub.Origin().AuthClaims().(AuthClaims); !ok || !claims.IsAdmin {
			return errors.New("Permission denied")
		}

		return nil
	})

	ps.HandleSubscription("notification", func(conn *websocket.WSConn) (pubsub.Subscription, error) {
		// Add here any kind of checks
		if conn.ID().String() == "" {
			// If function return nil, it means the subscription was rejected.
			return nil, errors.New("Something wrong happened")
		}

		// If everything is fine, create subscription resolving the name of the channel to a private channel
		subscription := ps.CreateSubscription(ctx, conn, fmt.Sprintf("notification-%s", conn.AuthClaims().(AuthClaims).UserName))

		// Let's add a middleware to do some kind of message manipulation
		subscription.Use(func(msg *websocket.WSMessage) *websocket.WSMessage {
			// This example middleware would change every message on this subscription to the following:
			return websocket.NewMessage(msg.Conn(), "notification", []string{"Fake Message"})
		})

		// Then we return the subscription
		return subscription, nil
	})

	go ps.Run(ctx)

	go handleShutdown(done)

	fmt.Println("Running on :8080")
	<-ctx.Done()
}
