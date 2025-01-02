package nats

import (
	"context"
	"fmt"
	"infra/api/internal/config"
	"infra/api/internal/logger"
	"infra/pkg/nats/natsdomain"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NatsInfra struct {
	*natsdomain.Ns
}

func Init(config *config.Config, log logger.Logger) *NatsInfra {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nc, err := nats.Connect(config.Nats.Servers,
		nats.MaxReconnects(100),
		nats.ReconnectWait(3*time.Second),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			log.TemplNatsInfo("disconnected", nc.ConnectedUrl())
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.TemplNatsInfo("reconnected", nc.ConnectedUrl())
		}))
	if err != nil {
		log.TemplNatsError("Connect failed", nc.ConnectedUrl(), err)
		os.Exit(0)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		panic(err)
	}

	// initStream(ctx, js)

	InitResponsesStream(ctx, js)

	msg, err := nc.Request(natsdomain.SubjPing.String(), []byte("ping"), time.Second*5) // check connection
	if err != nil {
		panic("NATS: connect failed: " + err.Error())
	}
	if string(msg.Data) != "pong" {
		panic("NATS: wrong response")
	}

	fmt.Println("nats: Connected to", nc.ConnectedAddr())
	// defer nc.Drain()
	return &NatsInfra{&natsdomain.Ns{Nc: nc, Js: js}}
}

func InitResponsesStream(ctx context.Context, js jetstream.JetStream) (jetstream.Stream, error) {
	return js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "responses",
		Subjects: natsdomain.ResponseSubjects[:],
	})
}
