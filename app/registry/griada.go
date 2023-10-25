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
	// current implementation always sends KeyDeviceType, so griada configs are always being owerriten
	// so we should check if RemoteKeyHost or SignKeyIssuer is present in request body (this means that we trying to update keys data)
	// to ensure that we should update griada config
	dataBoutKeysIsUpdating := r.RemoteKeyHost != "" || r.SignKeyIssuer != ""
	if r.KeyDeviceType != "" && dataBoutKeysIsUpdating {
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
