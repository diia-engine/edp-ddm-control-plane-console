package cluster

import (
	"crypto/x509"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type HostnameConfig struct {
	Hostname string `json:"hostname"`
	Cert     string `json:"cert"`
}

func (a *App) uploadPEMDNS(ctx *gin.Context) (rsp router.Response, retErr error) {
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get form file")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read file data")
	}

	pemInfo, err := registry.DecodePEM(data)
	if err != nil {
		return router.MakeJSONResponse(http.StatusUnprocessableEntity, err.Error()), nil
	}

	hostname, ok := ctx.GetPostForm("hostname")
	if !ok {
		return router.MakeStatusResponse(http.StatusUnprocessableEntity), nil
	}

	certPool, _ := x509.SystemCertPool()
	for _, v := range pemInfo.X509CaCACert {
		certPool.AddCert(v)
	}

	if _, err := pemInfo.X509Cert[0].Verify(x509.VerifyOptions{
		DNSName:       hostname,
		CurrentTime:   time.Now(),
		Intermediates: x509.NewCertPool(),
		Roots:         certPool,
	}); err != nil {
		return router.MakeJSONResponse(http.StatusUnprocessableEntity, err.Error()), nil
	}

	vaultPath := fmt.Sprintf("registry-kv/cluster/domains/%s/%s", hostname, time.Now().Format("20060201T150405Z"))

	if err := registry.CreateVaultSecrets(a.Vault, map[string]map[string]interface{}{
		vaultPath: {
			"caCertificate": pemInfo.CACert,
			"certificate":   pemInfo.Cert,
			"key":           pemInfo.PrivateKey,
		},
	}, false); err != nil {
		return nil, fmt.Errorf("unable to create vault secrets, %w", err)
	}

	return router.MakeJSONResponse(http.StatusOK, vaultPath), nil
}

func (a *App) checkKeycloakHostnameUsed(ctx *gin.Context) (router.Response, error) {
	hostname := ctx.Param("hostname")

	cbs, err := a.Codebase.GetAllByType(codebase.RegistryCodebaseType)
	if err != nil {
		return nil, fmt.Errorf("unable to get codebases, %w", err)
	}

	for _, cb := range cbs {
		vals, err := registry.GetValuesFromGit(cb.Name, registry.MasterBranch, a.Gerrit)
		if err != nil {
			return nil, fmt.Errorf("unable to get values for registry, %w", err)
		}

		if vals.Keycloak.CustomHost == hostname {
			return router.MakeStatusResponse(http.StatusUnprocessableEntity), nil
		}
	}

	return router.MakeStatusResponse(http.StatusOK), nil
}

func (a *App) keycloakDNS(ctx *gin.Context) (router.Response, error) {
	hostnames, ok := ctx.GetPostForm("hostnames")
	if !ok {
		return router.MakeStatusResponse(http.StatusUnprocessableEntity), nil
	}

	var hostnamesData []CustomHost
	if err := json.Unmarshal([]byte(hostnames), &hostnamesData); err != nil {
		return nil, fmt.Errorf("unable to decode hostnames, %w", err)
	}

	values, err := a.getValuesDict(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load values, %w", err)
	}

	values.Keycloak.CustomHosts = hostnamesData
	values.OriginalYaml[KeycloakValuesIndex] = values.Keycloak

	if err := a.createValuesMergeRequestCtx(ctx, MRTypeClusterKeycloakDNS, "update cluster keycloak dns",
		values.OriginalYaml); err != nil {
		return nil, fmt.Errorf("unable to create mr, %w", err)
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}
