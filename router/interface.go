package router

import (
	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

type Interface interface {
	GET(relativePath string, handler func(ctx *gin.Context) (Response, error))
	POST(relativePath string, handler func(ctx *gin.Context) (Response, error))
	//ContextWithUserAccessToken(ctx *gin.Context) context.Context
	AddView(route string, view View)
	AddValidator(tag string, valid validator.Func) error
}
