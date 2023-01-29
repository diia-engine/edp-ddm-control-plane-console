package registry

import "github.com/gin-gonic/gin"

const (
	supAuthBrowserFlowWidget  = "dso-officer-auth-flow"
	supAuthBrowserFlowIdGovUa = "id-gov-ua-officer-redirector"
)

func (a *App) prepareSupplierAuthConfig(ctx *gin.Context, r *registry, values map[string]interface{},
	secrets map[string]map[string]interface{}) error {

	if r.SupAuthBrowserFlow == supAuthBrowserFlowWidget {

	} else if r.SupAuthBrowserFlow == supAuthBrowserFlowIdGovUa {

	}

	return nil
}
