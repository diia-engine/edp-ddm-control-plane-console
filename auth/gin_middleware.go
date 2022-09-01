package auth

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func MakeGinMiddleware(o *OAuth2, tokenSessionKey, tokenValidSessionKey, filterPath string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		tokenValid := tokenIsValid(ctx, tokenSessionKey)
		ctx.Set(tokenValidSessionKey, tokenValid)

		if !strings.Contains(ctx.Request.RequestURI, filterPath) {
			ctx.Next()
			return
		}

		if !tokenValid {
			ginStartAuth(o, ctx)
			return
		}

		ctx.Next()
	}
}

func tokenIsValid(ctx *gin.Context, tokenSessionKey string) bool {
	session := sessions.Default(ctx)
	tsRaw := session.Get(tokenSessionKey)
	if tsRaw == nil {
		return false
	}

	token, ok := tsRaw.(*oauth2.Token)
	if !ok {
		return false
	}

	return token.Valid()
}

func ginStartAuth(o *OAuth2, ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, o.AuthCodeURL())
}
