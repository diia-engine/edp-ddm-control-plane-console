package cluster

import (
	"context"
	"ddm-admin-console/app/registry"
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
	Schedule      string `yaml:"schedule" json:"schedule"`
	ExpiresInDays int    `yaml:"expires_in_days" json:"expiresInDays"`
}

type BackupScheduleForm struct {
	NexusSchedule               string `form:"nexus-schedule" binding:"required,cron-expression"`
	NexusExpiresInDays          string `form:"nexus-expires-in-days" binding:"required,cron-expires"`
	ControlPlaneSchedule        string `form:"control-plane-schedule" binding:"required,cron-expression"`
	ControlPlaneExpiresInDays   string `form:"control-plane-expires-in-days" binding:"required,cron-expires"`
	UserManagementSchedule      string `form:"user-management-schedule" binding:"required,cron-expression"`
	UserManagementExpiresInDays string `form:"user-management-expires-in-days" binding:"required,cron-expires"`
	MonitoringSchedule          string `form:"monitoring-schedule" binding:"required,cron-expression"`
	MonitoringExpiresInDays     string `form:"monitoring-expires-in-days" binding:"required,cron-expires"`
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
	Nexus          ScheduleItem `yaml:"controlPlaneNexus" json:"nexus"`
	ControlPlane   ScheduleItem `yaml:"controlPlane" json:"controlPlane"`
	UserManagement ScheduleItem `yaml:"userManagement" json:"userManagement"`
	Monitoring     ScheduleItem `yaml:"monitoring" json:"monitoring"`
}

func (a *App) backupSchedule(ctx *gin.Context) (router.Response, error) {
	var bs BackupScheduleForm

	if err := ctx.ShouldBind(&bs); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		errorsMap := make(map[string][]interface{})
		for _, v := range validationErrors {
			errorsMap[v.Field()] = append(errorsMap[v.Field()], v.Tag())
		}

		return router.MakeJSONResponse(http.StatusUnprocessableEntity, gin.H{"errors": errorsMap}), nil
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
		targetLabel:   MRTargetClusterBackupSchedule,
		name:          fmt.Sprintf("backup-schedule-%s-%d", a.Config.CodebaseName, time.Now().Unix()),
	}); err != nil {
		return nil, errors.Wrap(err, "unable to create merge request")
	}

	return router.MakeJSONResponse(http.StatusOK, gin.H{"errors": nil}), nil
}

func CronExpressionValidator(fl validator.FieldLevel) bool {
	expression := fl.Field().String()
	if _, err := cron.ParseStandard(expression); err != nil {
		return false
	}

	return true
}

func CronDaysValidator(fl validator.FieldLevel) bool {
	expression := fl.Field().String()
	if _, err := strconv.ParseInt(expression, 10, 64); err != nil {
		return false
	}

	if expression == "0" {
		return false
	}

	return true
}

func (a *App) loadKeycloakDefaultHostname(ctx context.Context, values *Values, rspParams gin.H) error {
	hostname, err := registry.LoadKeycloakDefaultHostname(ctx, a.KeycloakDefaultHostname, a.EDPComponent)
	if err != nil {
		return fmt.Errorf("unable to load keycloak hostname, %w", err)
	}

	rspParams["keycloakHostname"] = hostname
	return nil
}

func (a *App) loadBackupScheduleConfig(_ context.Context, values *Values, rspParams gin.H) error {
	rspParams["backupSchedule"] = values.Velero.Backup.ToForm()

	return nil
}
