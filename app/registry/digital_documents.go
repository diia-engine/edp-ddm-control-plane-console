package registry

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

const (
	digitalDocumentsIndex = "digitalDocuments"
)

func (a *App) prepareDigitalDocuments(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) (bool, error) {
	if r.DigitalDocuments != "" {
		var dd map[string]interface{}
		if err := json.Unmarshal([]byte(r.DigitalDocuments), &dd); err != nil {
			return false, fmt.Errorf("unable to decode digital documents %w", err)
		}

		if reflect.DeepEqual(values.OriginalYaml[digitalDocumentsIndex], dd) {
			return false, nil
		}

		values.OriginalYaml[digitalDocumentsIndex] = dd
		return true, nil
	}

	return false, nil
}
