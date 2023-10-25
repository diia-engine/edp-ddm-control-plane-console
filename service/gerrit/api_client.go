package gerrit

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	goGerrit "github.com/andygrunwald/go-gerrit"
	"github.com/pkg/errors"
	"gopkg.in/resty.v1"
	coreV1Api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (s *Service) initRestyClient() error {
	var secret coreV1Api.Secret
	if err := s.k8sClient.Get(
		context.Background(),
		types.NamespacedName{
			Namespace: s.Namespace,
			Name:      fmt.Sprintf("%s-admin-password", s.RootGerritName),
		},
		&secret,
	); err != nil {
		return errors.Wrap(err, "unable to get admin secret")
	}

	gerritURL := strings.ReplaceAll(s.GerritAPIUrlTemplate, "{HOST}",
		fmt.Sprintf("%s.%s", s.RootGerritName, s.Namespace))

	s.apiClient = resty.New().
		SetHostURL(gerritURL).
		SetBasicAuth(string(secret.Data["user"]), string(secret.Data["password"])).
		SetDisableWarn(true)

	var err error
	s.goGerritClient, err = goGerrit.NewClient(strings.ReplaceAll(gerritURL, "/a/", "/"), s.goGerritHTTPClient)
	if err != nil {
		return errors.Wrap(err, "unable to init gerrit go client")
	}
	s.goGerritClient.Authentication.SetBasicAuth(string(secret.Data["user"]), string(secret.Data["password"]))

	return nil
}

func (s *Service) GoGerritClient() *goGerrit.Client {
	return s.goGerritClient
}

func checkErr(rsp *resty.Response, err error) error {
	if err != nil {
		return errors.Wrap(err, "error during request")
	}

	if rsp.StatusCode() >= 300 {
		return errors.Errorf("wrong response code: %d, content: %s", rsp.StatusCode(), rsp.String())
	}

	return nil
}

// GetFileFromBranch gets the content of a file from the HEAD revision of a certain branch.
// The content is returned as base64 encoded string.
func (s *Service) GetFileFromBranch(projectName, branch, fileLocation string) (string, error) {
	content, _, err := s.GoGerritClient().Projects.GetBranchContent(projectName, branch, fileLocation)
	if err != nil {
		return "", fmt.Errorf("unable to get branch content: %w", err)
	}

	return content, nil
}

func (s *Service) GetFileContents(ctx context.Context, projectName, branch, filePath string) (string, error) {
	filePath = url.PathEscape(filePath)
	path := fmt.Sprintf("projects/%s/branches/%s/files/%s/content", projectName, branch, filePath)
	rsp, err := s.apiClient.NewRequest().SetContext(ctx).
		Get(path)
	if err := checkErr(rsp, err); err != nil {
		return "", errors.Wrap(err, "unable to get file content")
	}

	bts, err := base64.StdEncoding.DecodeString(rsp.String())
	if err != nil {
		return "", errors.Wrap(err, "unable to decode response")
	}

	return string(bts), nil
}

func (s *Service) GetMRFiles(ctx context.Context, changeID string) ([]string, error) {
	var rawRsp map[string]interface{}
	rsp, err := s.apiClient.NewRequest().SetContext(ctx).SetBody(&rawRsp).
		Get(fmt.Sprintf("/changes/%s/revisions/current/files", changeID))

	if err := checkErr(rsp, err); err != nil {
		return nil, errors.Wrap(err, "unable to get mr files")
	}

	files := make([]string, len(rawRsp))
	for k := range rawRsp {
		if k == "/COMMIT_MSG" {
			continue
		}

		files = append(files, k)
	}

	return files, nil
}

type Commit struct {
	Commit  string `json:"commit"`
	Subject string `json:"subject"`
}

func (s *Service) GetMergeListCommits(ctx context.Context, changeID, revision string) ([]Commit, error) {
	rq, _ := s.GoGerritClient().NewRequest("GET",
		fmt.Sprintf("changes/%s/revisions/%s/mergelist", changeID, revision), nil)
	rq = rq.WithContext(ctx)

	var commits []Commit

	_, err := s.GoGerritClient().Do(rq, &commits)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get merge list commits")
	}

	return commits, nil
}
