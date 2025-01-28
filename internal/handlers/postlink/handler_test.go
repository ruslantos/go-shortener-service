package postlink

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_Handle_Success(t *testing.T) {
	extend := "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf"
	service := &MocklinksService{}
	service.EXPECT().Add(extend).Return("short", nil)
	h := New(service)
	req, err := http.NewRequest(http.MethodPost, "", io.NopCloser(strings.NewReader(extend)))
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "http://localhost:8080/short", rr.Body.String())
}

func TestHandler_Handle_ErrorEmptyBody(t *testing.T) {
	extend := ""
	service := &MocklinksService{}
	h := New(service)

	req, err := http.NewRequest(http.MethodPost, "", io.NopCloser(strings.NewReader(extend)))
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_Handle_ErrorLinkService(t *testing.T) {
	extend := "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf"
	service := &MocklinksService{}
	service.EXPECT().Add(extend).Return("short", errors.New("some error"))
	h := New(service)
	req, err := http.NewRequest(http.MethodPost, "", io.NopCloser(strings.NewReader(extend)))
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "add short link error: some error\n", rr.Body.String())
}
