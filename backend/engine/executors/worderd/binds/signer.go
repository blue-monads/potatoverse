package binds

import (
	"fmt"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func (s *BindingServer) handleSigner(method string, params []any, ctx *xtypes.HttpEvent) (any, error) {
	signer := s.app.Signer()
	switch method {
	case "parse_space":
		if len(params) < 1 {
			return nil, fmt.Errorf("missing token")
		}
		token := params[0].(string)
		return signer.ParseSpace(token)
	}
	return nil, fmt.Errorf("unknown signer method: %s", method)
}
