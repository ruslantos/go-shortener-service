package shorten

//func TestHandler_Handle_Success(t *testing.T) {
//	extend := "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf"
//	storage := &MocklinksStorage{}
//	storage.EXPECT().AddLink(extend).Return("short")
//	h := New(storage)
//	req := ShortenRequest{
//		URL: extend,
//	}
//	marshalled, err := json.Marshal(req)
//	assert.NoError(t, err)
//
//	in := &http.Request{
//		Method: http.MethodPost,
//		Body:   io.NopCloser(bytes.NewReader(marshalled)),
//	}
//
//	out := httptest.NewRecorder()
//	c, _ := gin.CreateTestContext(out)
//	c.Request = in
//	h.Handle(c)
//
//	assert.Equal(t, http.StatusCreated, out.Code)
//	assert.Equal(t, `{"result":"http://localhost:8080/short"}`, out.Body.String())
//}

//func TestHandler_Handle_ErrorEmptyBody(t *testing.T) {
//	extend := ""
//	storage := &MocklinksStorage{}
//	h := New(storage)
//	in := &http.Request{
//		Method: http.MethodPost,
//		Body:   io.NopCloser(strings.NewReader(extend)),
//	}
//
//	out := httptest.NewRecorder()
//	c, _ := gin.CreateTestContext(out)
//	c.Request = in
//	h.Handle(c)
//
//	assert.Equal(t, http.StatusBadRequest, out.Code)
//}
