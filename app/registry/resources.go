package registry

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func (a *App) prepareRegistryResources(r *registry, values map[string]interface{}) error {
	if r.Resources != "" {
		var resources map[string]interface{}
		if err := json.Unmarshal([]byte(r.Resources), &resources); err != nil {
			return errors.Wrap(err, "unable to decode resources")
		}

		global, ok := values["global"]
		if !ok {
			global = make(map[string]interface{})
		}

		globalDict := global.(map[string]interface{})

		globalDict["registry"] = resources
		values["global"] = globalDict
	}

	return nil
}
