package dashboard

import (
	"ddm-admin-console/router"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (a *App) main(_ *gin.Context) (response router.Response, retErr error) {
	return router.MakeHTMLResponse(200, "dashboard/index.html", gin.H{}), nil
}

func (a *App) dashboard(ctx *gin.Context) (response router.Response, retErr error) {
	components, err := a.edpComponentService.GetAll(ctx, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get edp components")
	}

	var gerritLink, jenkinsLink string
	for _, comp := range components {
		if comp.Spec.Type == "jenkins" {
			jenkinsLink = comp.Spec.Url
		}

		if comp.Spec.Type == "gerrit" {
			gerritLink = comp.Spec.Url
		}
	}

	return router.MakeHTMLResponse(200, "dashboard/dashboard.html", gin.H{
		"page":        "dashboard",
		"gerritLink":  gerritLink,
		"jenkinsLink": jenkinsLink,
	}), nil
}
