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
