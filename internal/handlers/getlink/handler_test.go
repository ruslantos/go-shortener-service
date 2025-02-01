package getlink

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_Handle_Success(t *testing.T) {
	service := &MocklinksService{}
	service.EXPECT().Get("short").Return("extend", nil)
	h := New(service)
	req, err := http.NewRequest(http.MethodGet, "short", nil)
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	assert.Equal(t, "extend", rr.Header().Get("Location"))
}

func TestHandler_Handle_BadRequest(t *testing.T) {
	storage := &MocklinksService{}
	storage.EXPECT().Get("short").Return("", errors.New("some error"))
	h := New(storage)
	req, err := http.NewRequest(http.MethodGet, "short", nil)
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "failed to get long ling: some error\n", rr.Body.String())
}
