package header

import (
	"corpc/compressor"
	"encoding/binary"
	"errors"
	"sync"
)

// ResponseHeader request header structure looks like:
// +--------------+---------+----------------+-------------+----------+
// | CompressType |    ID   |      Error     | ResponseLen | Checksum |
// +--------------+---------+----------------+-------------+----------+
// |    uint16    | uvarint | uvarint+string |    uvarint  |  uint32  |
// +--------------+---------+----------------+-------------+----------+
type ResponseHeader struct {
	sync.RWMutex
	CompressType CompressType
	ID           uint64
	Error        string
	ResponseLen  uint32
	Checksum     uint32
}

func (r *ResponseHeader) Marshal() []byte {
	r.RLock()
	defer r.RUnlock()
	idx := 0
	header := make([]byte, MaxHeaderSize+len(r.Error))

	binary.LittleEndian.PutUint16(header[idx:], uint16(r.CompressType))
	idx += Uint16Size

	idx += binary.PutUvarint(header[idx:], r.ID)
	idx += writeString(header[idx:], r.Error)
	idx += binary.PutUvarint(header[idx:], uint64(r.ResponseLen))

	binary.LittleEndian.PutUint32(header[idx:], r.Checksum)
	idx += Uint32Size
	return header[:idx]
}

func (r *ResponseHeader) Unmarshal(data []byte) (err error) {
	r.Lock()
	defer r.Unlock()
	if len(data) == 0 {
		return errors.New("resp unmarshal err: len(data) = 0")
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.New("resp unmarshal err")
		}
	}()
	idx, size := 0, 0
	r.CompressType = CompressType(binary.LittleEndian.Uint16(data[idx:]))
	idx += Uint16Size

	r.ID, size = binary.Uvarint(data[idx:])
	idx += size

	r.Error, size = readString(data[idx:])
	idx += size

	length, size := binary.Uvarint(data[idx:])
	r.ResponseLen = uint32(length)
	idx += size

	r.Checksum = binary.LittleEndian.Uint32(data[idx:])
	return
}

func (r *ResponseHeader) GetCompressType() compressor.CompressType {
	r.Lock()
	defer r.Unlock()
	return compressor.CompressType(r.CompressType)
}

func (r *ResponseHeader) Reset() {
	r.Lock()
	defer r.Unlock()
	r.CompressType = 0
	r.ID = 0
	r.Error = ""
	r.ResponseLen = 0
	r.Checksum = 0
}
