package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

type responseWriterGzip struct {
	gin.ResponseWriter
	gzipWriter *gzip.Writer
	status     int
}

func Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isHeadersExist(c) {
			c.Next()
			return
		}

		var b bytes.Buffer
		body, _ := io.ReadAll(c.Request.Body)
		gz := gzip.NewWriter(&b)
		defer gz.Close()

		writer := &responseWriterGzip{ResponseWriter: c.Writer, gzipWriter: gz}
		c.Writer = writer

		c.Next()

		if err := gz.Close(); err != nil {
			c.Error(err)
			return
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Content-Type", writer.Header().Get("Content-Type"))
		//c.Writer.Write(b.Bytes())

		c.Writer.WriteHeader(writer.status)
		c.Writer.Write(body)
	}
}

// WriteHeader captures the status code.
func (w *responseWriterGzip) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Write writes the response body to the gzip writer.
func (w *responseWriterGzip) Write(data []byte) (int, error) {
	return w.gzipWriter.Write(data)
}

func isHeadersExist(c *gin.Context) bool {
	return (strings.Contains(c.GetHeader("Accept-Encoding"), "gzip")) &&
		(c.GetHeader("Content-Type") == "application/json" ||
			c.GetHeader("Content-Type") == "text/html")
}
