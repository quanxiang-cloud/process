package rpc

import (
	"context"
	"git.internal.yunify.com/qxp/misc/logger"
	"git.internal.yunify.com/qxp/process/pkg/config"
	"git.internal.yunify.com/qxp/process/rpc/pb"
	"github.com/tal-tech/go-zero/zrpc"
	"google.golang.org/grpc/metadata"
)

// Client rpc client
type Client struct {
	c zrpc.Client
}

// NewClient new a client
func NewClient(conf *config.Configs) *Client {
	c := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{conf.FlowRPCServer},
		App:       "process",
		Token:     "process",
		Timeout:   20000,
	})
	return &Client{
		c: c,
	}
}

// NodeEventPublish Publish
func (c *Client) NodeEventPublish(ctx context.Context, req *pb.NodeEventReq) (*pb.NodeEventResp, error) {
	conn := pb.NewNodeEventClient(c.c.Conn())

	headersCTX := logger.STDHeader(ctx)
	reqHeader := metadata.New(headersCTX)
	ctx = metadata.NewOutgoingContext(ctx, reqHeader)

	return conn.Event(ctx, req)
}
