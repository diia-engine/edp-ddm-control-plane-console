package registry

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	KeyManagementVaultPath = "key-management"
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
	VaultSecretPath() string
	KeyDataChanged() bool
	KeyVerificationChanged() bool
}

type DigitalSignature struct {
	Data DigitalSignatureData `yaml:"data" json:"data"`
	Env  DigitalSignatureEnv  `yaml:"env" json:"env"`
}

type DigitalSignatureData struct {
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

func PrepareRegistryKeys(reg KeyManagement, rq *http.Request, secretData map[string]map[string]interface{},
	values map[string]interface{}, repoFiles map[string]string) (bool, error) {

	if reg.KeyVerificationChanged() {
		caCertFl, _, err := rq.FormFile("ca-cert")
		if err != nil {
			return false, fmt.Errorf("no ca-cert file")
		}

		caJSONFl, _, err := rq.FormFile("ca-json")
		if err != nil {
			return false, fmt.Errorf("no ca-json file")
		}

		if err := setCASecretData(repoFiles, caCertFl, caJSONFl); err != nil {
			return false, fmt.Errorf("unable to set ca secret data for registry, err: %w", err)
		}
	}

	if reg.KeyDataChanged() {
		ds := DigitalSignature{
			Env: DigitalSignatureEnv{
				SignKeyDeviceType: reg.KeyDeviceType(),
			},
			Data: DigitalSignatureData{
				AllowedKeysYml: reg.VaultSecretPath(),
			},
		}

		keySecretData := make(map[string]interface{})

		if err := setKeySecretDataFromRegistry(reg, rq, keySecretData, &ds); err != nil {
			return false, fmt.Errorf("unable to set key vars from registry form, err: %w", err)
		}

		if err := setAllowedKeysSecretData(reg, keySecretData, &ds); err != nil {
			return false, fmt.Errorf("unable to set allowed keys secret data, err: %w", err)
		}

		secretData[reg.VaultSecretPath()] = keySecretData
		values["digital-signature"] = ds
	}

	return reg.KeyVerificationChanged() || reg.KeyDataChanged(), nil
}

func setCASecretData(repoFiles map[string]string, caCertFl, caJSONFl multipart.File) error {
	caCertBytes, err := ioutil.ReadAll(caCertFl)
	if err != nil {
		return fmt.Errorf("unable to read file, err: %w", err)
	}
	repoFiles["config/dso/CACertificates.p7b"] = string(caCertBytes)

	casJSONBytes, err := ioutil.ReadAll(caJSONFl)
	if err != nil {
		return fmt.Errorf("unable to read file, err: %w", err)
	}
	repoFiles["config/dso/CAs.json"] = string(casJSONBytes)

	return nil
}

func setKeySecretDataFromRegistry(reg KeyManagement, rq *http.Request,
	keySecretData map[string]interface{}, ds *DigitalSignature) error {

	if reg.KeyDeviceType() == KeyDeviceTypeFile {
		key6Fl, _, err := rq.FormFile("key6")
		if err != nil {
			return fmt.Errorf("unable to get key6 file, %w", err)
		}

		key6Bytes, err := ioutil.ReadAll(key6Fl)
		if err != nil {
			return errors.Wrap(err, "unable to read file")
		}

		keySecretData["Key-6.dat"] = base64.StdEncoding.EncodeToString(key6Bytes)
		keySecretData["sign.key.file.issuer"] = reg.SignKeyIssuer()
		keySecretData["sign.key.file.password"] = reg.SignKeyPwd()

		ds.Env.SignKeyFileIssuer = reg.VaultSecretPath()
		ds.Env.SignKeyFilePassword = reg.VaultSecretPath()
		ds.Data.Key6Dat = reg.VaultSecretPath()
	} else if reg.KeyDeviceType() == KeyDeviceTypeHardware {
		keySecretData["sign.key.hardware.type"] = reg.RemoteType()
		keySecretData["sign.key.hardware.device"] = fmt.Sprintf("%s:%s (%s)",
			reg.RemoteSerialNumber(), reg.RemoteKeyPort(), reg.RemoteKeyHost())
		keySecretData["sign.key.hardware.password"] = reg.RemoteKeyPassword()
		keySecretData["osplm.ini"] = reg.INIConfig()

		ds.Data.OsplmIni = reg.VaultSecretPath()
		ds.Env.SignKeyHardwareDevice = reg.VaultSecretPath()
		ds.Env.SignKeyHardwarePassword = reg.VaultSecretPath()
		ds.Env.SignKeyHardwareType = reg.VaultSecretPath()
	}

	return nil
}

func setAllowedKeysSecretData(reg KeyManagement, keySecretData map[string]interface{}, ds *DigitalSignature) error {
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
		keySecretData["allowed-keys.yml"] = string(allowedKeysYaml)
		ds.Data.AllowedKeysYml = reg.VaultSecretPath()
	}

	return nil
}
