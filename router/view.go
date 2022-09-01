package router

import "github.com/gin-gonic/gin"

type View interface {
	Get(ctx *gin.Context) (*Response, error)
	Post(ctx *gin.Context) (*Response, error)
}

func (r *Router) AddView(route string, view View) {
	r.GET(route, view.Get)
	r.POST(route, view.Post)
}
