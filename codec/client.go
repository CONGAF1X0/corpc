package codec

import (
	"bufio"
	"corpc/compressor"
	"corpc/header"
	"corpc/serializer"
	"hash/crc32"
	"io"
	"net/rpc"
	"sync"
)

type clientCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	CompressType compressor.CompressType
	serializer   serializer.Serializer
	response     header.ResponseHeader
	mu           sync.Mutex
	pending      map[uint64]string
}

func NewClientCodec(conn io.ReadWriteCloser, compressType compressor.CompressType,
	serializer serializer.Serializer) rpc.ClientCodec {
	return &clientCodec{
		r:            bufio.NewReader(conn),
		w:            bufio.NewWriter(conn),
		c:            conn,
		CompressType: compressType,
		serializer:   serializer,
		pending:      make(map[uint64]string),
	}
}

func (c *clientCodec) WriteRequest(request *rpc.Request, param interface{}) error {
	c.mu.Lock()
	c.pending[request.Seq] = request.ServiceMethod
	c.mu.Unlock()

	if _, ok := compressor.Compressors[c.CompressType]; !ok {
		return NotFoundCompressorError
	}
	reqBody, err := c.serializer.Marshal(param)
	if err != nil {
		return err
	}
	compressedReqBody, err := compressor.Compressors[c.CompressType].Zip(reqBody)
	if err != nil {
		return err
	}
	h := header.RequestPool.Get().(*header.RequestHeader)
	defer func() {
		h.ResetHeader()
		header.RequestPool.Put(h)
	}()
	h.ID = request.Seq
	h.Method = request.ServiceMethod
	h.RequestLen = uint32(len(compressedReqBody))
	h.Checksum = crc32.ChecksumIEEE(compressedReqBody)

	if err = sendFrame(c.w, h.Marshal()); err != nil {
		return err
	}
	if err = write(c.w, compressedReqBody); err != nil {
		return err
	}
	c.w.(*bufio.Writer).Flush()
	return nil
}

func (c *clientCodec) ReadResponseHeader(response *rpc.Response) error {
	c.response.Reset()
	data, err := recvFrame(c.r)
	if err != nil {
		return err
	}
	if err = c.response.Unmarshal(data); err != nil {
		return err
	}
	c.mu.Lock()
	response.Seq = c.response.ID
	response.Error = c.response.Error
	response.ServiceMethod = c.pending[response.Seq]
	delete(c.pending, response.Seq)
	c.mu.Unlock()
	return nil
}

func (c *clientCodec) ReadResponseBody(param interface{}) error {
	if param == nil {
		if c.response.ResponseLen != 0 {
			if err := read(c.r, make([]byte, c.response.ResponseLen)); err != nil {
				return err
			}
			return nil
		}
	}
	respBody := make([]byte, c.response.ResponseLen)
	if err := read(c.r, respBody); err != nil {
		return err
	}
	if c.response.Checksum != 0 {
		if crc32.ChecksumIEEE(respBody) != c.response.Checksum {
			return UnexpectedChecksumError
		}
	}
	if _, ok := compressor.Compressors[c.response.GetCompressType()]; !ok {
		return NotFoundCompressorError
	}
	resp, err := compressor.Compressors[c.response.GetCompressType()].Unzip(respBody)
	if err != nil {
		return err
	}
	return c.serializer.Unmarshal(resp, param)
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}
