package codebase

import (
	"testing"

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
