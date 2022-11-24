package cluster

import (
	"ddm-admin-console/router"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	backupValuesIndex = "backup"
)

type ScheduleItem struct {
	Schedule      string `yaml:"schedule"`
	ExpiresInDays string `yaml:"expiresInDays"`
}

type BackupScheduleForm struct {
	NexusSchedule               string `form:"nexus-schedule"`
	NexusExpiresInDays          string `form:"nexus-expires-in-days"`
	ControlPlaneSchedule        string `form:"control-plane-schedule"`
	ControlPlaneExpiresInDays   string `form:"control-plane-expires-in-days"`
	UserManagementSchedule      string `form:"user-management-schedule"`
	UserManagementExpiresInDays string `form:"user-management-expires-in-days"`
	MonitoringSchedule          string `form:"monitoring-schedule"`
	MonitoringExpiresInDays     string `form:"monitoring-expires-in-days"`
}

func (bs BackupSchedule) ToForm() BackupScheduleForm {
	return BackupScheduleForm{
		UserManagementExpiresInDays: bs.UserManagement.ExpiresInDays,
		UserManagementSchedule:      bs.UserManagement.Schedule,
		MonitoringExpiresInDays:     bs.Monitoring.ExpiresInDays,
		MonitoringSchedule:          bs.Monitoring.Schedule,
		ControlPlaneExpiresInDays:   bs.ControlPlane.ExpiresInDays,
		ControlPlaneSchedule:        bs.ControlPlane.Schedule,
		NexusSchedule:               bs.Nexus.Schedule,
		NexusExpiresInDays:          bs.Nexus.ExpiresInDays,
	}
}

func (bs BackupScheduleForm) ToNestedStruct() BackupSchedule {
	return BackupSchedule{
		Nexus: ScheduleItem{
			ExpiresInDays: bs.NexusExpiresInDays,
			Schedule:      bs.NexusSchedule,
		},
		ControlPlane: ScheduleItem{
			Schedule:      bs.ControlPlaneSchedule,
			ExpiresInDays: bs.ControlPlaneExpiresInDays,
		},
		Monitoring: ScheduleItem{
			Schedule:      bs.MonitoringSchedule,
			ExpiresInDays: bs.MonitoringExpiresInDays,
		},
		UserManagement: ScheduleItem{
			Schedule:      bs.UserManagementSchedule,
			ExpiresInDays: bs.UserManagementExpiresInDays,
		},
	}
}

type BackupSchedule struct {
	Nexus          ScheduleItem `yaml:"controlPlaneNexus"`
	ControlPlane   ScheduleItem `yaml:"controlPlane"`
	UserManagement ScheduleItem `yaml:"userManagement"`
	Monitoring     ScheduleItem `yaml:"monitoring"`
}

func (a *App) backupSchedule(ctx *gin.Context) (router.Response, error) {
	var bs BackupScheduleForm

	if err := ctx.ShouldBind(&bs); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeHTMLResponse(200, "cluster/edit.html",
			gin.H{"page": "cluster", "errorsMap": validationErrors, "backupSchedule": bs}), nil
	}

	vals, err := a.getValuesDict(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	vals[backupValuesIndex] = bs.ToNestedStruct()
	bts, err := yaml.Marshal(vals)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode yaml")
	}

	if err := a.createValuesMergeRequest(ctx, &valuesMrConfig{
		values:        string(bts),
		authorEmail:   ctx.GetString(router.UserEmailSessionKey),
		authorName:    ctx.GetString(router.UserNameSessionKey),
		commitMessage: "update platform backup schedule config",
		targetLabel:   MRTypeClusterBackupSchedule,
		name:          fmt.Sprintf("backup-schedule-%s-%d", a.Config.CodebaseName, time.Now().Unix()),
	}); err != nil {
		return nil, errors.Wrap(err, "unable to create merge request")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) loadBackupScheduleConfig(valuesDict map[string]interface{}, rspParams gin.H) error {
	var (
		bs  BackupSchedule
		bsf BackupScheduleForm
	)

	backupConfig, ok := valuesDict[backupValuesIndex]
	if !ok {
		rspParams["backupSchedule"] = bsf
		return nil
	}

	bts, err := yaml.Marshal(backupConfig)
	if err != nil {
		return errors.Wrap(err, "unable to encode backup schedule")
	}

	if err := yaml.Unmarshal(bts, &bs); err != nil {
		return errors.Wrap(err, "unable to encode backup schedule")
	}

	rspParams["backupSchedule"] = bs.ToForm()
	return nil
}
