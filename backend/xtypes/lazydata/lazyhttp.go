package lazydata

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type LazyHTTP struct {
	ctx      *gin.Context
	once     sync.Once
	raw      []byte
	readErr  error
}

func NewLazyHTTP(ctx *gin.Context) *LazyHTTP {
	return &LazyHTTP{ctx: ctx}
}

func (l *LazyHTTP) ensureRead() {
	l.once.Do(func() {
		l.raw, l.readErr = io.ReadAll(l.ctx.Request.Body)
	})
}

func (l *LazyHTTP) AsBytes() ([]byte, error) {
	l.ensureRead()
	return l.raw, l.readErr
}

func (l *LazyHTTP) AsMap() (map[string]any, error) {
	l.ensureRead()
	if l.readErr != nil {
		return nil, l.readErr
	}

	var data map[string]any
	if err := json.Unmarshal(l.raw, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (l *LazyHTTP) AsJson(target any) error {
	l.ensureRead()
	if l.readErr != nil {
		return l.readErr
	}
	return json.Unmarshal(l.raw, target)
}

func (l *LazyHTTP) GetFieldAsInt(path string) int {
	l.ensureRead()
	if l.readErr != nil {
		return 0
	}
	return int(gjson.GetBytes(l.raw, path).Int())
}

func (l *LazyHTTP) GetFieldAsFloat(path string) float64 {
	l.ensureRead()
	if l.readErr != nil {
		return 0
	}
	return gjson.GetBytes(l.raw, path).Float()
}

func (l *LazyHTTP) GetFieldAsString(path string) string {
	l.ensureRead()
	if l.readErr != nil {
		return ""
	}
	return gjson.GetBytes(l.raw, path).String()
}

func (l *LazyHTTP) GetFieldAsBool(path string) bool {
	l.ensureRead()
	if l.readErr != nil {
		return false
	}
	return gjson.GetBytes(l.raw, path).Bool()
}
