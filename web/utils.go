package web

import (
	"encoding/json"
	"encoding/xml"
	"github.com/go-resty/resty/v2"
	"github.com/leoleoasd/EduOJBackend/app/response"
	"github.com/minio/minio-go/v7"
	"github.com/suntt2019/EduOJJudger/base"
	"net/http"
)

// convert response to backend errors, if the response isn't a backend error, return nil
func HandleBackendErrorResponse(httpResp *resty.Response) error {
	resp := response.Response{}
	if err := json.Unmarshal(httpResp.Body(), &resp); err != nil {
		return nil
	}
	switch httpResp.StatusCode() {
	case http.StatusForbidden:
		if resp == response.ErrorResp("PERMISSION_DENIED", nil) {
			return base.ErrPermissionDeniedResponse
		}
	case http.StatusNotFound:
		if resp == response.ErrorResp("NOT_FOUND", nil) {
			return base.ErrNotFoundResponse
		}
	case http.StatusInternalServerError:
		if resp.Message == "INTERNAL_ERROR" {
			return base.ErrBackendErrorResponse
		}
	}
	return base.ErrUnexpectedMessageResponse
}

// convert response to storage errors, if the response isn't a storage error, return nil
func HandleStorageErrorResponse(httpResp *resty.Response) error {
	resp := minio.ErrorResponse{}
	if err := xml.Unmarshal(httpResp.Body(), &resp); err != nil {
		return nil
	}
	switch httpResp.StatusCode() {
	case http.StatusNotFound:
		return base.ErrStorageNotFound
	case http.StatusForbidden:
		return base.ErrStorageAccessDenied
	}
	return base.ErrStorageOtherError
}
