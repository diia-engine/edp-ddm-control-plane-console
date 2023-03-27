package codebase

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-version"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// CodebaseSpec defines the desired state of Codebase
// +k8s:openapi-gen=true
const (
	Create              Strategy       = "create"
	Clone               Strategy       = "clone"
	Default             VersioningType = "default"
	ViewTimeFormat                     = "02.01.2006 15:04"
	DataTableTimeFormat                = "2006-01-02 15:04:05"
	statusActive                       = "active"
	AdminsAnnotation                   = "registry-parameters/administrators"
	RepoNotReady                       = "NOT_READY"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GitServerSpec defines the desired state of GitServer
// +k8s:openapi-gen=true
type GitServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	GitHost                  string `json:"gitHost"`
	GitUser                  string `json:"gitUser"`
	HttpsPort                int32  `json:"httpsPort"`
	SshPort                  int32  `json:"sshPort"`
	NameSshKeySecret         string `json:"nameSshKeySecret"`
	CreateCodeReviewPipeline bool   `json:"createCodeReviewPipeline"`
}

// GitServerStatus defines the observed state of GitServer
// +k8s:openapi-gen=true

type GitServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Available       bool      `json:"available"`
	LastTimeUpdated time.Time `json:"last_time_updated"`
	Status          string    `json:"status"`
	Username        string    `json:"username"`
	Action          string    `json:"action"`
	Result          string    `json:"result"`
	DetailedMessage string    `json:"detailed_message"`
	Value           string    `json:"value"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitServer is the Schema for the gitservers API
// +k8s:openapi-gen=true
type GitServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitServerSpec   `json:"spec,omitempty"`
	Status GitServerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitServerList contains a list of GitServer
type GitServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitServer `json:"items"`
}

type VersioningType string

type Strategy string

type Versioning struct {
	Type      VersioningType `json:"type"`
	StartFrom *string        `json:"startFrom,omitempty"`
}

type Repository struct {
	Url string `json:"url"`
}

type Route struct {
	Site string `json:"site"`
	Path string `json:"path"`
}

type Perf struct {
	Name        string   `json:"name"`
	DataSources []string `json:"dataSources"`
}

type CodebaseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Lang                        string      `json:"lang"`
	Description                 *string     `json:"description"`
	Framework                   *string     `json:"framework"`
	BuildTool                   string      `json:"buildTool"`
	Strategy                    Strategy    `json:"strategy"`
	Repository                  *Repository `json:"repository"`
	Route                       *Route      `json:"route"`
	TestReportFramework         *string     `json:"testReportFramework"`
	Type                        string      `json:"type"`
	GitServer                   string      `json:"gitServer"`
	GitUrlPath                  *string     `json:"gitUrlPath"`
	JenkinsSlave                *string     `json:"jenkinsSlave"`
	JobProvisioning             *string     `json:"jobProvisioning"`
	DeploymentScript            string      `json:"deploymentScript"`
	Versioning                  Versioning  `json:"versioning"`
	JiraServer                  *string     `json:"jiraServer,omitempty"`
	CommitMessagePattern        *string     `json:"commitMessagePattern"`
	TicketNamePattern           *string     `json:"ticketNamePattern"`
	CiTool                      string      `json:"ciTool"`
	Perf                        *Perf       `json:"perf"`
	DefaultBranch               string      `json:"defaultBranch"`
	JiraIssueMetadataPayload    *string     `json:"jiraIssueMetadataPayload"`
	EmptyProject                bool        `json:"emptyProject"`
	BranchToCopyInDefaultBranch string      `json:"branchToCopyInDefaultBranch"`
}

func (in *Codebase) CanBeDeleted() bool {
	for _, cb := range in.Branches {
		if cb.Status.Value != statusActive {
			return false
		}
	}

	return in.Status.Available && in.Status.Value == statusActive
}

func (in *Codebase) ForegroundDeletion() bool {
	return in.DeletionTimestamp != nil
}

func (in *Codebase) FormattedCreatedAtTimezone(timezone string) string {
	loc, _ := time.LoadLocation(timezone)
	return in.CreationTimestamp.In(loc).Format(ViewTimeFormat)
}

func (in *Codebase) CreatedAtTimezone(timezone string) string {
	loc, _ := time.LoadLocation(timezone)
	return in.CreationTimestamp.In(loc).Format(DataTableTimeFormat)
}

func (in *Codebase) StrStatus() string {
	status := in.Status.Value
	if status == "" {
		status = "active"
	}

	return status
}

func (in *Codebase) Available() bool {
	return !in.ForegroundDeletion() && in.StrStatus() != "failed"
}

func (in *Codebase) LocaleStatus() string {
	return fmt.Sprintf("status-%s", in.StrStatus())
}

func (in *CodebaseBranch) LocaleStatus() string {
	return fmt.Sprintf("status-%s", in.StrStatus())
}

func (in *CodebaseBranch) StrStatus() string {
	status := in.Status.Value
	if status == "" {
		status = "active"
	}

	return status
}

func (in *CodebaseBranch) CreateGerritLink(baseURL string) string {
	return fmt.Sprintf("%s/%s", baseURL, "dashboard/self")
}

func (in *CodebaseBranch) CreateJenkinsLink(baseURL string) string {
	return fmt.Sprintf("%v/job/%s/view/%s", baseURL, in.Spec.CodebaseName, strings.ToUpper(in.Spec.BranchName))
}

func (in *Codebase) FormattedCreatedAt() string {
	return in.CreationTimestamp.Format(ViewTimeFormat)
}

func (in *Codebase) Description() string {
	if in.Spec.Description == nil {
		return ""
	}

	return *in.Spec.Description
}

// CodebaseStatus defines the observed state of Codebase
// +k8s:openapi-gen=true
type CodebaseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Available       bool       `json:"available"`
	LastTimeUpdated time.Time  `json:"lastTimeUpdated"`
	Status          string     `json:"status"`
	Username        string     `json:"username"`
	Action          ActionType `json:"action"`
	Result          Result     `json:"result"`
	DetailedMessage string     `json:"detailedMessage"`
	Value           string     `json:"value"`
	FailureCount    int64      `json:"failureCount"`
	Git             string     `json:"git"`
}

type ActionType string
type Result string

const (
	AcceptCodebaseRegistration       ActionType = "accept_codebase_registration"
	GerritRepositoryProvisioning     ActionType = "gerrit_repository_provisioning"
	JenkinsConfiguration             ActionType = "jenkins_configuration"
	SetupDeploymentTemplates         ActionType = "setup_deployment_templates"
	AcceptCodebaseBranchRegistration ActionType = "accept_codebase_branch_registration"
	PutS2I                           ActionType = "put_s2i"
	PutJenkinsFolder                 ActionType = "put_jenkins_folder"
	CleanData                        ActionType = "clean_data"
	ImportProject                    ActionType = "import_project"
	PutVersionFile                   ActionType = "put_version_file"
	PutGitlabCIFile                  ActionType = "put_gitlab_ci_file"
	PutBranchForGitlabCiCodebase     ActionType = "put_branch_for_gitlab_ci_codebase"
	PutCodebaseImageStream           ActionType = "put_codebase_image_stream"
	TriggerReleaseJob                ActionType = "trigger_release_job"
	TriggerDeletionJob               ActionType = "trigger_deletion_job"
	PerfDataSourceCrUpdate           ActionType = "perf_data_source_cr_update"

	Success Result = "success"
	Error   Result = "error"
)

// Codebase is the Schema for the codebases API
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Codebase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec     CodebaseSpec     `json:"spec,omitempty"`
	Status   CodebaseStatus   `json:"status,omitempty"`
	Branches []CodebaseBranch `json:"-"`
	Version  *version.Version `json:"-"`
}

func (in *Codebase) Admins() string {
	if in.Annotations != nil && in.Annotations[AdminsAnnotation] != "" {
		admins, err := base64.StdEncoding.DecodeString(in.Annotations[AdminsAnnotation])
		if err != nil {
			return err.Error()
		}

		return string(admins)
	}

	return ""
}

// CodebaseList contains a list of Codebase
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CodebaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Codebase `json:"items"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CodebaseBranchSpec defines the desired state of CodebaseBranch
// +k8s:openapi-gen=true
type CodebaseBranchSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	CodebaseName string  `json:"codebaseName"`
	BranchName   string  `json:"branchName"`
	FromCommit   string  `json:"fromCommit"`
	Version      *string `json:"version,omitempty"`
	Release      bool    `json:"release"`
}

