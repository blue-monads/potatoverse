package httpx

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

func ReadForm(ctx *gin.Context) ([]byte, error) {
	fh, err := ctx.FormFile("file")
	if err != nil {
		pp.Println("1err", err)
		return nil, err
	}

	file, err := fh.Open()
	if err != nil {
		pp.Println("open err", err)
		return nil, err
	}

	defer file.Close()

	return io.ReadAll(file)
}
