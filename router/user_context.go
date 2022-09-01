package router

import (
	"context"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

const (
	AuthTokenSessionKey                = "access-token"
	AuthTokenValidSessionKey           = "access-token-valid"
	UserNameSessionKey                 = "user-full-name"
	UserEmailSessionKey                = "user-email"
	CanViewClusterManagementSessionKey = "can-view-cluster-management"
	CanViewRegistriesSessionKey        = "can-view-registries"
	CanCreateRegistriesSessionKey      = "can-create-registries"
)

func (r *Router) ContextWithUserAccessToken(ctx *gin.Context) context.Context {
	session := sessions.Default(ctx)
	token := session.Get(AuthTokenSessionKey)
	if token == nil {
		return context.Background()
	}

	tokenData, ok := token.(*oauth2.Token)
	if !ok {
		return context.Background()
	}

	return context.WithValue(context.Background(), AuthTokenSessionKey, tokenData.AccessToken)
}

func UserDataMiddleware(ctx *gin.Context) {
	tokenIsValid := ctx.GetBool(AuthTokenValidSessionKey)
	if !tokenIsValid {
		ctx.Next()
		return
	}

	session := sessions.Default(ctx)
	vars := []string{UserNameSessionKey, CanViewClusterManagementSessionKey, CanViewRegistriesSessionKey,
		CanCreateRegistriesSessionKey, UserEmailSessionKey}

	for _, v := range vars {
		val := session.Get(v)
		if val != nil {
			ctx.Set(v, val)
		}
	}

	ctx.Next()
}
