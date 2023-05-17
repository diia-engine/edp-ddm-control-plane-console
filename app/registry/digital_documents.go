package registry

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	digitalDocumentsIndex = "digitalDocuments"
)

func (a *App) prepareDigitalDocuments(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) error {
	if r.DigitalDocuments != "" {
		var dd DigitalDocuments
		if err := json.Unmarshal([]byte(r.DigitalDocuments), &dd); err != nil {
			return fmt.Errorf("unable to decode digital documents %w", err)
		}
		values.OriginalYaml[digitalDocumentsIndex] = dd
	}

	return nil
}
