package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Logger interface {
	Error(msg string, fields ...zap.Field)
}

type Router struct {
	engine *gin.Engine
	logger Logger
}

type Response struct {
	code         int
	viewTemplate string
	params       gin.H
	isRedirect   bool
}

func MakeResponse(code int, viewTemplate string, params gin.H) *Response {
	return &Response{
		code:         code,
		viewTemplate: viewTemplate,
		params:       params,
	}
}

func MakeStatusResponse(code int) *Response {
	return &Response{code: code}
}

func MakeRedirectResponse(code int, path string) *Response {
	return &Response{
		code:         code,
		viewTemplate: path,
		isRedirect:   true,
	}
}

func (r *Router) makeViewResponder(handler func(ctx *gin.Context) (*Response, error)) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		rsp, err := handler(ctx)
		if err != nil {
			r.logger.Error(fmt.Sprintf("%+v", err))
			ctx.String(500, "%+v", err)
			return
		}

		if rsp == nil {
			r.logger.Error("empty view response")
			ctx.String(500, "empty view response")
			return
		}

		if rsp.isRedirect {
			ctx.Redirect(rsp.code, rsp.viewTemplate)
			return
		}

		if rsp.viewTemplate == "" {
			ctx.Status(rsp.code)
			return
		}

		rsp.params = r.parseValidationErrors(rsp.params)
		rsp.params = r.includeSessionVars(ctx, rsp.params)

		ctx.HTML(rsp.code, rsp.viewTemplate, rsp.params)
	}
}

func (r *Router) includeSessionVars(ctx *gin.Context, params gin.H) gin.H {
	params["username"] = ctx.GetString(UserNameSessionKey)
	params["canViewRegistries"] = ctx.GetBool(CanViewRegistriesSessionKey) || ctx.GetBool(CanCreateRegistriesSessionKey)
	params["canViewClusterManagement"] = ctx.GetBool(CanViewClusterManagementSessionKey)

	return params
}

func (r *Router) GET(relativePath string, handler func(ctx *gin.Context) (*Response, error)) {
	r.engine.GET(relativePath, r.makeViewResponder(handler))
}

func (r *Router) POST(relativePath string, handler func(ctx *gin.Context) (*Response, error)) {
	r.engine.POST(relativePath, r.makeViewResponder(handler))
}

func Make(engine *gin.Engine, logger Logger) *Router {
	return &Router{
		engine: engine,
		logger: logger,
	}
}
