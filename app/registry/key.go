package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"reflect"

	"ddm-admin-console/service/vault"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	KeyManagementVaultPath = "key-management"
	fileKeyPasswordIndex   = "sign.key.file.password"
	fileKeyDataIndex       = "Key-6.dat"
	hardKeyPasswordIndex   = "sign.key.hardware.password"
	hardKeyOsplmIniIndex   = "osplm.ini"
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
	Data DigitalSignatureData               `yaml:"data" json:"data"`
	Env  DigitalSignatureEnv                `yaml:"env" json:"env"`
	Keys map[string]DigitalSignatureKeyData `yaml:"keys" json:"keys"`
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

type DigitalSignatureKeyData struct {
	DeviceType        string   `yaml:"device-type,omitempty" json:"device-type"`
	File              string   `yaml:"file,omitempty" json:"file"`
	Password          string   `yaml:"password,omitempty" json:"password"`
	Issuer            string   `yaml:"issuer,omitempty" json:"issuer"`
	Type              string   `yaml:"type,omitempty" json:"type"`
	Device            string   `yaml:"device,omitempty" json:"device"`
	OsplmIni          string   `yaml:"osplm.ini,omitempty" json:"osplm.ini"`
	AllowedRegistries []string `yaml:"allowedRegistries,omitempty" json:"allowedRegistries"`
}

func PrepareRegistryKeys(
	reg KeyManagement,
	rq *http.Request,
	secretData map[string]map[string]any,
	values *Values,
	repoFiles map[string]string,
	vaultService vault.ServiceInterface,
) (bool, error) {
	tab := rq.PostFormValue("tab")
	keysManagementChanged := false
	if tab == "keysManagement" {
		keySecretData := make(map[string]map[string]any)
		ds := values.DigitalSignature
		valuesChanged, err := setKeysArray(reg.VaultSecretPath(), rq, keySecretData, &ds, vaultService)
		if err != nil {
			return false, fmt.Errorf("unable to set key vars from registry form, err: %w", err)
		}
		if len(keySecretData) > 0 {
			for k := range keySecretData {
				secretData[k] = keySecretData[k]
			}
		}
		values.OriginalYaml["digital-signature"] = ds
		keysManagementChanged = valuesChanged
	}
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
		values.OriginalYaml["digital-signature"] = ds
	}

	return reg.KeyVerificationChanged() || reg.KeyDataChanged() || keysManagementChanged, nil
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
	keySecretData map[string]interface{}, ds *DigitalSignature,
) error {
	if reg.KeyDeviceType() == KeyDeviceTypeFile {
		key6Fl, _, err := rq.FormFile("key6")
		if err != nil {
			return fmt.Errorf("unable to get key6 file, %w", err)
		}

		key6Bytes, err := ioutil.ReadAll(key6Fl)
		if err != nil {
			return errors.Wrap(err, "unable to read file")
		}

		keySecretData[fileKeyDataIndex] = base64.StdEncoding.EncodeToString(key6Bytes)
		keySecretData["sign.key.file.issuer"] = reg.SignKeyIssuer()
		keySecretData[fileKeyPasswordIndex] = reg.SignKeyPwd()

		ds.Env.SignKeyFileIssuer = reg.VaultSecretPath()
		ds.Env.SignKeyFilePassword = reg.VaultSecretPath()
		ds.Data.Key6Dat = reg.VaultSecretPath()
	} else if reg.KeyDeviceType() == KeyDeviceTypeHardware {
		keySecretData["sign.key.hardware.type"] = reg.RemoteType()
		keySecretData["sign.key.hardware.device"] = fmt.Sprintf("%s:%s (%s)",
			reg.RemoteSerialNumber(), reg.RemoteKeyPort(), reg.RemoteKeyHost())
		keySecretData[hardKeyPasswordIndex] = reg.RemoteKeyPassword()
		keySecretData[hardKeyOsplmIniIndex] = reg.INIConfig()

		ds.Data.OsplmIni = reg.VaultSecretPath()
		ds.Env.SignKeyHardwareDevice = reg.VaultSecretPath()
		ds.Env.SignKeyHardwarePassword = reg.VaultSecretPath()
		ds.Env.SignKeyHardwareType = reg.VaultSecretPath()
	}

	return nil
}

