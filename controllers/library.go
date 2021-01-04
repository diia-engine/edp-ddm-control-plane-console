package controllers

import (
	"ddm-admin-console/console"
	validation2 "ddm-admin-console/controllers/validation"
	"ddm-admin-console/models/command"
	edperror "ddm-admin-console/models/error"
	"ddm-admin-console/models/query"
	"ddm-admin-console/service"
	cbs "ddm-admin-console/service/codebasebranch"
	"ddm-admin-console/util"
	"ddm-admin-console/util/auth"
	"ddm-admin-console/util/consts"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	"go.uber.org/zap"
)

const otherLanguage = "other"

type LibraryController struct {
	beego.Controller
	EDPTenantService service.EDPTenantService
	CodebaseService  service.CodebaseService
	BranchService    cbs.Service
	GitServerService service.GitServerService
	SlaveService     service.SlaveService
	JobProvisioning  service.JobProvisioning

	IntegrationStrategies []string
	BuildTools            []string
	VersioningTypes       []string
	DeploymentScript      []string
	CiTools               []string
}

func (c *LibraryController) GetLibraryListPage() {
	flash := beego.ReadFromRequest(&c.Controller)
	codebases, err := c.CodebaseService.GetCodebasesByCriteria(query.CodebaseCriteria{
		Type: query.Library,
	})
	codebases = addCodebaseInProgressIfAny(codebases, c.GetString(paramWaitingForCodebase))
	if err != nil {
		c.Abort("500")
		return
	}

	if flash.Data["success"] != "" {
		c.Data["Success"] = true
	}
	if flash.Data["error"] != "" {
		c.Data["DeletionError"] = flash.Data["error"]
	}
	c.Data["Codebases"] = codebases
	c.Data["EDPVersion"] = console.EDPVersion
	c.Data["Username"] = c.Ctx.Input.Session("username")
	c.Data["HasRights"] = auth.IsAdmin(c.GetSession("realm_roles").([]string))
	c.Data["Type"] = query.Library
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML()) //nolint
	c.Data["BasePath"] = console.BasePath
	c.Data["DiagramPageEnabled"] = console.DiagramPageEnabled
	c.TplName = tplCodebase
}

func (c *LibraryController) GetCreatePage() {
	flash := beego.ReadFromRequest(&c.Controller)
	if flash.Data["error"] != "" {
		c.Data["Error"] = flash.Data["error"]
	}

	isVcsEnabled, err := c.EDPTenantService.GetVcsIntegrationValue()
	if err != nil {
		c.Abort("500")
		return
	}

	contains := doesIntegrationStrategiesContainImportStrategy(c.IntegrationStrategies)
	if contains {
		log.Info("Import strategy is used.")

		gitServers, err := c.GitServerService.GetServers(query.GitServerCriteria{Available: true})
		if err != nil {
			c.Abort("500")
			return
		}
		log.Debug("Fetched Git Servers", zap.Any("git servers", gitServers))

		c.Data["GitServers"] = gitServers
	}

	s, err := c.SlaveService.GetAllSlaves()
	if err != nil {
		c.Abort("500")
		return
	}

	p, err := c.JobProvisioning.GetAllJobProvisioners(query.JobProvisioningCriteria{Scope: util.GetStringP(scope)})
	if err != nil {
		c.Abort("500")
		return
	}

	c.Data["EDPVersion"] = console.EDPVersion
	c.Data["Username"] = c.Ctx.Input.Session("username")
	c.Data["HasRights"] = auth.IsAdmin(c.GetSession("realm_roles").([]string))
	c.Data["IsVcsEnabled"] = isVcsEnabled
	c.Data["Type"] = query.Library
	c.Data["CodeBaseIntegrationStrategy"] = true
	c.Data["IntegrationStrategies"] = c.IntegrationStrategies
	c.Data["JenkinsSlaves"] = s
	c.Data["BuildTools"] = c.BuildTools
	c.Data["JobProvisioners"] = p
	c.Data["VersioningTypes"] = c.VersioningTypes
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML()) //nolint
	c.Data["BasePath"] = console.BasePath
	c.Data["DiagramPageEnabled"] = console.DiagramPageEnabled
	c.Data["CiTools"] = c.CiTools
	c.TplName = "create_library.html"
}

func (c *LibraryController) Create() {
	flash := beego.NewFlash()
	codebase := c.extractLibraryRequestData()
	errMsg := validateLibraryRequestData(codebase)
	if errMsg != nil {
		log.Error("Failed to validate library request data", zap.String("err", errMsg.Message))
		flash.Error(errMsg.Message)
		flash.Store(&c.Controller)
		c.Redirect(fmt.Sprintf("%s/admin/library/create", console.BasePath), 302)
		return
	}
	logLibraryRequestData(codebase)

	createdObject, err := c.CodebaseService.CreateCodebase(codebase)
	if err != nil {
		c.checkError(err, flash, codebase.Name, codebase.GitURLPath)
		return
	}

	log.Info("Library object is saved into cluster", zap.String("library", createdObject.Name))
	flash.Success("Library object is created.")
	flash.Store(&c.Controller)
	c.Redirect(fmt.Sprintf("%s/admin/library/overview?%s=%s#codebaseSuccessModal", console.BasePath, paramWaitingForCodebase, codebase.Name), 302)
}

