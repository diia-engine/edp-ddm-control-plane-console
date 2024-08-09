package registry

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

const (
	computeResourcesIndex = "computeResources"
)

func (a *App) prepareComputeResources(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) (bool, error) {
	if r.ComputeResources != "" {
		var computeResources ComputeResources
		if err := json.Unmarshal([]byte(r.ComputeResources), &computeResources); err != nil {
			return false, fmt.Errorf("unable to decode compute resources %w", err)
		}

		if reflect.DeepEqual(values.OriginalYaml[computeResourcesIndex], computeResources) {
			return false, nil
		}

		globalInterface, ok := values.OriginalYaml[GlobalValuesIndex]
		if !ok {
			globalInterface = make(map[string]interface{})
		}
		globalDict := globalInterface.(map[string]interface{})
		globalDict[computeResourcesIndex] = computeResources
		values.OriginalYaml[GlobalValuesIndex] = globalDict
		return true, nil
	}

	return false, nil
}
