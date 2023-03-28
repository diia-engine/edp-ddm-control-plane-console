package merge_request

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMergeMaps(t *testing.T) {
	dstYaml := `global:
  crunchyPostgres:
    storageSize: 10Gi
    backups:
      postgresql:
        parameters:
          max_connections: 200
      pgbackrest:
        repos:
          schedules:
            full: 0 15 * * *`

	srcYaml := `global:
  crunchyPostgres:
    storageSize: 10Gi
    backups:
      postgresql:
        parameters:
          max_connections: 200
      deploymentMode: production`

	var srcMap, dstMap map[string]interface{}

	err := yaml.Unmarshal([]byte(srcYaml), &srcMap)
	assert.NoError(t, err)

	err = yaml.Unmarshal([]byte(dstYaml), &dstMap)
	assert.NoError(t, err)

	out := mergeMaps(srcMap, dstMap)

	global, ok := out["global"]
	assert.True(t, ok)

	globalDict, ok := global.(map[string]interface{})
	assert.True(t, ok)

	crunchyPostgres, ok := globalDict["crunchyPostgres"]
	assert.True(t, ok)

	crunchyPostgresDict, ok := crunchyPostgres.(map[string]interface{})
	assert.True(t, ok)

	backups, ok := crunchyPostgresDict["backups"]
	assert.True(t, ok)

	backupsDict, ok := backups.(map[string]interface{})
	assert.True(t, ok)

	_, ok = backupsDict["postgresql"]
	assert.True(t, ok)

	_, ok = backupsDict["pgbackrest"]
	assert.True(t, ok)

	_, ok = backupsDict["deploymentMode"]
	assert.True(t, ok)
}