// CodebaseBranchStatus defines the observed state of CodebaseBranch
// +k8s:openapi-gen=true
type CodebaseBranchStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	LastTimeUpdated     time.Time  `json:"lastTimeUpdated"`
	VersionHistory      []string   `json:"versionHistory"`
	LastSuccessfulBuild *string    `json:"lastSuccessfulBuild,omitempty"`
	Build               *string    `json:"build,omitempty"`
	Status              string     `json:"status"`
	Username            string     `json:"username"`
	Action              ActionType `json:"action"`
	Result              Result     `json:"result"`
	DetailedMessage     string     `json:"detailedMessage"`
	Value               string     `json:"value"`
	FailureCount        int64      `json:"failureCount"`
}

// CodebaseBranch is the Schema for the codebasebranches API
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CodebaseBranch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CodebaseBranchSpec   `json:"spec,omitempty"`
	Status CodebaseBranchStatus `json:"status,omitempty"`
}

// CodebaseBranchList contains a list of CodebaseBranch
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CodebaseBranchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CodebaseBranch `json:"items"`
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Codebase) DeepCopyInto(out *Codebase) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Codebase.
func (in *Codebase) DeepCopy() *Codebase {
	if in == nil {
		return nil
	}
	out := new(Codebase)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Codebase) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodebaseBranch) DeepCopyInto(out *CodebaseBranch) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodebaseBranch.
