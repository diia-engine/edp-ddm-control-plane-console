package validation

import (
	"ddm-admin-console/models/command"
	"ddm-admin-console/models/query"
	"ddm-admin-console/util"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/astaxie/beego/validation"
	gateV1alpha1 "github.com/epmd-edp/cd-pipeline-operator/v2/pkg/apis/edp/v1alpha1"
)

const codebaseAutotests = "autotests"

type ErrMsg struct {
	Message    string
	StatusCode int
}

func ValidCodebaseRequestData(codebase command.CreateCodebase) *ErrMsg {
	valid := validation.Validation{}
	var resErr error

	_, err := valid.Valid(codebase)
	resErr = err

	if codebase.Strategy == "import" {
		valid.Match(codebase.GitURLPath, regexp.MustCompile("^\\/.*$"), "Spec.GitURLPath") //nolint
	}

	if codebase.Repository != nil {
		_, err := valid.Valid(codebase.Repository)

		isAvailable := util.IsGitRepoAvailable(codebase.Repository.URL, codebase.Repository.Login, codebase.Repository.Password)

		if !isAvailable {
			err := &validation.Error{Key: "repository", Message: "Repository doesn't exist or invalid login and password."}
			valid.Errors = append(valid.Errors, err)
		}

		resErr = err
	}

	if codebase.Route != nil {
		if len(codebase.Route.Path) > 0 {
			_, err := valid.Valid(codebase.Route)
			resErr = err
		} else {
			valid.Match(codebase.Route.Site, regexp.MustCompile("^$|^[a-z][a-z0-9-]*[a-z0-9]$"), "Route.Site.Match")
		}
	}

	if codebase.Vcs != nil {
		_, err := valid.Valid(codebase.Vcs)
		resErr = err
	}

	if codebase.Database != nil {
		_, err := valid.Valid(codebase.Database)
		resErr = err
	}

	if !IsCodebaseTypeAcceptable(codebase.Type) {
		err := &validation.Error{Key: "repository", Message: "codebase type should be: application, autotests  or library"}
		valid.Errors = append(valid.Errors, err)
	}

	if codebase.Type == codebaseAutotests && codebase.Strategy != "clone" {
		err := &validation.Error{Key: "repository", Message: "strategy for autotests must be 'clone'"}
		valid.Errors = append(valid.Errors, err)
	}

	if codebase.Type == codebaseAutotests && codebase.Repository == nil {
		err := &validation.Error{Key: "repository", Message: "repository for autotests can't be null"}
		valid.Errors = append(valid.Errors, err)
	}

	if resErr != nil {
		return &ErrMsg{"An internal error has occurred on server while validating application's form fields.", http.StatusInternalServerError}
	}

	if valid.Errors == nil {
		return nil
	}

	return &ErrMsg{string(CreateErrorResponseBody(valid)), http.StatusBadRequest}
}

func CreateCodebaseLogRequestData(app command.CreateCodebase) strings.Builder {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Request data to create CR is valid. name=%v, strategy=%v, lang=%v, buildTool=%v, multiModule=%v, framework=%v",
		app.Name, app.Strategy, app.Lang, app.BuildTool, app.MultiModule, app.Framework))

	if app.Repository != nil {
		result.WriteString(fmt.Sprintf(", repositoryUrl=%s, repositoryLogin=%s", app.Repository.URL, app.Repository.Login))
	}

	if app.Vcs != nil {
		result.WriteString(fmt.Sprintf(", vcsLogin=%s", app.Vcs.Login))
	}

	if app.Route != nil {
		result.WriteString(fmt.Sprintf(", routeSite=%s, routePath=%s", app.Route.Site, app.Route.Path))
	}

	if app.Database != nil {
		result.WriteString(fmt.Sprintf(", dbKind=%s, dbМersion=%s, dbCapacity=%s, dbStorage=%s", app.Database.Kind, app.Database.Version, app.Database.Capacity, app.Database.Storage))
	}
	return result
}

func validateQualityGates(valid validation.Validation, qualityGates []gateV1alpha1.QualityGate) (bool, error) {
	isQualityGatesValid := true

	if qualityGates != nil {
		for _, qualityGate := range qualityGates {
			isValid, err := valid.Valid(qualityGate)
			if err != nil {
				return false, err
			}
			isQualityGatesValid = isValid

			if (qualityGate.QualityGateType == codebaseAutotests && (qualityGate.AutotestName == nil || qualityGate.BranchName == nil)) ||
				(qualityGate.QualityGateType == "manual" && (qualityGate.AutotestName != nil || qualityGate.BranchName != nil)) {
				isQualityGatesValid = false
			}
		}
	} else {
		valid.Errors = append(valid.Errors, &validation.Error{Key: "qualityGates", Message: "can not be empty"})
		isQualityGatesValid = false
	}

	return isQualityGatesValid, nil
}

func CreateErrorResponseBody(valid validation.Validation) []byte {
	errJSON, _ := json.Marshal(extractErrors(valid))
	errResponse := struct {
		Message string
		Content string
	}{
		"Body of request are not valid.",
		string(errJSON),
	}
	response, _ := json.Marshal(errResponse)
	return response
}

func extractErrors(valid validation.Validation) []string {
	var errMap []string
	for _, err := range valid.Errors {
		errMap = append(errMap, fmt.Sprintf("Validation failed on %s: %s", err.Key, err.Message))
	}
	return errMap
}

func IsCodebaseTypeAcceptable(getParam string) bool {
	if _, ok := query.CodebaseTypes[getParam]; ok {
		return true
	}
	return false
}

func ValidateCodebaseUpdateRequestData(c command.UpdateCodebaseCommand) *ErrMsg {
	v := validation.Validation{}
	_, err := v.Valid(c)
	if err != nil {
		return &ErrMsg{"an error has occurred while validating Codebase update request body.",
			http.StatusInternalServerError}
	}

	if v.Errors == nil {
		return nil
	}

	return &ErrMsg{string(CreateErrorResponseBody(v)), http.StatusBadRequest}
}
