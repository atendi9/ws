package xhttp

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"sync"

	"github.com/atendi9/ws/io/pkg/forge"
)

var (
	// gzipWriterPool is a [sync.Pool] for reusing [gzip.Writer] instances.
	gzipWriterPool = sync.Pool{
		New: func() any {
			w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
			return w
		},
	}
	// flateWriterPool is a [sync.Pool] for reusing [flate.Writer] instances.
	flateWriterPool = sync.Pool{
		New: func() any {
			w, _ := flate.NewWriter(io.Discard, flate.DefaultCompression)
			return w
		},
	}
)

type Compression struct {
	Threshold int `json:"threshold,omitempty"`
}

type PerMessageDeflate = Compression

// Compress takes an [io.Reader] and returns a [forge.Interface] compressed using the specified encoding.
// If the encoding is unknown, it returns the data as is.
func (c *Compression) Compress(encoding string, data io.Reader) (forge.Interface, error) {
	buf := forge.NewBytesBuffer(nil)
	switch encoding {
	case "gzip":
		gz := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gz)

		gz.Reset(buf)
		_, err := io.Copy(gz, data)
		if err != nil {
			return nil, err
		}
		if err := gz.Close(); err != nil {
			return nil, err
		}
		return buf, nil

	case "deflate":
		fl := flateWriterPool.Get().(*flate.Writer)
		defer flateWriterPool.Put(fl)

		fl.Reset(buf)
		_, err := io.Copy(fl, data)
		if err != nil {
			return nil, err
		}
		if err := fl.Close(); err != nil {
			return nil, err
		}
		return buf, nil

	default:
		_, err := io.Copy(buf, data)
		return buf, err
	}
}

// NewWriter returns an [io.WriteCloser] that compresses data written to it using the specified encoding.
// It is intended to be used with an [http.ResponseWriter].
func (c *Compression) NewWriter(w io.Writer, encoding string) (io.WriteCloser, error) {
	switch encoding {
	case "gzip":
		return gzip.NewWriterLevel(w, gzip.DefaultCompression)
	case "deflate":
		return flate.NewWriter(w, flate.DefaultCompression)
	default:
		return nil, fmt.Errorf("unsupported compression encoding: %s", encoding)
	}
}
