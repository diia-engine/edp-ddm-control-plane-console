package dashboard

import (
	"ddm-admin-console/router"

	"github.com/gin-gonic/gin"
)

func (a *App) main(_ *gin.Context) (response router.Response, retErr error) {
	return router.MakeHTMLResponse(200, "dashboard/index.html", gin.H{}), nil
}
