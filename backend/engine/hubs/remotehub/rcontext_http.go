package remotehub

import (
	"bytes"
	"errors"
	"io"

	"github.com/gin-gonic/gin"
)

type HttpBindContext struct {
	Http           *gin.Context
	PackageId      int64
	PackageVersion int64
	SpaceId        int64
	RequestId      string

	bodyData     []byte
	responseData []byte
}

var _ RContext = (*HttpBindContext)(nil)

func (h *HttpBindContext) GetSpaceId() int64 {
	return h.SpaceId
}

func (h *HttpBindContext) GetPackageId() int64 {
	return h.PackageId
}

func (h *HttpBindContext) GetPackageVersion() int64 {
	return h.PackageVersion
}

func (h *HttpBindContext) GetData() ([]byte, error) {
	if h.bodyData != nil {
		return h.bodyData, nil
	}
	if h.Http == nil || h.Http.Request == nil || h.Http.Request.Body == nil {
		return []byte{}, nil
	}
	data, err := io.ReadAll(h.Http.Request.Body)
	if err != nil {
		return nil, err
	}
	h.bodyData = data
	h.Http.Request.Body = io.NopCloser(bytes.NewBuffer(data))
	return data, nil
}

func (h *HttpBindContext) SetData(data []byte) error {
	h.responseData = data
	return nil
}

func (h *HttpBindContext) GetResponseData() []byte {
	return h.responseData
}

func (h *HttpBindContext) GetMeta(key string) (string, error) {
	if h.Http == nil {
		return "", errors.New("http context is nil")
	}
	if val := h.Http.Param(key); val != "" {
		return val, nil
	}
	if val := h.Http.Query(key); val != "" {
		return val, nil
	}
	if val := h.Http.GetHeader(key); val != "" {
		return val, nil
	}
	return "", errors.New("meta key not found: " + key)
}
