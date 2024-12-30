package postlink

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Handle_Success(t *testing.T) {
	extend := "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf"
	storage := &MocklinksStorage{}
	storage.EXPECT().AddLink(extend).Return("short")
	h := New(storage)
	in := &http.Request{
		Method: http.MethodPost,
		Body:   io.NopCloser(strings.NewReader(extend)),
	}

	out := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(out)
	c.Request = in
	h.Handle(c)

	assert.Equal(t, http.StatusCreated, out.Code)
	assert.Equal(t, "http://localhost:8080/short", out.Body.String())
}

func TestHandler_Handle_ErrorEmptyBody(t *testing.T) {
	extend := ""
	storage := &MocklinksStorage{}
	h := New(storage)
	in := &http.Request{
		Method: http.MethodPost,
		Body:   io.NopCloser(strings.NewReader(extend)),
	}

	out := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(out)
	c.Request = in
	h.Handle(c)

	assert.Equal(t, http.StatusBadRequest, out.Code)
}

func TestHandler_Handle_ErrorGet(t *testing.T) {
	extend := "123"
	storage := &MocklinksStorage{}
	h := New(storage)
	in := &http.Request{
		Method: http.MethodGet,
		Body:   io.NopCloser(strings.NewReader(extend)),
	}

	out := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(out)
	c.Request = in
	h.Handle(c)

	assert.Equal(t, http.StatusBadRequest, out.Code)
}
