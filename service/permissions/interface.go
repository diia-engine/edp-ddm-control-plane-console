package permissions

import (
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/k8s"

	"github.com/gin-gonic/gin"
)

type ServiceInterface interface {
	SetPermission(token string, registryName string, permission RegistryPermission)
	GetPermission(token, registryName string) (*RegistryPermission, error)
	DeleteToken(tok string)
	DeleteRegistry(name string)
	FilterCodebases(ginContext *gin.Context, cbs []codebase.Codebase, k8sService k8s.ServiceInterface) ([]codebase.WithPermissions, error)
	LoadUserRegistries(ctx *gin.Context) error
	DeleteTokenContext(ctx *gin.Context) error
}
