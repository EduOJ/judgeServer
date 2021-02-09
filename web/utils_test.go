package web_test

import (
	"encoding/json"
	"github.com/leoleoasd/EduOJBackend/app/response"
	"github.com/stretchr/testify/assert"
	"github.com/suntt2019/EduOJJudger/base"
	"github.com/suntt2019/EduOJJudger/web"
	"net/http"
	"strconv"
	"testing"
)

func backendError(wr http.ResponseWriter, r *http.Request) {
	// r.RequestURI[12:15] get three digits status number
	statusCode, err := strconv.ParseInt(r.RequestURI[12:15], 10, 64)
	if err != nil {
		panic(err)
	}
	wr.WriteHeader(int(statusCode))
	// r.RequestURI[16:] remove "/backendErr/xxx/" xxx is the status number
	resp := response.ErrorResp(r.RequestURI[16:], nil)
	j, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	_, err = wr.Write(j)
	if err != nil {
		panic(err)
	}
}

func TestHandleBackendErrorResponse(t *testing.T) {
	t.Parallel()
	t.Run("PermissionDeniedResponse", func(t *testing.T) {
		t.Parallel()
		resp, err := R().Get("/backendErr/403/PERMISSION_DENIED")
		assert.Nil(t, err)
		assert.Equal(t, base.ErrPermissionDeniedResponse, web.HandleBackendErrorResponse(resp))
	})
	t.Run("NotFoundResponse", func(t *testing.T) {
		t.Parallel()
		resp, err := R().Get("/backendErr/404/NOT_FOUND")
		assert.Nil(t, err)
		assert.Equal(t, base.ErrNotFoundResponse, web.HandleBackendErrorResponse(resp))
	})
	t.Run("InternalErrorResponse", func(t *testing.T) {
		t.Parallel()
		resp, err := R().Get("/backendErr/500/INTERNAL_ERROR")
		assert.Nil(t, err)
		assert.Equal(t, base.ErrBackendErrorResponse, web.HandleBackendErrorResponse(resp))
	})
	t.Run("UnexpectedMessageResponse", func(t *testing.T) {
		t.Parallel()
		resp, err := R().Get("/backendErr/400/UNEXPECTED_MESSAGE")
		assert.Nil(t, err)
		assert.Equal(t, base.ErrUnexpectedMessageResponse, web.HandleBackendErrorResponse(resp))
	})
	t.Run("FailToUnmarshal", func(t *testing.T) {
		t.Parallel()
		resp, err := R().Get("/echoURI/test_handle_error_response_fail_to_unmarshal")
		assert.Nil(t, err)
		assert.Nil(t, web.HandleBackendErrorResponse(resp))
	})
}
