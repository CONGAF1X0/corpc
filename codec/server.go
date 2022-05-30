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

type serverCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	request    header.RequestHeader
	serializer serializer.Serializer
	mutex      sync.Mutex
	seq        uint64
	pending    map[uint64]uint64
}

func NewServerCodec(conn io.ReadWriteCloser, serializer serializer.Serializer) rpc.ServerCodec {
	return &serverCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		serializer: serializer,
		pending:    make(map[uint64]uint64),
	}
}

func (s *serverCodec) ReadRequestHeader(request *rpc.Request) error {
	s.request.ResetHeader()
	data, err := recvFrame(s.r)
	if err != nil {
		return err
	}
	if err = s.request.Unmarshal(data); err != nil {
		return err
	}
	s.mutex.Lock()
	s.seq++
	s.pending[s.seq] = s.request.ID
	request.ServiceMethod = s.request.Method
	request.Seq = s.seq
	s.mutex.Unlock()
	return nil
}

func (s *serverCodec) ReadRequestBody(param interface{}) error {
	if param == nil {
		if s.request.RequestLen != 0 {
			if err := read(s.r, make([]byte, s.request.RequestLen)); err != nil {
				return err
			}
			return nil
		}
	}
	reqBody := make([]byte, s.request.RequestLen)
	err := read(s.r, reqBody)
	if err != nil {
		return err
	}
	if s.request.Checksum != 0 {
		if crc32.ChecksumIEEE(reqBody) != s.request.Checksum {
			return UnexpectedChecksumError
		}
	}
	var req []byte
	if c, ok := compressor.Compressors[s.request.GetCompressType()]; ok {
		req, err = c.Unzip(reqBody)
		if err != nil {
			return err
		}
	} else {
		return NotFoundCompressorError
	}
	return s.serializer.Unmarshal(req, param)
}

func (s *serverCodec) WriteResponse(response *rpc.Response, param interface{}) error {
	s.mutex.Lock()
	id, ok := s.pending[response.Seq]
	if !ok {
		s.mutex.Unlock()
		return InvalidSequenceError
	}
	delete(s.pending, response.Seq)
	s.mutex.Unlock()

	if response.Error != "" {
		param = nil
	}
	if _, ok := compressor.
		Compressors[s.request.GetCompressType()]; !ok {
		return NotFoundCompressorError
	}

	var respBody []byte
	var err error
	if param != nil {
		respBody, err = s.serializer.Marshal(param)
		if err != nil {
			return err
		}
	}
	compressedRespBody, err := compressor.Compressors[s.request.GetCompressType()].Zip(respBody)
	if err != nil {
		return err
	}
	h := header.ResponsePool.Get().(*header.ResponseHeader)
	defer func() {
		h.Reset()
		header.ResponsePool.Put(h)
	}()
	h.ID = id
	h.Error = response.Error
	h.ResponseLen = uint32(len(compressedRespBody))
	h.Checksum = crc32.ChecksumIEEE(compressedRespBody)
	h.CompressType = s.request.CompressType
	if err = sendFrame(s.w, h.Marshal()); err != nil {
		return err
	}
	if err = write(s.w, compressedRespBody); err != nil {
		return err
	}
	s.w.(*bufio.Writer).Flush()
	return nil
}

func (s *serverCodec) Close() error {
	return s.c.Close()
}
