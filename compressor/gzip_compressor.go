package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
)

type GzipCompressor struct {
}

func (_ GzipCompressor) Zip(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := gzip.NewWriter(buf)
	defer w.Close()
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (_ GzipCompressor) Unzip(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	r, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err = ioutil.ReadAll(r)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return data, nil
}
