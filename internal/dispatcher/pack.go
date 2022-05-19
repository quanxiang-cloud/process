package dispatcher

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	code        = "code"
	msg         = "msg"
	data        = "data"
	defaultCode = -1
)

// Pack response body pack
type Pack interface {
	Format() map[string]interface{}
	Get() interface{}
}

// Resp response
type Resp map[string]interface{}

type packOpt func(Resp)

// WithError response with error
func WithError(err error) func(Resp) {
	return func(resp Resp) {
		resp[code] = defaultCode
		resp[msg] = err.Error()
	}
}

// WithPack response with pack
func WithPack(pack Pack) func(Resp) {
	return func(resp Resp) {
		if pack == nil {
			return
		}
		for key, value := range pack.Format() {
			resp[key] = value
		}
	}
}

// Format format response
func Format(c *gin.Context, opts ...packOpt) {
	resp := make(map[string]interface{})
	resp[code] = 0

	for _, opt := range opts {
		opt(resp)
	}
	c.JSON(http.StatusOK, resp)
}

// Body response body
type Body struct {
	Data   interface{} `json:"data"`
}

// Format body format
func (b *Body) Format() map[string]interface{} {

	return map[string]interface{}{
		data: b.Data,
		}
}

// Get get data
func (b *Body) Get() interface{} {
	return b.Data
}