func setKeysArray(
	secretPath string,
	rq *http.Request,
	keySecretData map[string]map[string]any,
	ds *DigitalSignature,
	vaultService vault.ServiceInterface,
) (bool, error) {
	keysJson := rq.PostFormValue("keysJSON")
	osplmIni := rq.PostFormValue("osplmIni")
	var keys []struct {
		DeviceType          string   `json:"deviceType,omitempty"`
		HardKeyName         string   `json:"hardKeyName,omitempty"`
		HardKeyType         string   `json:"hardKeyType,omitempty"`
		HardKeyPassword     string   `json:"hardKeyPassword,omitempty"`
		HardKeySerialNumber string   `json:"hardKeySerialNumber,omitempty"`
		HardKeyIssuer       string   `json:"hardKeyIssuer,omitempty"`
		HardKeyIssuerPort   string   `json:"hardKeyIssuerPort,omitempty"`
		HardKeyIssuerHost   string   `json:"hardKeyIssuerHost,omitempty"`
		HardKeyHost         string   `json:"hardKeyHost,omitempty"`
		HardKeyPort         string   `json:"hardKeyPort,omitempty"`
		HardKeyDevice       string   `json:"hardKeyDevice,omitempty"`
		FileKeyName         string   `json:"fileKeyName,omitempty"`
		FileKeyPassword     string   `json:"fileKeyPassword,omitempty"`
		FileKeyIssuer       string   `json:"fileKeyIssuer,omitempty"`
		FileKeyFile         string   `json:"fileKeyFile,omitempty"`
		AllowedRegistries   []string `json:"allowedRegistries,omitempty"`
	}
	if err := json.Unmarshal([]byte(keysJson), &keys); err != nil {
		return false, fmt.Errorf("unable to decode keys from request %w", err)
	}
	valuesChanged := false

	// remove keys
	var keysToRemove []string
	for storedKeyName := range ds.Keys {
		storedKeyExistsInRequest := false
		for _, incomingKey := range keys {
			if storedKeyName == incomingKey.HardKeyName || storedKeyName == incomingKey.FileKeyName {
				storedKeyExistsInRequest = true
				break
			}
		}
		if !storedKeyExistsInRequest {
			keysToRemove = append(keysToRemove, storedKeyName)
		}
	}
	for _, key := range keysToRemove {
		delete(ds.Keys, key)
	}
	if len(keysToRemove) > 0 {
		valuesChanged = true
	}
	if ds.Keys == nil {
		ds.Keys = make(map[string]DigitalSignatureKeyData)
	}

	// check if osplmIni changed
	storedOsplmIni := vaultService.GetPropertyFromVault(ds.Data.OsplmIni, hardKeyOsplmIniIndex)
	if storedOsplmIni != osplmIni && osplmIni != "" {
		valuesChanged = true
		// Check if the inner map exists, if not, initialize it
		if keySecretData[secretPath] == nil {
			keySecretData[secretPath] = make(map[string]any)
		}
		keySecretData[secretPath][hardKeyOsplmIniIndex] = osplmIni
		ds.Data.OsplmIni = secretPath
	}

	// add keys
	for _, key := range keys {
		if key.DeviceType == KeyDeviceTypeFile {
			if storedKey, keyExists := ds.Keys[key.FileKeyName]; keyExists {
				key6Fl, _, _ := rq.FormFile(key.FileKeyName)
				if key6Fl == nil && reflect.DeepEqual(storedKey.AllowedRegistries, key.AllowedRegistries) {
					// file not present in input data, so file is not updated, so keys are same
					continue
				}
				// if only allowed registries are changed, no need to update the key
				if !reflect.DeepEqual(ds.Keys[key.FileKeyName].AllowedRegistries, key.AllowedRegistries) {
					valuesChanged = true
					modifiedKey := ds.Keys[key.FileKeyName]
					modifiedKey.AllowedRegistries = key.AllowedRegistries
					ds.Keys[key.FileKeyName] = modifiedKey
					continue
				}
				storedKey.Password = vaultService.GetPropertyFromVault(storedKey.Password, fileKeyPasswordIndex)
				storedKey.File = vaultService.GetPropertyFromVault(storedKey.File, fileKeyDataIndex)
				key6Bytes, _ := io.ReadAll(key6Fl)
				inputKeyFileData := base64.StdEncoding.EncodeToString(key6Bytes)
				if storedKey.Issuer == key.FileKeyIssuer && storedKey.File == inputKeyFileData && storedKey.Password == key.FileKeyPassword && reflect.DeepEqual(storedKey.AllowedRegistries, key.AllowedRegistries) {
					// input and stored keys are same
					continue
				}
			}

			key6Fl, _, err := rq.FormFile(key.FileKeyName)
			if err != nil {
				return false, fmt.Errorf("unable to get key6 file, %w", err)
			}
			key6Bytes, err := io.ReadAll(key6Fl)
			if err != nil {
				return false, fmt.Errorf("unable to read file %w", err)
			}

			currentFileKeyPath := secretPath + "-" + key.FileKeyName
			// Check if the inner map exists, if not, initialize it
			if keySecretData[currentFileKeyPath] == nil {
				keySecretData[currentFileKeyPath] = make(map[string]any)
			}
			keySecretData[currentFileKeyPath][fileKeyDataIndex] = base64.StdEncoding.EncodeToString(key6Bytes)
			keySecretData[currentFileKeyPath][fileKeyPasswordIndex] = key.FileKeyPassword

			modifiedKey := ds.Keys[key.FileKeyName]
			modifiedKey.DeviceType = key.DeviceType
			modifiedKey.File = currentFileKeyPath
			modifiedKey.Password = currentFileKeyPath
			modifiedKey.Issuer = key.FileKeyIssuer
			modifiedKey.AllowedRegistries = key.AllowedRegistries
			ds.Keys[key.FileKeyName] = modifiedKey
			valuesChanged = true
		} else if key.DeviceType == KeyDeviceTypeHardware {
			if storedKey, keyExists := ds.Keys[key.HardKeyName]; keyExists {
				storedKeyRawPassword := vaultService.GetPropertyFromVault(storedKey.Password, hardKeyPasswordIndex)
				if storedKey.DeviceType == key.DeviceType &&
					storedKey.Type == key.HardKeyType &&
					storedKey.Issuer == key.HardKeyIssuer &&
					reflect.DeepEqual(storedKey.AllowedRegistries, key.AllowedRegistries) &&
					(storedKey.Password == key.HardKeyPassword || storedKeyRawPassword == key.HardKeyPassword) &&
					(storedKey.Device == fmt.Sprintf("%s:%s (%s)", key.HardKeySerialNumber, key.HardKeyPort, key.HardKeyHost) || storedKey.Device == key.HardKeyDevice) {
					// hardKeys are same
					continue
				}
				if !reflect.DeepEqual(ds.Keys[key.HardKeyName].AllowedRegistries, key.AllowedRegistries) {
					valuesChanged = true
					modifiedKey := ds.Keys[key.HardKeyName]
					modifiedKey.AllowedRegistries = key.AllowedRegistries
					ds.Keys[key.HardKeyName] = modifiedKey
					continue
				}
			}

			currentHardKeyPath := secretPath + "-" + key.HardKeyName

			// Check if the inner map exists, if not, initialize it
			if keySecretData[currentHardKeyPath] == nil {
				keySecretData[currentHardKeyPath] = make(map[string]any)
			}
			keySecretData[currentHardKeyPath][hardKeyPasswordIndex] = key.HardKeyPassword

			modifiedKey := ds.Keys[key.HardKeyName]
			modifiedKey.DeviceType = key.DeviceType
			modifiedKey.Type = key.HardKeyType
			modifiedKey.Issuer = key.HardKeyIssuer
			modifiedKey.Device = fmt.Sprintf("%s:%s (%s)", key.HardKeySerialNumber, key.HardKeyPort, key.HardKeyHost)
			modifiedKey.Password = currentHardKeyPath
			modifiedKey.AllowedRegistries = key.AllowedRegistries
			ds.Keys[key.HardKeyName] = modifiedKey
			valuesChanged = true
		}
	}

	return valuesChanged, nil
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
