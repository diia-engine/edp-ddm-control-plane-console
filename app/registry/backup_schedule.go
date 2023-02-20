package registry

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	registryBackupIndex    = "registryBackup"
	MRActionBackupSchedule = "backup-schedule"
)

func (a *App) prepareBackupSchedule(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) error {

	if r.BackupScheduleEnabled == "" && values.Global.RegistryBackup.Enabled {
		values.Global.RegistryBackup.Enabled = false
		values.OriginalYaml[registryBackupIndex] = values.Global.RegistryBackup
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

		globalRaw, ok := values.OriginalYaml["global"]
		if !ok {
			globalRaw = make(map[string]interface{})
		}

		globalDict := globalRaw.(map[string]interface{})
		globalDict[registryBackupIndex] = values.Global.RegistryBackup
		values.OriginalYaml["global"] = globalDict

		*mrActions = append(*mrActions, MRActionBackupSchedule)
	}

	return nil
}
