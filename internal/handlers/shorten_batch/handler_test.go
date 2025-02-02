package shorten_batch

//func TestHandler_Handle_Success(t *testing.T) {
//	extend := "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf"
//	service := &MocklinksService{}
//	service.EXPECT().Add(extend).Return("short", nil)
//	h := New(service)
//	in := ShortenRequest{
//		URL: extend,
//	}
//	marshalled, err := json.Marshal(in)
//	assert.NoError(t, err)
//	req, err := http.NewRequest(http.MethodPost, "/api/shorten", io.NopCloser(bytes.NewReader(marshalled)))
//	assert.NoError(t, err)
//	rr := httptest.NewRecorder()
//
//	h.Handle(rr, req)
//	assert.Equal(t, http.StatusCreated, rr.Code)
//	assert.Equal(t, `{"result":"http://localhost:8080/short"}`, rr.Body.String())
//}
//
//func TestHandler_Handle_Error(t *testing.T) {
//	extend := ""
//	service := &MocklinksService{}
//	service.EXPECT().Add(extend).Return("short", errors.New("some error"))
//	h := New(service)
//	in := ShortenRequest{
//		URL: extend,
//	}
//	marshalled, err := json.Marshal(in)
//	assert.NoError(t, err)
//	req, err := http.NewRequest(http.MethodPost, "/api/shorten", io.NopCloser(bytes.NewReader(marshalled)))
//	assert.NoError(t, err)
//	rr := httptest.NewRecorder()
//
//	h.Handle(rr, req)
//	assert.Equal(t, http.StatusInternalServerError, rr.Code)
//}
