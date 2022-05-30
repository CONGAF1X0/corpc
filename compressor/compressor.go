package compressor

type CompressType uint16

const (
	Raw CompressType = iota
	Gzip
)

var Compressors = map[CompressType]Compressor{
	Raw: RawCompressor{},
	Gzip: GzipCompressor{},
}

type Compressor interface {
	Zip([]byte) ([]byte, error)
	Unzip([]byte) ([]byte, error)
}
