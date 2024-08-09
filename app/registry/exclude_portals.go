package registry

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

const (
	excludePortalsIndex = "excludePortals"
)

func (a *App) prepareExcludePortals(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) (bool, error) {
	if len(r.ExcludePortals) != 0 {
		if reflect.DeepEqual(values.OriginalYaml[excludePortalsIndex], r.ExcludePortals) {
			return false, nil
		}

		globalInterface, ok := values.OriginalYaml[GlobalValuesIndex]
		if !ok {
			globalInterface = make(map[string]interface{})
		}
		globalDict := globalInterface.(map[string]interface{})
		globalDict[excludePortalsIndex] = deleteEmptyValue(r.ExcludePortals)
		values.OriginalYaml[GlobalValuesIndex] = globalDict
		return true, nil
	}

	return false, nil
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
