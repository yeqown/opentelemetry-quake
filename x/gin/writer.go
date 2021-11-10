package otelgin

import (
	"bytes"

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
