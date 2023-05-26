package registry

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	registryBackupIndex    = "registryBackup"
	MRActionBackupSchedule = "backup-schedule"
	globalIndex            = "global"
)

func (a *App) prepareBackupSchedule(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) (bool, error) {
	globalRaw, ok := values.OriginalYaml[globalIndex]
	if !ok {
		globalRaw = make(map[string]interface{})
	}
	globalDict := globalRaw.(map[string]interface{})

	if r.BackupScheduleEnabled == "" && values.Global.RegistryBackup.Enabled {
		values.Global.RegistryBackup.Enabled = false
		globalDict[registryBackupIndex] = values.Global.RegistryBackup
		values.OriginalYaml[globalIndex] = globalDict

		*mrActions = append(*mrActions, MRActionBackupSchedule)
		return true, nil
	}

	if r.BackupScheduleEnabled != "" {
		days, err := strconv.ParseInt(r.CronScheduleDays, 10, 64)
		if err != nil {
			return false, fmt.Errorf("wrong backup days: %w", err)
		}

		valuesChanged, err := a.registryBackupValuesChanged(values, r, int(days))
		if err != nil {
			return false, fmt.Errorf("unable to check if backup values changed, %w", err)
		}

		if !valuesChanged {
			return false, nil
		}

		values.Global.RegistryBackup.ExpiresInDays = int(days)
		values.Global.RegistryBackup.Schedule = r.CronSchedule
		values.Global.RegistryBackup.Enabled = true

		values.Global.RegistryBackup.OBC.CronExpression = r.OBCCronExpression
		values.Global.RegistryBackup.OBC.BackupBucket = r.OBCBackupBucket
		values.Global.RegistryBackup.OBC.Endpoint = r.OBCEndpoint

		if r.OBCLogin != "" && r.OBCPassword != "" {
			vaultPath := a.vaultRegistryPathKey(r.Name, fmt.Sprintf("%s-%s", "buckets-backup",
				time.Now().Format("20060201T150405Z")))
			values.Global.RegistryBackup.OBC.Credentials = vaultPath

			secrets[vaultPath] = map[string]interface{}{
				a.Config.BackupBucketAccessKeyID:     r.OBCLogin,
				a.Config.BackupBucketSecretAccessKey: r.OBCPassword,
			}
		}

		globalDict[registryBackupIndex] = values.Global.RegistryBackup
		values.OriginalYaml[globalIndex] = globalDict

		*mrActions = append(*mrActions, MRActionBackupSchedule)

		return true, nil
	}

	return false, nil
}

func (a *App) registryBackupValuesChanged(values *Values, r *registry, cronScheduleDays int) (bool, error) {
	if values.Global.RegistryBackup.ExpiresInDays != cronScheduleDays ||
		values.Global.RegistryBackup.Schedule != r.CronSchedule ||
		values.Global.RegistryBackup.OBC.CronExpression != r.OBCCronExpression ||
		values.Global.RegistryBackup.OBC.BackupBucket != r.OBCBackupBucket ||
		values.Global.RegistryBackup.OBC.Endpoint != r.OBCEndpoint {
		return true, nil
	}

	if values.Global.RegistryBackup.OBC.Credentials == "" && r.OBCLogin != "" {
		return true, nil
	}

	if values.Global.RegistryBackup.OBC.Credentials == "" {
		return false, nil
	}

	secretData, err := a.Vault.Read(values.Global.RegistryBackup.OBC.Credentials)
	if err != nil {
		return false, fmt.Errorf("unable to read secret, %w", err)
	}

	OBCLogin, ok := secretData[a.Config.BackupBucketAccessKeyID]
	if !ok {
		return true, nil
	}

	OBCPassword, ok := secretData[a.Config.BackupBucketSecretAccessKey]
	if !ok {
		return true, nil
	}

	if OBCLogin.(string) != r.OBCLogin || OBCPassword.(string) != r.OBCPassword {
		return true, nil
	}

	return false, nil
}
