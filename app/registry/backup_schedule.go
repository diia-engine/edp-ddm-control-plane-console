package registry

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	registryBackupIndex = "registryBackup"
)

func (a *App) prepareBackupSchedule(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}) error {

	if r.BackupScheduleEnabled == "" && values.RegistryBackup.Enabled {
		values.RegistryBackup.Enabled = false
		values.OriginalYaml[registryBackupIndex] = values.RegistryBackup
		return nil
	}

	days, err := strconv.ParseInt(r.CronScheduleDays, 10, 64)
	if err != nil {
		return fmt.Errorf("wrong backup days: %w", err)
	}

	values.RegistryBackup.ExpiresInDays = int(days)
	values.RegistryBackup.Schedule = r.CronSchedule

	values.OriginalYaml[registryBackupIndex] = values.RegistryBackup

	return nil
}
