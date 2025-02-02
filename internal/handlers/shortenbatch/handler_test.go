package shortenbatch

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ruslantos/go-shortener-service/internal/models"
)

func TestHandler_Handle_Success(t *testing.T) {
	service := &MocklinksService{}
	linksIn := []models.Links{
		{CorrelationID: "123", OriginalURL: "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf"},
		{CorrelationID: "456", OriginalURL: "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf2"},
	}
	linksOut := []models.Links{
		{CorrelationID: "123", OriginalURL: "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf", ShortURL: "qwerty1"},
		{CorrelationID: "456", OriginalURL: "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf2", ShortURL: "qwerty2"},
	}
	service.EXPECT().AddBatch(linksIn).Return(linksOut, nil)
	h := New(service)
	in := ShortenBatchRequest{
		{CorrelationID: linksIn[0].CorrelationID, OriginalURL: linksIn[0].OriginalURL},
		{CorrelationID: linksIn[1].CorrelationID, OriginalURL: linksIn[1].OriginalURL},
	}
	out := ShortenBatchResponse{
		{CorrelationID: linksOut[0].CorrelationID, ShortURL: "http://localhost:8080/qwerty1"},
		{CorrelationID: linksOut[1].CorrelationID, ShortURL: "http://localhost:8080/qwerty2"},
	}
	marshalledIn, err := json.Marshal(in)
	assert.NoError(t, err)
	marshalledOut, err := json.Marshal(out)
	assert.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, "/api/shorten/batch", io.NopCloser(bytes.NewReader(marshalledIn)))
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, string(marshalledOut), rr.Body.String())
}
