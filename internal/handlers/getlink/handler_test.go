package getlink

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Handle_Success(t *testing.T) {
	storage := &MocklinksStorage{}
	storage.EXPECT().GetLink("short").Return("extend", true)
	h := New(storage)
	in := &http.Request{
		Method: http.MethodGet,
		URL:    pointer.To(url.URL{Path: "short"})}

	out := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(out)
	c.Request = in
	h.Handle(c)

	assert.Equal(t, http.StatusTemporaryRedirect, out.Code)
	assert.Equal(t, "extend", out.Header().Get("Location"))
}

func TestHandler_Handle_BadRequest(t *testing.T) {
	storage := &MocklinksStorage{}
	storage.EXPECT().GetLink("short").Return("", false)
	h := New(storage)
	in := &http.Request{
		Method: http.MethodGet,
		URL:    pointer.To(url.URL{Path: "short"})}

	out := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(out)
	c.Request = in
	h.Handle(c)

	assert.Equal(t, http.StatusBadRequest, out.Code)
	assert.Equal(t, "", out.Header().Get("Location"))
}

// func TestHandler_Handle_InvalidPath(t *testing.T) {
//	storage := &MocklinksStorage{}
//	storage.EXPECT().GetLink("short").Return("", false)
//	h := New(storage)
//	in := &http.Request{
//		Method: http.MethodGet,
//		URL:    pointer.To(url.URL{Path: ""})}
//
//	out := httptest.NewRecorder()
//	c, _ := gin.CreateTestContext(out)
//	c.Request = in
//	h.Handle(c)
//
//	assert.Equal(t, http.StatusBadRequest, out.Code)
//	assert.Equal(t, "", out.Header().Get("Location"))
//}
