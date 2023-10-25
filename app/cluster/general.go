package cluster

import (
	"context"
	"ddm-admin-console/router"
	"fmt"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gin-gonic/gin"
)

func (a *App) updateGeneral(ctx *gin.Context) (router.Response, error) {
	platformNameValue := ctx.PostForm("platform-name")
	mainValue := ctx.PostForm("main")
	faviconValue := ctx.PostForm("favicon")
	languageValue := ctx.PostForm("language")

	values, err := getValuesFromGit(a.Config.CodebaseName, masterBranch, a.Gerrit)
	if err != nil {
		return nil, fmt.Errorf("unable to get values contents, %w", err)
	}

	timeStamp := strings.ToLower(time.Now().Format("20060201T150405Z"))

	values.Global.PlatformName = platformNameValue
	values.Global.LogosPath = fmt.Sprintf("configmap:platform-logos-%s", timeStamp)
	values.Global.Language = languageValue

	if err := a.createConfigMapImages(ctx, mainValue, faviconValue, timeStamp); err != nil {
		return nil, fmt.Errorf("unable to update images, %w", err)
	}

	if err := a.createValuesMergeRequestCtx(ctx, MRTypeGeneral, "update cluster general config", values); err != nil {
		return nil, fmt.Errorf("unable to create general merge request, %w", err)
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) loadGeneral(ctx context.Context, values *Values, rspParams gin.H) error {
	parts := strings.Split(values.Global.LogosPath, "-")

	if len(parts) < 3 {
		return fmt.Errorf("unable to parse logo path")
	}

	timeStamp := parts[2]

	logo, err := a.Services.K8S.GetConfigMap(ctx, fmt.Sprintf("platform-logos-%s", timeStamp), "control-plane")

	if err != nil {
		return fmt.Errorf("unable to get logo images, %w", err)
	}

	rspParams["language"] = values.Global.Language
	rspParams["platformName"] = values.Global.PlatformName
	rspParams["logoMain"] = logo.Data["logoMain"]
	rspParams["logoFavicon"] = logo.Data["logoFavicon"]

	return nil
}

func (a *App) createConfigMapImages(ctx context.Context, mainValue string, faviconValue string, timeStamp string) error {
	data := map[string]string{
		"logoMain":    mainValue,
		"logoFavicon": faviconValue,
	}

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("platform-logos-%s", timeStamp),
			Namespace: "control-plane",
		},
		Data: data,
	}

	if err := a.Services.K8S.CreateConfigMap(ctx, cm, "control-plane"); err != nil {
		return fmt.Errorf("unable to create configmap, %w", err)
	}

	return nil
}