func (in *CodebaseBranch) DeepCopy() *CodebaseBranch {
	if in == nil {
		return nil
	}
	out := new(CodebaseBranch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CodebaseBranch) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodebaseBranchList) DeepCopyInto(out *CodebaseBranchList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CodebaseBranch, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodebaseBranchList.
func (in *CodebaseBranchList) DeepCopy() *CodebaseBranchList {
	if in == nil {
		return nil
	}
	out := new(CodebaseBranchList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CodebaseBranchList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodebaseBranchSpec) DeepCopyInto(out *CodebaseBranchSpec) {
	*out = *in
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodebaseBranchSpec.
func (in *CodebaseBranchSpec) DeepCopy() *CodebaseBranchSpec {
	if in == nil {
		return nil
	}
	out := new(CodebaseBranchSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodebaseBranchStatus) DeepCopyInto(out *CodebaseBranchStatus) {
	*out = *in
	if in.VersionHistory != nil {
		in, out := &in.VersionHistory, &out.VersionHistory
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.LastSuccessfulBuild != nil {
		in, out := &in.LastSuccessfulBuild, &out.LastSuccessfulBuild
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodebaseBranchStatus.
func (in *CodebaseBranchStatus) DeepCopy() *CodebaseBranchStatus {
	if in == nil {
		return nil
	}
	out := new(CodebaseBranchStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodebaseList) DeepCopyInto(out *CodebaseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Codebase, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodebaseList.
func (in *CodebaseList) DeepCopy() *CodebaseList {
	if in == nil {
		return nil
	}
	out := new(CodebaseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CodebaseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodebaseSpec) DeepCopyInto(out *CodebaseSpec) {
	*out = *in
	if in.Description != nil {
		in, out := &in.Description, &out.Description
		*out = new(string)
		**out = **in
	}
	if in.Framework != nil {
		in, out := &in.Framework, &out.Framework
		*out = new(string)
		**out = **in
	}
	if in.Repository != nil {
		in, out := &in.Repository, &out.Repository
		*out = new(Repository)
		**out = **in
	}
	if in.Route != nil {
		in, out := &in.Route, &out.Route
		*out = new(Route)
		**out = **in
	}
	if in.TestReportFramework != nil {
		in, out := &in.TestReportFramework, &out.TestReportFramework
		*out = new(string)
		**out = **in
	}
	if in.GitUrlPath != nil {
		in, out := &in.GitUrlPath, &out.GitUrlPath
		*out = new(string)
		**out = **in
	}
	in.Versioning.DeepCopyInto(&out.Versioning)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodebaseSpec.
func (in *CodebaseSpec) DeepCopy() *CodebaseSpec {
	if in == nil {
		return nil
	}
	out := new(CodebaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CodebaseStatus) DeepCopyInto(out *CodebaseStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CodebaseStatus.
func (in *CodebaseStatus) DeepCopy() *CodebaseStatus {
	if in == nil {
		return nil
	}
	out := new(CodebaseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Versioning) DeepCopyInto(out *Versioning) {
	*out = *in
	if in.StartFrom != nil {
		in, out := &in.StartFrom, &out.StartFrom
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Versioning.
func (in *Versioning) DeepCopy() *Versioning {
	if in == nil {
		return nil
	}
	out := new(Versioning)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServer) DeepCopyInto(out *GitServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServer.
func (in *GitServer) DeepCopy() *GitServer {
	if in == nil {
		return nil
	}
	out := new(GitServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerList) DeepCopyInto(out *GitServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GitServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerList.
func (in *GitServerList) DeepCopy() *GitServerList {
	if in == nil {
		return nil
	}
	out := new(GitServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerSpec) DeepCopyInto(out *GitServerSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerSpec.
func (in *GitServerSpec) DeepCopy() *GitServerSpec {
	if in == nil {
		return nil
	}
	out := new(GitServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitServerStatus) DeepCopyInto(out *GitServerStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitServerStatus.
func (in *GitServerStatus) DeepCopy() *GitServerStatus {
	if in == nil {
		return nil
	}
	out := new(GitServerStatus)
	in.DeepCopyInto(out)
	return out
}
