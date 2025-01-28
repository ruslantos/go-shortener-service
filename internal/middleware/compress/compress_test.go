package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type testWriter struct {
	*bytes.Buffer
}

func (w *testWriter) WriteHeader(b int) {
	return
}

func (w *testWriter) Header() http.Header {
	return http.Header{}
}

func TestGzipMiddlewareWriter(t *testing.T) {
	buf := bytes.Buffer{}
	w := testWriter{&buf}
	mdl := GzipMiddlewareWriter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	}))

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	mdl.ServeHTTP(&w, req)

	body := buf.Bytes()

	b2 := bytes.NewBuffer(body)
	r, err := gzip.NewReader(b2)
	require.NoError(t, err)
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, "test", string(data))
}
