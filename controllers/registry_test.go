package controllers

import (
	"bytes"
	edperror "ddm-admin-console/models/error"
	"ddm-admin-console/models/query"
	_ "ddm-admin-console/templatefunction"
	"ddm-admin-console/test"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/astaxie/beego"
)

func TestListRegistry_GetSuccess(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{}

	beego.Router("/list-registry", MakeListRegistry(codebaseService))
	request, _ := http.NewRequest("GET", "/list-registry", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Fatal("list registry return wrong response code")
	}
}

func TestListRegistry_GetFailure(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{
		GetCodebasesByCriteriaK8sError: errors.New("error on codebase list"),
	}

	beego.Router("/list-registry-failure", MakeListRegistry(codebaseService))
	request, _ := http.NewRequest("GET", "/list-registry-failure", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("no error on list registry fatal")
	}
}

func TestCreatRegistry_Get(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{}
	beego.Router("/create-registry-get", MakeCreateRegistry(codebaseService))
	request, _ := http.NewRequest("GET", "/create-registry-get", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on registry create get")
	}
}

func TestCreatRegistry_Post_ValidationError(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{}
	ctrl := MakeCreateRegistry(codebaseService)
	beego.Router("/create-registry-failure", ctrl)
	request, _ := http.NewRequest("POST", "/create-registry-failure", bytes.NewReader([]byte{}))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 422 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on validation error")
	}
}

