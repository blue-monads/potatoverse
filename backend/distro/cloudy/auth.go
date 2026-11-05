package cloudy

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hako/branca"
)

const (
	UTypeAdmin  = "admin"
	UTypeNormal = "normal"

	purposeAccess = "access"
	purposeVerify = "verify"

	ctxClaimKey = "cloudy_claim"
	ctxUserKey  = "cloudy_user"
)

type Claim struct {
	UserID  int64  `json:"uid"`
	Email   string `json:"email"`
	UType   string `json:"utype"`
	Slug    string `json:"slug,omitempty"`
	Purpose string `json:"purpose"`
}

func newBranca(masterSecret string) *branca.Branca {
	sum := sha256.Sum256([]byte(masterSecret))
	b := branca.NewBranca(string(sum[:]))
	b.SetTTL(60 * 60 * 24 * 7) // 7 days
	return b
}

func (a *CloudyApp) encodeClaim(claim *Claim) (string, error) {
	raw, err := json.Marshal(claim)
	if err != nil {
		return "", err
	}
	return a.branca.EncodeToString(string(raw))
}

func (a *CloudyApp) decodeClaim(token string) (*Claim, error) {
	raw, err := a.branca.DecodeToString(token)
	if err != nil {
		return nil, err
	}
	claim := &Claim{}
	if err := json.Unmarshal([]byte(raw), claim); err != nil {
		return nil, err
	}
	if claim.UserID == 0 || (claim.UType != UTypeAdmin && claim.UType != UTypeNormal) {
		return nil, fmt.Errorf("invalid claim")
	}

	return claim, nil
}

func extractBearer(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	for _, prefix := range []string{"Bearer ", "TokenV1 "} {
		if len(header) >= len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
			return strings.TrimSpace(header[len(prefix):])
		}
	}
	return header
}

func (a *CloudyApp) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := extractBearer(c.GetHeader("Authorization"))
		if tok == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			c.Abort()
			return
		}

		claim, err := a.decodeClaim(tok)
		if err != nil || claim.Purpose != purposeAccess {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		user, err := a.resolveClaimUser(claim)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			c.Abort()
			return
		}
		if user.IsDisabled {
			c.JSON(http.StatusForbidden, gin.H{"error": "account disabled"})
			c.Abort()
			return
		}
		if !user.IsVerified {
			c.JSON(http.StatusForbidden, gin.H{"error": "email not verified"})
			c.Abort()
			return
		}

		// refresh claim from db in case it changed
		claim.UserID = user.ID
		claim.UType = user.UType
		claim.Email = user.Email

		c.Set(ctxClaimKey, claim)
		c.Set(ctxUserKey, user)
		c.Next()
	}
}

func (a *CloudyApp) adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claim := getClaim(c)
		if claim == nil || claim.UType != UTypeAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func getClaim(c *gin.Context) *Claim {
	v, ok := c.Get(ctxClaimKey)
	if !ok {
		return nil
	}
	claim, _ := v.(*Claim)
	return claim
}

func getUser(c *gin.Context) *User {
	v, ok := c.Get(ctxUserKey)
	if !ok {
		return nil
	}
	user, _ := v.(*User)
	return user
}

func validUType(utype string) bool {
	return utype == UTypeAdmin || utype == UTypeNormal
}
