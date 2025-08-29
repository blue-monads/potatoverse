package httpx

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ztrue/tracerr"
)

var (
	MessageOk = []byte(`{"message": "success"}`)
)

func WriteJSON(c *gin.Context, resp any, err error) {
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if resp == nil {
		WriteOk(c)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func WriteFinal(c *gin.Context, err error) {
	if err != nil {
		WriteErr(c, err)
		return
	}
	WriteOk(c)
}

func WriteAuthErr(c *gin.Context, err error) {
	c.JSON(http.StatusUnauthorized, gin.H{"message": (err.Error())})
}

func WriteErr(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"message": (err.Error())})
}

func WriteErrString(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"message": msg})
}

func WriteOk(c *gin.Context) {
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(MessageOk)
}

func WriteBinary(c *gin.Context, data []byte) {
	w := c.Writer
	w.Write(data)
	w.Header().Del("Content-Type")
	w.Header().Add("Content-Type", "application/octet-stream")
}

func ApiBaseURL(host, tenantId string) string {
	return fmt.Sprintf("http://%s/z/api/%s/v2", host, tenantId)
}

type Rutil struct{}

func (r *Rutil) WriteJSON(c *gin.Context, resp any, err error) {
	if err != nil {
		tracerr.PrintSourceColor(err)
	}

	WriteJSON(c, resp, err)
}

func (r *Rutil) WriteFinal(c *gin.Context, err error) {
	if err != nil {
		tracerr.PrintSourceColor(err)
	}

	WriteFinal(c, err)
}

func (r *Rutil) WriteErr(c *gin.Context, msg string) {
	WriteErrString(c, msg)
}
