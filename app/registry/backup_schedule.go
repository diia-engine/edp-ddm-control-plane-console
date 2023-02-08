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

	if r.BackupScheduleEnabled == "" && values.RegistryBackup.Enabled {
		values.RegistryBackup.Enabled = false
		values.OriginalYaml[registryBackupIndex] = values.RegistryBackup
		return nil
	}

	if r.BackupScheduleEnabled != "" {
		days, err := strconv.ParseInt(r.CronScheduleDays, 10, 64)
		if err != nil {
			return fmt.Errorf("wrong backup days: %w", err)
		}

		values.RegistryBackup.ExpiresInDays = int(days)
		values.RegistryBackup.Schedule = r.CronSchedule
		values.RegistryBackup.Enabled = true

		values.OriginalYaml[registryBackupIndex] = values.RegistryBackup
		*mrActions = append(*mrActions, MRActionBackupSchedule)
	}

	return nil
}
