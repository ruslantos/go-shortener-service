package postlink

//func TestHandler_Handle_Success(t *testing.T) {
//	extend := "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf"
//	storage := &MocklinksStorage{}
//	storage.EXPECT().AddLink(extend).Return("short")
//	filesSevice := &Mockfile{}
//	filesSevice.EXPECT().WriteEvent(&files.Event{
//		ID:          mock.Anything,
//		ShortURL:    "short",
//		OriginalURL: extend,
//	}).Return(nil)
//	h := New(storage, filesSevice)
//	req, err := http.NewRequest(http.MethodPost, "", io.NopCloser(strings.NewReader(extend)))
//	assert.NoError(t, err)
//	rr := httptest.NewRecorder()
//
//	h.Handle(rr, req)
//
//	assert.Equal(t, http.StatusCreated, rr.Code)
//	assert.Equal(t, "http://localhost:8080/short", rr.Body.String())
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
