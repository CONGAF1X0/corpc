package corpc

import (
	"corpc/codec"
	"corpc/compressor"
	"corpc/serializer"
	"io"
	"net/rpc"
)

type Client struct {
	*rpc.Client
}

type options struct {
	compressType compressor.CompressType
	serializer   serializer.Serializer
}

type Option func(o *options)

func WithCompress(c compressor.CompressType) Option {
	return func(o *options) {
		o.compressType = c
	}
}

func WithSerializer(s serializer.Serializer) Option {
	return func(o *options) {
		o.serializer = s
	}
}

func NewClient(conn io.ReadWriteCloser, opts ...Option) *Client {
	options := options{
		compressType: compressor.Raw,
		serializer:   serializer.Proto,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return &Client{rpc.NewClientWithCodec(codec.NewClientCodec(conn, options.compressType, options.serializer))}
}

func (c *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return c.Client.Call(serviceMethod, args, reply)
}

func (c *Client) AsyncCall(serviceMethod string, args interface{}, reply interface{}) chan *rpc.Call {
	return c.Client.Go(serviceMethod, args, reply, nil).Done
}
