package tracinggin

import (
	"bytes"
	"sync"

	"github.com/yeqown/opentelemetry-quake/pkg"

	"github.com/gin-gonic/gin"
)

type respBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w respBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w respBodyWriter) String() string {
	return pkg.ToString(w.body.Bytes())
}

func (w respBodyWriter) releaseBuffer() {
	if w.body == nil {
		return
	}

	releaseBuffer(w.body)
}

func getResponseBodyWriter(c *gin.Context) *respBodyWriter {
	rbw := &respBodyWriter{
		ResponseWriter: c.Writer,
		body:           getBuffer(),
	}

	return rbw
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 512))
	},
}

func getBuffer() *bytes.Buffer {
	b, ok := bufferPool.Get().(*bytes.Buffer)
	if !ok {
		return bytes.NewBuffer(make([]byte, 0, 512))
	}

	return b
}

func releaseBuffer(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}
