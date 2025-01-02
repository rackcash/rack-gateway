package nats

import (
	"context"
	"infra/blockchain/rack/config"
	"infra/pkg/nats/natsdomain"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func Init(config *config.Config) (*natsdomain.Ns, jetstream.Consumer) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nc, err := nats.Connect(config.Nats.Servers)
	if err != nil {
		panic(err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "currencies",
		Subjects: natsdomain.SubjectsJetStream[:],
	})
	if err != nil {
		panic(err)
	}

	return &natsdomain.Ns{Nc: nc, Js: js}, initConsumer(ctx, stream)
}

func initConsumer(ctx context.Context, stream jetstream.Stream) jetstream.Consumer {

	c, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:        "CONS", // TODO: remove
		AckPolicy:      jetstream.AckExplicitPolicy,
		FilterSubjects: natsdomain.SubjectsJetStream[:],
		// FilterSubject: "currencies.test",
	})
	if err != nil {
		panic(err)
	}

	return c

}
