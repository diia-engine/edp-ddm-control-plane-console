package registry

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type KeyManagement interface {
	KeyDeviceType() string
	AllowedKeysIssuer() []string
	AllowedKeysSerial() []string
	SignKeyIssuer() string
	SignKeyPwd() string
	RemoteType() string
	RemoteSerialNumber() string
	RemoteKeyPort() string
	RemoteKeyHost() string
	RemoteKeyPassword() string
	INIConfig() string
	KeysRequired() bool
	SecretPath() string
}

type DigitalSignature struct {
	Data DigitalSignatureData `yaml:"data" json:"data"`
	Env  DigitalSignatureEnv  `yaml:"env" json:"env"`
}

type DigitalSignatureData struct {
	CACertificates string `yaml:"CACertificates" json:"CACertificates"`
	CAs            string `yaml:"CAs" json:"CAs"`
	Key6Dat        string `yaml:"Key-6-dat" json:"Key-6-dat"`
	AllowedKeysYml string `yaml:"allowed-keys-yml" json:"allowed-keys-yml"`
	OsplmIni       string `yaml:"osplm.ini" json:"osplm.ini"`
}

type DigitalSignatureEnv struct {
	SignKeyDeviceType       string `yaml:"sign.key.device-type" json:"sign.key.device-type"`
	SignKeyFileIssuer       string `yaml:"sign.key.file.issuer" json:"sign.key.file.issuer"`
	SignKeyFilePassword     string `yaml:"sign.key.file.password" json:"sign.key.file.password"`
	SignKeyHardwareDevice   string `yaml:"sign.key.hardware.device" json:"sign.key.hardware.device"`
	SignKeyHardwarePassword string `yaml:"sign.key.hardware.password" json:"sign.key.hardware.password"`
	SignKeyHardwareType     string `yaml:"sign.key.hardware.type" json:"sign.key.hardware.type"`
}

func PrepareRegistryKeys(reg KeyManagement, rq *http.Request, secretData map[string]map[string]interface{}, values map[string]interface{}) (bool, error) {
	createKeys, key6Fl, caCertFl, caJSONFl, err := validateRegistryKeys(rq, reg)
	if err != nil {
		return false, errors.Wrap(err, "unable to validate registry keys")
	}
	if !createKeys {
		return false, nil
	}

	ds := DigitalSignature{
		Env: DigitalSignatureEnv{
			SignKeyDeviceType: reg.KeyDeviceType(),
		},
		Data: DigitalSignatureData{
			CACertificates: reg.SecretPath(),
			CAs:            reg.SecretPath(),
			AllowedKeysYml: reg.SecretPath(),
		},
	}

	keySecretData := make(map[string]interface{})

	if err := setCASecretData(keySecretData, caCertFl, caJSONFl); err != nil {
		return false, errors.Wrap(err, "unable to set ca secret data for registry")
	}

	if err := setKeySecretDataFromRegistry(reg, key6Fl, filesSecretData, envVarsSecretData); err != nil {
		return false, errors.Wrap(err, "unable to set key vars from registry form")
	}

	if err := setAllowedKeysSecretData(filesSecretData, reg); err != nil {
		return false, errors.Wrap(err, "unable to set allowed keys secret data")
	}

	return true, nil
}

func setCASecretData(filesSecretData map[string]interface{}, caCertFl, caJSONFl multipart.File) error {
	caCertBytes, err := ioutil.ReadAll(caCertFl)
	if err != nil {
		return errors.Wrap(err, "unable to read file")
	}
	filesSecretData["CACertificates.p7b"] = string(caCertBytes)

	casJSONBytes, err := ioutil.ReadAll(caJSONFl)
	if err != nil {
		return errors.Wrap(err, "unable to read file")
	}
	filesSecretData["CAs.json"] = string(casJSONBytes)

	return nil
}

/**


  У випадку файлового ключа

      digital-signature:
        data:
          CACertificates: <path to vault>
          CAs: <path to vault>
          Key-6-dat: <path to vault>
          allowed-keys-yml: <path to vault>
          osplm.ini: ""
        env:
          sign.key.device-type: file
          sign.key.file.issuer: <path to vault>
          sign.key.file.password: <path to vault>
          sign.key.hardware.device: ""
          sign.key.hardware.password: ""
          sign.key.hardware.type: ""

  У випадку апаратного ключа

      digital-signature:
        data:
          CACertificates: <path to vault>
          CAs: <path to vault>
          Key-6-dat: ""
          allowed-keys-yml: <path to vault>
          osplm.ini: <path to vault>
        env:
          sign.key.device-type: hardware
          sign.key.file.issuer: ""
          sign.key.file.password: ""
          sign.key.hardware.device: <path to vault>
          sign.key.hardware.password: <path to vault>
          sign.key.hardware.type: <path to vault>



*/

func setKeySecretDataFromRegistry(reg KeyManagement, key6Fl multipart.File,
	keySecretData map[string]interface{}, ds *DigitalSignature) error {

	if reg.KeyDeviceType() == KeyDeviceTypeFile {
		key6Bytes, err := ioutil.ReadAll(key6Fl)
		if err != nil {
			return errors.Wrap(err, "unable to read file")
		}

		keySecretData["Key-6.dat"] = string(key6Bytes)
		keySecretData["sign.key.file.issuer"] = reg.SignKeyIssuer()
		keySecretData["sign.key.file.password"] = reg.SignKeyPwd()

		ds.Env.SignKeyFileIssuer = reg.SecretPath()
		ds.Env.SignKeyFilePassword = reg.SecretPath()
		ds.Data.Key6Dat = reg.SecretPath()
	} else if reg.KeyDeviceType() == KeyDeviceTypeHardware {
		keySecretData["sign.key.hardware.type"] = reg.RemoteType()
		keySecretData["sign.key.hardware.device"] = fmt.Sprintf("%s:%s (%s)",
			reg.RemoteSerialNumber(), reg.RemoteKeyPort(), reg.RemoteKeyHost())
		keySecretData["sign.key.hardware.password"] = reg.RemoteKeyPassword()
		keySecretData["osplm.ini"] = reg.INIConfig()

		ds.Data.OsplmIni = reg.SecretPath()
		ds.Env.SignKeyHardwareDevice = reg.SecretPath()
		ds.Env.SignKeyHardwarePassword = reg.SecretPath()
		ds.Env.SignKeyHardwareType = reg.SecretPath()
	}

	return nil
}

func setAllowedKeysSecretData(filesSecretData map[string][]byte, reg KeyManagement) error {
	//TODO tmp hack, remote in future
	filesSecretData["allowed-keys.yml"] = []byte{}
	//end todo

	allowedKeysIssuer := reg.AllowedKeysIssuer()
	allowedKeysSerial := reg.AllowedKeysSerial()

	if len(allowedKeysIssuer) > 0 {
		var allowedKeysConf allowedKeysConfig
		for i := range allowedKeysIssuer {
			allowedKeysConf.AllowedKeys = append(allowedKeysConf.AllowedKeys, allowedKey{
				Issuer: allowedKeysIssuer[i],
				Serial: allowedKeysSerial[i],
			})
		}
		allowedKeysYaml, err := yaml.Marshal(&allowedKeysConf)
		if err != nil {
			return errors.Wrap(err, "unable to encode allowed keys to yaml")
		}
		filesSecretData["allowed-keys.yml"] = allowedKeysYaml
	}

	return nil
}