func TestCreatRegistry_Post_CodebaseExists(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{
		CreateError: edperror.NewCodebaseAlreadyExistsError(),
	}
	ctrl := MakeCreateRegistry(codebaseService)
	beego.Router("/create-registry-k8s-error", ctrl)

	formData := url.Values{
		"name":        []string{"tests"},
		"description": []string{"test"},
	}

	request, _ := http.NewRequest("POST", "/create-registry-k8s-error", strings.NewReader(formData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 422 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on k8s namespace exists")
	}
}

func TestCreatRegistry_Post_ValidationErrorName(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	ctrl := MakeCreateRegistry(test.MockCodebaseService{})
	beego.Router("/create-registry-error-name", ctrl)

	formData := url.Values{
		"name":        []string{"test!s"},
		"description": []string{"test"},
	}

	request, _ := http.NewRequest("POST", "/create-registry-error-name", strings.NewReader(formData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 422 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on name validation")
	}
}

func TestCreatRegistry_Post_Success(t *testing.T) {
	codebaseService := test.MockCodebaseService{}

	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	ctrl := MakeCreateRegistry(codebaseService)
	beego.Router("/create-registry-success", ctrl)

	formData := url.Values{
		"name":        []string{"test"},
		"description": []string{"test"},
	}

	request, _ := http.NewRequest("POST", "/create-registry-success", strings.NewReader(formData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 303 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on namespace creation")
	}
}

func TestEditRegistry_GetFailure(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	codebaseService := test.MockCodebaseService{
		GetCodebaseByNameK8sError: errors.New("k8s fatal error"),
	}
	ctrl := MakeEditRegistry(codebaseService)

	beego.Router("/edit-registry-get-failure/:name", ctrl)
	request, _ := http.NewRequest("GET", "/edit-registry-get-failure/test", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit failure")
	}
}

func TestEditRegistry_PostFailure_k8sFatal(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{
		UpdateDescriptionError: errors.New("k8s fatal"),
	}
	ctrl := MakeEditRegistry(cbMock)

	beego.Router("/edit-registry-failure/:name", ctrl)
	request, _ := http.NewRequest("POST", "/edit-registry-failure/test", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit failure")
	}
}

func TestEditRegistry_PostFailure_LongDescription(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{}
	ctrl := MakeEditRegistry(cbMock)

	formData := url.Values{
		"description": []string{`test11111111111111111111111111111111111111111111111111111111111111111111111test1111111
1111111111111111111111111111111111111111111111111111111111111111test11111111111111111111111111111111111111111111111111
111111111111111111111test11111111111111111111111111111111111111111111111111111111111111111111111test1111111111111111111
1111111111111111111111111111111111111111111111111111test11111111111111111111111111111111111111111111111111111111111111
111111111test11111111111111111111111111111111111111111111111111111111111111111111111test111111111111111111111111111111
11111111111111111111111111111111111111111test11111111111111111111111111111111111111111111111111111111111111111111111t
est11111111111111111111111111111111111111111111111111111111111111111111111test111111111111111111111111111111111111111
11111111111111111111111111111111`},
	}

	beego.Router("/edit-registry-failure-description/:name", ctrl)
	request, _ := http.NewRequest("POST", "/edit-registry-failure-description/test", strings.NewReader(formData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 422 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit failure")
	}
}

func TestEditRegistry_PostSuccess(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{}
	ctrl := MakeEditRegistry(cbMock)

	formData := url.Values{
		"description": []string{"test1"},
	}

	beego.Router("/edit-registry-success-description/:name", ctrl)
	request, _ := http.NewRequest("POST", "/edit-registry-success-description/test", strings.NewReader(formData.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 303 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit success")
	}
}

func TestEditRegistry_GetSuccess(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{
		GetCodebaseByNameResult: &query.Codebase{},
	}
	ctrl := MakeEditRegistry(cbMock)

	beego.Router("/edit-registry-success/:name", ctrl)
	request, _ := http.NewRequest("GET", "/edit-registry-success/test", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Log(responseWriter.Body.String())
		t.Fatal("wrong response code on registry edit")
	}
}

func TestListRegistry_DeleteRegistry_FailureGetCodebase(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	mockErr := errors.New("GetCodebaseByNameError fatal")
	cbMock := test.MockCodebaseService{
		GetCodebaseByNameK8sError: mockErr,
	}
	listRegistryCtrl := MakeListRegistry(cbMock)

	beego.Router("/delete-registry-FailureGetCodebase", listRegistryCtrl)
	request, _ := http.NewRequest("POST", "/delete-registry-FailureGetCodebase", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on delete registry")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Fatal("no error in response body")
	}
}

func TestListRegistry_DeleteRegistry_FailureDeleteCodebase(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	mockErr := errors.New("DeleteCodebase fatal")
	cbMock := test.MockCodebaseService{
		GetCodebaseByNameK8sResult: &query.Codebase{},
		DeleteError:                mockErr,
	}
	listRegistryCtrl := MakeListRegistry(cbMock)

	beego.Router("/delete-registry-DeleteCodebase", listRegistryCtrl)
	request, _ := http.NewRequest("POST", "/delete-registry-DeleteCodebase", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code on delete registry")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Fatal("no error in response body")
	}
}

func TestListRegistry_DeleteRegistry(t *testing.T) {
	rw, ctrl := initBeegoCtrl()
	cbMock := test.MockCodebaseService{
		GetCodebaseByNameK8sResult: &query.Codebase{},
	}
	listRegistryCtrl := MakeListRegistry(cbMock)
	ctrl.Ctx.Input.SetParam("registry-name", "test")
	listRegistryCtrl.Controller = ctrl

	listRegistryCtrl.Post()

	if rw.Code != 303 {
		t.Log(rw.Code)
		t.Fatal("wrong response code on delete registry")
	}
}

func TestViewRegistry_Get(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}

	cbMock := test.MockCodebaseService{
		GetCodebaseByNameResult: &query.Codebase{
			CodebaseBranch: []*query.CodebaseBranch{
				{},
			},
		},
	}
	eds := test.MockEDPComponentService{
		GetEDPComponentResult: &query.EDPComponent{},
	}

	beego.Router("/view-registry", MakeViewRegistry(cbMock, eds))
	request, _ := http.NewRequest("GET", "/view-registry", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 200 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code")
	}
}

func TestViewRegistry_Get_FailureGetCodebaseByName(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}
	mockErr := errors.New("GetCodebaseByName fatal")

	cbMock := test.MockCodebaseService{
		GetCodebaseByNameError: mockErr,
	}
	eds := test.MockEDPComponentService{}

	beego.Router("/view-registry-FailureGetCodebaseByName", MakeViewRegistry(cbMock, eds))
	request, _ := http.NewRequest("GET", "/view-registry-FailureGetCodebaseByName", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Fatal("wrong error return in response body")
	}
}

func TestViewRegistry_Get_FailureCreateLinksForGerritProvider(t *testing.T) {
	if err := test.InitBeego(); err != nil {
		t.Fatal(err)
	}
	mockErr := errors.New("GetEDPComponentError fatal")

	cbMock := test.MockCodebaseService{
		GetCodebaseByNameResult: &query.Codebase{
			CodebaseBranch: []*query.CodebaseBranch{
				{},
			},
		},
	}
	eds := test.MockEDPComponentService{
		GetEDPComponentError: mockErr,
	}

	beego.Router("/view-registry-FailureCreateLinksForGerritProvider", MakeViewRegistry(cbMock, eds))
	request, _ := http.NewRequest("GET", "/view-registry-FailureCreateLinksForGerritProvider", nil)
	responseWriter := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(responseWriter, request)

	if responseWriter.Code != 500 {
		t.Log(responseWriter.Code)
		t.Fatal("wrong response code")
	}

	if !strings.Contains(responseWriter.Body.String(), mockErr.Error()) {
		t.Fatal("wrong error return in response body")
	}
}
