package getlink

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_Handle_Success(t *testing.T) {
	storage := &MocklinksStorage{}
	storage.EXPECT().GetLink("short").Return("extend", true)
	h := New(storage)
	req, err := http.NewRequest(http.MethodGet, "short", nil)
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	assert.Equal(t, "extend", rr.Header().Get("Location"))
}

func TestHandler_Handle_BadRequest(t *testing.T) {
	storage := &MocklinksStorage{}
	storage.EXPECT().GetLink("short").Return("", false)
	h := New(storage)
	req, err := http.NewRequest(http.MethodGet, "short", nil)
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
