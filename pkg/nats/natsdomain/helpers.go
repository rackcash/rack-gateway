package natsdomain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func (ns *Ns) JsPublish(subj string, jsonMsg []byte) error {
	return ns.jsPublishOpts(subj, jsonMsg)
}

// jetstream publish with msgId
func (ns *Ns) JsPublishMsgId(subj string, jsonMsg []byte, msgId string) error {
	return ns.jsPublishOpts(subj, jsonMsg, jetstream.WithMsgID(msgId))
}

func (ns *Ns) jsPublishOpts(subj string, jsonMsg []byte, opts ...jetstream.PublishOpt) error {
	_, err := ns.Js.Publish(context.Background(), subj, jsonMsg, opts...)
	if err != nil {
		return err
	}
	return nil
}

// nats core
func (ns *Ns) ReqAndRecv(subject SubjType, jsonMsg []byte) ([]byte, error) {
	var reconnects int = 4
	var timeout time.Duration = 7 * time.Second
	var err error
	var response *nats.Msg

	fmt.Println("TIMEOUT", timeout, "RECONNECTS", reconnects)

	for reconnects > 0 {
		response, err = sendrecv(ns.Nc, timeout, subject, jsonMsg)
		if err != nil {
			fmt.Printf("NATS ERROR: %v. Subj: %s, jsonMsg: %s\n", err, subject.String(), string(jsonMsg))
			if errors.Is(err, nats.ErrNoResponders) {
				return []byte{0}, err
			}
			reconnects -= 1
			continue
		}
		break
	}

	if err != nil {
		return []byte{0}, err
	}

	if response != nil {
		return response.Data, nil
	}
	//  TODO: сделать проверку через n.HelpersIsError

	return []byte{0}, fmt.Errorf("unknown error: data == nil && err == nil")
}

func (ns *Ns) InitBuckets(ctx context.Context) error {
	for _, bucket := range KvBuckets {
		_, err := ns.Js.CreateKeyValue(ctx, jetstream.KeyValueConfig{Bucket: bucket})
		if err != nil {
			if errors.Is(err, jetstream.ErrBucketExists) {
				fmt.Printf("Bucket %s already exists\n", bucket)
				continue
			}
			return err
		}
	}
	return nil
}

func (ns *Ns) KvWatch(ctx context.Context, bucket BucketType, key string) (jetstream.KeyWatcher, error) {
	kv, err := kvGet(ns.Js, bucket.String())
	if err != nil {
		return nil, err
	}

	watcher, err := kv.Watch(ctx, key)
	if err != nil {
		return nil, err
	}
	return watcher, nil

}

func kvGet(js jetstream.JetStream, bucket string) (jetstream.KeyValue, error) {
	return js.KeyValue(context.Background(), bucket)
}

func (ns *Ns) KvPut(bucket BucketType, key string, value []byte) {
	fmt.Println("SET KEY VALUE", key)
	kv, err := kvGet(ns.Js, bucket.String())
	if err != nil {
		panic(err)
	}

	_, err = kv.Put(context.Background(), key, value)
	if err != nil {
		panic(err)
	}
}

func (ns *Ns) KVPutStr(bucket BucketType, key string, value string) {
	kv, err := kvGet(ns.Js, bucket.String())
	if err != nil {
		panic(err)
	}

	_, err = kv.PutString(context.Background(), key, value)
	if err != nil {
		panic(err)
	}
}

func (ns *Ns) KVGet(bucket BucketType, key string) (jetstream.KeyValueEntry, error) {
	kv, err := kvGet(ns.Js, bucket.String())
	if err != nil {
		panic(err)
	}

	return kv.Get(context.Background(), key)
}

func sendrecv(nc *nats.Conn, timeout time.Duration, subj SubjType, jsonMsg []byte) (*nats.Msg, error) {
	resp, err := nc.Request(subj.String(), jsonMsg, timeout)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// for nats jetstream
func NewMsgId(invoiceId string, action ActionType) string {
	// return invoiceId + "_" + uuid.NewString(
	return invoiceId + "_" + string(action)
}
