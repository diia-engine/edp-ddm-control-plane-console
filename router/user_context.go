package router

import (
	"context"
	"errors"

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

var ErrTokenNotFound = errors.New("token not found")

func ExtractToken(ctx *gin.Context) (*oauth2.Token, error) {
	session := sessions.Default(ctx)
	token := session.Get(AuthTokenSessionKey)
	if token == nil {
		return nil, ErrTokenNotFound
	}

	tokenData, ok := token.(*oauth2.Token)
	if !ok {
		return nil, ErrTokenNotFound
	}

	return tokenData, nil
}

func ContextWithUserAccessToken(ctx *gin.Context) context.Context {
	tokenData, err := ExtractToken(ctx)
	if err != nil {
		return context.Background()
	}

	return ContextWithUserAccessTokenString(tokenData.AccessToken)
}

func ContextWithUserAccessTokenString(token string) context.Context {
	return context.WithValue(context.Background(), AuthTokenSessionKey, token)
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
