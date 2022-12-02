package cluster

import (
	"ddm-admin-console/router"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

const (
	veleroValuesIndex = "velero"
)

type ScheduleItem struct {
	Schedule      string `yaml:"schedule"`
	ExpiresInDays int    `yaml:"expires_in_days"`
}

type BackupScheduleForm struct {
	NexusSchedule               string `form:"nexus-schedule" binding:"required,cron-expression"`
	NexusExpiresInDays          string `form:"nexus-expires-in-days" binding:"required,only-integer"`
	ControlPlaneSchedule        string `form:"control-plane-schedule" binding:"required,cron-expression"`
	ControlPlaneExpiresInDays   string `form:"control-plane-expires-in-days" binding:"required,only-integer"`
	UserManagementSchedule      string `form:"user-management-schedule" binding:"required,cron-expression"`
	UserManagementExpiresInDays string `form:"user-management-expires-in-days" binding:"required,only-integer"`
	MonitoringSchedule          string `form:"monitoring-schedule" binding:"required,cron-expression"`
	MonitoringExpiresInDays     string `form:"monitoring-expires-in-days" binding:"required,only-integer"`
}

func (bs BackupSchedule) ToForm() BackupScheduleForm {
	return BackupScheduleForm{
		UserManagementExpiresInDays: strconv.Itoa(bs.UserManagement.ExpiresInDays),
		UserManagementSchedule:      bs.UserManagement.Schedule,
		MonitoringExpiresInDays:     strconv.Itoa(bs.Monitoring.ExpiresInDays),
		MonitoringSchedule:          bs.Monitoring.Schedule,
		ControlPlaneExpiresInDays:   strconv.Itoa(bs.ControlPlane.ExpiresInDays),
		ControlPlaneSchedule:        bs.ControlPlane.Schedule,
		NexusExpiresInDays:          strconv.Itoa(bs.Nexus.ExpiresInDays),
		NexusSchedule:               bs.Nexus.Schedule,
	}
}

func mustInt(s string) int {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}

	return int(i)
}

func (bsf BackupScheduleForm) ToNestedStruct() BackupSchedule {
	return BackupSchedule{
		Nexus: ScheduleItem{
			ExpiresInDays: mustInt(bsf.NexusExpiresInDays),
			Schedule:      bsf.NexusSchedule,
		},
		ControlPlane: ScheduleItem{
			Schedule:      bsf.ControlPlaneSchedule,
			ExpiresInDays: mustInt(bsf.ControlPlaneExpiresInDays),
		},
		Monitoring: ScheduleItem{
			Schedule:      bsf.MonitoringSchedule,
			ExpiresInDays: mustInt(bsf.MonitoringExpiresInDays),
		},
		UserManagement: ScheduleItem{
			Schedule:      bsf.UserManagementSchedule,
			ExpiresInDays: mustInt(bsf.UserManagementExpiresInDays),
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

	values, err := a.getValuesDict(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	valuesDict := values.OriginalYaml

	velero, ok := valuesDict[veleroValuesIndex]
	if !ok {
		velero = make(map[string]interface{})
	}
	veleroDict := velero.(map[string]interface{})

	veleroDict["backup"] = bs.ToNestedStruct()
	valuesDict[veleroValuesIndex] = veleroDict

	bts, err := yaml.Marshal(valuesDict)
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

func CronExpressionValidator(fl validator.FieldLevel) bool {
	expression := fl.Field().String()
	if _, err := cron.ParseStandard(expression); err != nil {
		return false
	}

	return true
}

func OnlyIntegerValidator(fl validator.FieldLevel) bool {
	expression := fl.Field().String()
	if _, err := strconv.ParseInt(expression, 10, 24); err != nil {
		return false
	}

	return true
}

func (a *App) loadBackupScheduleConfig(values *Values, rspParams gin.H) error {
	rspParams["backupSchedule"] = values.Velero.Backup.ToForm()

	return nil
}
