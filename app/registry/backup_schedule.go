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
	secrets map[string]map[string]interface{}, mrActions *[]string) error {
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
		return nil
	}

	if r.BackupScheduleEnabled != "" {
		days, err := strconv.ParseInt(r.CronScheduleDays, 10, 64)
		if err != nil {
			return fmt.Errorf("wrong backup days: %w", err)
		}

		values.Global.RegistryBackup.ExpiresInDays = int(days)
		values.Global.RegistryBackup.Schedule = r.CronSchedule
		values.Global.RegistryBackup.Enabled = true

		values.Global.RegistryBackup.OBC.CronExpression = r.OBCCronExpression
		values.Global.RegistryBackup.OBC.BackupBucket = r.OBCBackupBucket
		values.Global.RegistryBackup.OBC.Endpoint = r.OBCEndpoint

		vaultPath := a.vaultRegistryPathKey(r.Name, fmt.Sprintf("%s-%s", "buckets-backup", time.Now().Format("20060201T150405Z")))
		values.Global.RegistryBackup.OBC.Credentials = vaultPath
		vaultPath = ModifyVaultPath(vaultPath)
		secretData := map[string]interface{}{
			a.Config.BackupBucketAccessKeyID:     r.OBCLogin,
			a.Config.BackupBucketSecretAccessKey: r.OBCPassword,
		}

		if _, err := a.Services.Vault.Write(
			vaultPath, map[string]interface{}{
				"data": secretData,
			}); err != nil {
			return fmt.Errorf("unable to write to vault: %w", err)
		}

		globalDict[registryBackupIndex] = values.Global.RegistryBackup
		values.OriginalYaml[globalIndex] = globalDict

		*mrActions = append(*mrActions, MRActionBackupSchedule)
	}

	return nil
}
