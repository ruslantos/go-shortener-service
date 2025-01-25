package middleware

import (
	"bytes"
	"compress/gzip"
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
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		var b bytes.Buffer
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
		c.Writer.Write(b.Bytes())
	}
}