func (c *LibraryController) checkError(err error, flash *beego.FlashData, name string, url *string) {
	switch err.(type) {
	case *edperror.CodebaseAlreadyExistsError:
		flash.Error("Library %v already exists.", name)
		flash.Store(&c.Controller)
		c.Redirect(fmt.Sprintf("%s/admin/library/create", console.BasePath), 302)
	case *edperror.CodebaseWithGitURLPathAlreadyExistsError:
		flash.Error("Library %v with %v project path already exists.", name, *url)
		flash.Store(&c.Controller)
		c.Redirect(fmt.Sprintf("%s/admin/library/create", console.BasePath), 302)
	default:
		log.Error("couldn't create codebase", zap.Error(err))
		c.Abort("500")
	}
}

func (c *LibraryController) extractLibraryRequestData() command.CreateCodebase {
	library := command.CreateCodebase{
		Lang:             c.GetString("appLang"),
		BuildTool:        c.GetString("buildTool"),
		Strategy:         strings.ToLower(c.GetString("strategy")),
		Type:             "library",
		DeploymentScript: c.GetString("deploymentScript"),
		Name:             c.GetString("appName"),
		CiTool:           c.GetString("ciTool"),
		DefaultBranch:    c.GetString("defaultBranchName"),
	}

	if js := c.GetString("jenkinsSlave"); len(js) > 0 {
		library.JenkinsSlave = &js
	}

	if jp := c.GetString("jobProvisioning"); len(jp) > 0 {
		library.JobProvisioning = &jp
	}

	if s := c.GetString("jiraServer"); len(s) > 0 {
		library.JiraServer = &s
	}

	if v := c.GetString("commitMessagePattern"); len(v) > 0 {
		library.CommitMessageRegex = &v
	}

	if v := c.GetString("ticketNamePattern"); len(v) > 0 {
		library.TicketNameRegex = &v
	}

	library.Versioning.Type = c.GetString("versioningType")
	startVersioningFrom := c.GetString("startVersioningFrom")
	sp := c.GetString("snapshotStaticField")
	library.Versioning.StartFrom = util.GetVersionOrNil(startVersioningFrom, sp)

	if consts.LanguageJava == library.Lang || otherLanguage == library.Lang {
		framework := c.GetString("framework")
		library.Framework = &framework
	}

	if library.Strategy == consts.ImportStrategy {
		library.GitServer = c.GetString("gitServer")
		gitRepoPath := c.GetString("gitRelativePath")
		library.GitURLPath = &gitRepoPath
	} else {
		library.GitServer = defaultGitServer
	}

	repoURL := c.GetString("gitRepoUrl")
	if repoURL != "" {
		library.Repository = &command.Repository{
			URL: repoURL,
		}

		isRepoPrivate, _ := c.GetBool("isRepoPrivate", false)
		if isRepoPrivate {
			library.Repository.Login = c.GetString("repoLogin")
			library.Repository.Password = c.GetString("repoPassword")
		}
	}

	vcsLogin := c.GetString("vcsLogin")
	vcsPassword := c.GetString("vcsPassword")
	if vcsLogin != "" && vcsPassword != "" {
		library.Vcs = &command.Vcs{
			Login:    vcsLogin,
			Password: vcsPassword,
		}
	}
	library.Username, _ = c.Ctx.Input.Session("username").(string)
	return library
}

func validateLibraryRequestData(library command.CreateCodebase) *validation2.ErrMsg {
	valid := validation.Validation{}

	_, err := valid.Valid(library)

	if library.Strategy == consts.ImportStrategy {
		valid.Match(library.GitURLPath, regexp.MustCompile("^\\/.*$"), "Spec.GitURLPath") //nolint
	}

	if library.Strategy == "clone" && library.Repository != nil {
		_, err = valid.Valid(library.Repository)

		isAvailable := util.IsGitRepoAvailable(library.Repository.URL, library.Repository.Login, library.Repository.Password)

		if !isAvailable {
			err := &validation.Error{Key: "repository", Message: "Repository doesn't exist or invalid login and password."}
			valid.Errors = append(valid.Errors, err)
		}
	}

	if library.Vcs != nil {
		_, err = valid.Valid(library.Vcs)
	}

	if err != nil {
		return &validation2.ErrMsg{
			Message:    "An internal error has occurred on server while validating autotest's form fields.",
			StatusCode: http.StatusInternalServerError}
	}

	if valid.Errors == nil {
		return nil
	}

	return &validation2.ErrMsg{Message: string(validation2.CreateErrorResponseBody(valid)),
		StatusCode: http.StatusBadRequest}
}

func logLibraryRequestData(library command.CreateCodebase) {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Request data to create codebase CR is valid. name=%s, strategy=%s, lang=%s, buildTool=%s",
		library.Name, library.Strategy, library.Lang, library.BuildTool))

	if library.Repository != nil {
		result.WriteString(fmt.Sprintf(", repositoryUrl=%s, repositoryLogin=%s", library.Repository.URL, library.Repository.Login))
	}

	if library.Vcs != nil {
		result.WriteString(fmt.Sprintf(", vcsLogin=%s", library.Vcs.Login))
	}

	log.Info(result.String())
}
