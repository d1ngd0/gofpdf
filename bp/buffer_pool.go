package bp

import (
	"bytes"
	"sync"
)

const maxBufferSize = 250 * 1024

// buffer pool to reduce GC
var buffers = sync.Pool{
	// New is called when a new instance is needed
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// GetBuffer fetches a buffer from the pool
func GetBuffer() *bytes.Buffer {
	return buffers.Get().(*bytes.Buffer)
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf *bytes.Buffer) {
	// if the buffer is too big lets just let the GC get rid of it.
	if buf.Cap() > maxBufferSize {
		return
	}

	buf.Reset()
	buffers.Put(buf)
}
