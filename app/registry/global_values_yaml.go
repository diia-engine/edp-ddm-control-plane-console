package registry

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	computeResourcesIndex = "computeResources"
	excludePortalsIndex   = "excludePortals"
	deploymentMode        = "deploymentMode"
	geoServerEnabled      = "geoServerEnabled"
	language              = "language"
)

func (a *App) prepareGlobalValuesYaml(
	ctx *gin.Context,
	r *registry,
	values *Values,
	_ map[string]map[string]any, _ *[]string,
) (
	bool,
	error,
) {
	globalInterface, ok := values.OriginalYaml[GlobalValuesIndex]
	if !ok {
		globalInterface = make(map[string]any)
	}

	globalDict := globalInterface.(map[string]any)

	globalDict[language] = r.Language

	if strings.Contains(ctx.FullPath(), "registry/create") {
		globalDict[deploymentMode] = r.DeploymentMode
		globalDict[geoServerEnabled] = r.GeoServerEnabled == "on"
	}

	if _, err := prepareComputeResources(r.ComputeResources, values, globalDict); err != nil {
		return false, fmt.Errorf("%w", err)
	}

	prepareExcludePortals(r.ExcludePortals, values, globalDict)

	values.OriginalYaml[GlobalValuesIndex] = globalDict

	return true, nil
}

func prepareExcludePortals(excludePortalsForm []string, values *Values, globalDict map[string]any) {
	var excludePortals = deleteEmptyValue(excludePortalsForm)

	if !reflect.DeepEqual(values.OriginalYaml[excludePortalsIndex], excludePortals) {
		globalDict[excludePortalsIndex] = excludePortals
	}
}

func prepareComputeResources(computeResourcesForm string, values *Values, globalDict map[string]any) (bool, error) {
	if computeResourcesForm != "" {
		var computeResources ComputeResources
		if err := json.Unmarshal([]byte(computeResourcesForm), &computeResources); err != nil {
			return false, fmt.Errorf("unable to decode compute resources %w", err)
		}

		if !reflect.DeepEqual(values.OriginalYaml[computeResourcesIndex], computeResources) {
			globalDict[computeResourcesIndex] = computeResources
		}
		return true, nil
	}
	return true, nil
}

func deleteEmptyValue(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
