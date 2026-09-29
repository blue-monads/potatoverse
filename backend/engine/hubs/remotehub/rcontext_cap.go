package remotehub

import (
	"errors"
	"fmt"

	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
)

type CapRContext struct {
	SpaceId        int64
	PackageId      int64
	PackageVersion int64
	Data           []byte
	ResponseData   []byte
	Meta           map[string]string
}

type SimpleRContext = CapRContext

var _ RContext = (*CapRContext)(nil)

func (s *CapRContext) GetSpaceId() int64 {
	return s.SpaceId
}

func (s *CapRContext) GetPackageId() int64 {
	return s.PackageId
}

func (s *CapRContext) GetPackageVersion() int64 {
	return s.PackageVersion
}

func (s *CapRContext) GetData() ([]byte, error) {
	return s.Data, nil
}

func (s *CapRContext) SetData(data []byte) error {
	s.ResponseData = data
	return nil
}

func (s *CapRContext) GetMeta(key string) (string, error) {
	if s.Meta == nil {
		return "", errors.New("meta key not found: " + key)
	}
	val, ok := s.Meta[key]
	if !ok {
		return "", errors.New("meta key not found: " + key)
	}
	return val, nil
}

func NewCapRContext(spaceId, packageId, packageVersion int64, params lazydata.LazyData) (*CapRContext, error) {
	var dataBytes []byte
	if params != nil {
		var err error
		dataBytes, err = params.AsBytes()
		if err != nil {
			dataBytes = []byte{}
		}
	}

	meta := make(map[string]string)
	if params != nil {
		if m, err := params.AsMap(); err == nil {
			for k, v := range m {
				if s, ok := v.(string); ok {
					meta[k] = s
				} else {
					meta[k] = fmt.Sprintf("%v", v)
				}
			}
		}
	}

	return &CapRContext{
		SpaceId:        spaceId,
		PackageId:      packageId,
		PackageVersion: packageVersion,
		Data:           dataBytes,
		Meta:           meta,
	}, nil
}
