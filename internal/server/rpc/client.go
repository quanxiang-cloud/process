package rpc

import (
	"context"
	"errors"
	"github.com/quanxiang-cloud/process/internal/pb"
	"github.com/quanxiang-cloud/process/pkg/config"

	"github.com/tal-tech/go-zero/zrpc"
)

const processBridge = "flow:9081"

// Client rpc client
type Client struct {
	c zrpc.Client
}

// NewClient new a client
func NewClient(conf *config.Configs) *Client {
	c := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{processBridge},
		App:       "process",
		Token:     "process",
		Timeout:   20000,
	})
	return &Client{
		c: c,
	}
}

// EventPublishReq event req
type EventPublishReq struct {
	EventType string
	EventName string
	EventData map[string]string
}

// Publish Publish
func (c *Client) Publish(req *EventPublishReq) error {
	rp := &pb.EventReq{
		EventType: req.EventType,
		EventName: req.EventName,
		EventData: req.EventData,
	}
	conn := pb.NewEventClient(c.c.Conn())
	r, err := conn.Publish(context.Background(), rp)
	if err != nil {
		return err
	}
	if r.Result != "success" {
		return errors.New(r.Result)
	}
	return nil
}
