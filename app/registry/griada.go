package registry

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

const (
	griadaIndex = "griada"
)

func (a *App) prepareGriada(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) (bool, error) {
	if r.KeyDeviceType != "" {
		var enabled bool = true

		if r.KeyDeviceType == "file" {
			enabled = false
		}

		griada := Griada{
			Enabled: enabled,
			Ip:      r.RemoteKeyHost,
			Port:    r.RemoteKeyPort,
			Mask:    r.RemoteKeyMask,
		}

		if reflect.DeepEqual(values.OriginalYaml[griadaIndex], griada) {
			return false, nil
		}

		values.OriginalYaml[griadaIndex] = griada
		return true, nil
	}
	return false, nil
}
