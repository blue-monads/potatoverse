package httpx

import (
	"io"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gin-gonic/gin"
)

func ReadForm(ctx *gin.Context) ([]byte, error) {
	fh, err := ctx.FormFile("file")
	if err != nil {
		qq.Println("1err", err)
		return nil, err
	}

	file, err := fh.Open()
	if err != nil {
		qq.Println("open err", err)
		return nil, err
	}

	defer file.Close()

	return io.ReadAll(file)
}
