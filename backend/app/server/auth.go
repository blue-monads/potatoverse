package server

import (
	"regexp"

	"github.com/gin-gonic/gin"
)

type loginDetails struct {
	Email    string `db:"email"`
	Password string `db:"password"`
}

var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

func (a *Server) login(ctx *gin.Context) {
	// data := &loginDetails{}
	// err := ctx.Bind(data)
	// if err != nil {
	// 	httpx.WriteAuthErr(ctx, err)
	// 	return
	// }

	// var user *models.User

	// if strings.Contains(data.Email, "@") {
	// 	user, err = a.db.GetUserByEmail(data.Email)
	// 	if err != nil {
	// 		httpx.WriteAuthErr(ctx, err)
	// 		return
	// 	}
	// } else if phoneRegex.MatchString(data.Email) {
	// 	panic("Implement login by phone")
	// } else {
	// 	panic("Implement login by username")
	// }

	// // fixme => hash it
	// if user.Password != data.Password {
	// 	httpx.WriteAuthErr(ctx, err)
	// 	return
	// }

	// token, err := a.signer.SignAccess(&signer.AccessClaim{
	// 	XID:    xid.New().String(),
	// 	UserId: int64(user.ID),
	// })

	// if err != nil {
	// 	httpx.WriteErr(ctx, err)
	// 	return
	// }

	// ctx.JSON(http.StatusOK, gin.H{
	// 	"access_token": token,
	// })

}
