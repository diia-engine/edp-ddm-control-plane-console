package codebase

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/yaml.v3"
)

func TestController_Reconcile(t *testing.T) {
	yml := `
global:
  notifications:
    email:
      type: external
      host: smtp.gmail.com
      port: 587
      address: registry@gmail.com
      password: 123
`

	var raw map[string]interface{}
	if err := yaml.Unmarshal([]byte(yml), &raw); err != nil {
		t.Fatal(err)
	}
	t.Log(raw)

	global, ok := raw["global"]
	t.Log(ok)

	globalDict, ok := global.(map[string]interface{})
	t.Log(ok)

	notifications, ok := globalDict["notifications"]
	t.Log(ok)
	t.Log(notifications)
}

func TestErrPostpone_Error(t *testing.T) {
	err := ErrPostpone(time.Second * 5)
	assert.True(t, IsErrPostpone(err))

	err2 := fmt.Errorf("unable to do something, %w", err)
	assert.True(t, IsErrPostpone(err2))

	pp, ok := errors.Unwrap(err2).(ErrPostpone)
	assert.True(t, ok)
	assert.Equal(t, pp.D(), time.Second*5)
}

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

	out := MergeMaps(srcMap, dstMap)

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
