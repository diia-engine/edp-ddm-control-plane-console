package registry

import (
	"context"
	"ddm-admin-console/router"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	goGerrit "github.com/andygrunwald/go-gerrit"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	currentRevision = "current"
)

func (a *App) abandonChange(ctx *gin.Context) (response router.Response, retErr error) {
	changeID := ctx.Param("change")

	if _, _, err := a.Gerrit.GoGerritClient().Changes.AbandonChange(changeID, &goGerrit.AbandonInput{
		Message: fmt.Sprintf("Abandoned by %s [%s]", ctx.GetString(router.UserNameSessionKey),
			ctx.GetString(router.UserEmailSessionKey)),
	}); err != nil {
		return nil, errors.Wrap(err, "unable to abandon change")
	}

	if err := a.updateMRStatus(ctx, changeID, "ABANDONED"); err != nil {
		return nil, errors.Wrap(err, "unable to change MR status")
	}

	return router.MakeHTMLResponse(200, "registry/change-abandoned.html", gin.H{}), nil
}

func (a *App) submitChange(ctx *gin.Context) (response router.Response, retErr error) {
	changeID := ctx.Param("change")

	if _, rsp, err := a.Gerrit.GoGerritClient().Changes.SetReview(changeID, currentRevision, &goGerrit.ReviewInput{
		Message: fmt.Sprintf("Submitted by %s [%s]", ctx.GetString(router.UserNameSessionKey),
			ctx.GetString(router.UserEmailSessionKey)),
		Labels: map[string]string{
			"Code-Review": "2",
			"Verified":    "1",
		},
	}); err != nil {
		if rsp != nil {
			body, _ := ioutil.ReadAll(rsp.Body)
			return nil, errors.Wrapf(err, "unable to review change, error: %s", string(body))
		}

		return nil, errors.Wrap(err, "unable to review change")
	}

	if _, rsp, err := a.Gerrit.GoGerritClient().Changes.SubmitChange(changeID, &goGerrit.SubmitInput{}); err != nil {
		if rsp != nil {
			body, _ := ioutil.ReadAll(rsp.Body)
			return nil, errors.Wrapf(err, "unable to submit change, error: %s", string(body))
		}

		return nil, errors.Wrap(err, "unable to submit change")
	}

	if err := a.updateMRStatus(ctx, changeID, "MERGED"); err != nil {
		return nil, errors.Wrap(err, "unable to change MR status")
	}

	return router.MakeHTMLResponse(200, "registry/change-submitted.html", gin.H{}), nil
}

func (a *App) updateMRStatus(ctx context.Context, changeID, status string) error {
	mr, err := a.Gerrit.GetMergeRequestByChangeID(ctx, changeID)
	if err != nil {
		return errors.Wrap(err, "unable to get MR")
	}

	mr.Status.Value = status

	for i := 0; i < 5; i++ {
		if err := a.Gerrit.UpdateMergeRequestStatus(ctx, mr); err != nil {
			if strings.Contains(err.Error(), "changed") {
				continue
			}

			return errors.Wrap(err, "unable to update MR status")
		}
	}

	return nil
}

func (a *App) viewChange(ctx *gin.Context) (response router.Response, retErr error) {
	changeID := ctx.Param("change")

	changeInfo, _, err := a.Gerrit.GoGerritClient().Changes.GetChangeDetail(changeID, &goGerrit.ChangeOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get gerrit change details")
	}

	changes, err := a.getChangeContents(changeInfo)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get changes")
	}

	return router.MakeHTMLResponse(200, "registry/change.html", gin.H{
		"page":     "registry",
		"changes":  changes,
		"change":   changeInfo,
		"changeID": changeID,
	}), nil
}

func (a *App) getChangeContents(changeInfo *goGerrit.ChangeInfo) (string, error) {
	files, _, err := a.Gerrit.GoGerritClient().Changes.ListFiles(changeInfo.ID, currentRevision, &goGerrit.FilesOptions{})
	if err != nil {
		return "", errors.Wrap(err, "unable to get change files")
	}

	changes := make([]string, 0, len(files)-1)
	for fileName := range files {
		if fileName == "/COMMIT_MSG" {
			continue
		}

		changesContent, err := a.getFileChanges(changeInfo.ID, fileName, changeInfo.Project, changeInfo.Branch)
		if err != nil {
			return "", errors.Wrap(err, "unable to get file changes")
		}

		changes = append(changes, changesContent)
	}

	out := strings.Join(changes, "")
	bts, err := json.Marshal(out)
	if err != nil {
		return "", errors.Wrap(err, "unable to encode changes")
	}

	return string(bts), nil
}

func (a *App) getFileChanges(changeID, fileName, projectName, branchName string) (string, error) {
	commitInfo, _, err := a.Gerrit.GoGerritClient().Changes.GetCommit(changeID, currentRevision, &goGerrit.CommitOptions{})
	if err != nil {
		return "", errors.Wrap(err, "unable to get change commit")
	}
	if len(commitInfo.Parents) == 0 {
		return "", errors.New("no parent commit for change found")
	}

	originalContent, originalHttpRsp, err := a.Gerrit.GoGerritClient().Projects.GetCommitContent(projectName,
		commitInfo.Parents[0].Commit, url.PathEscape(fileName))
	if err != nil && originalHttpRsp != nil && originalHttpRsp.StatusCode != 404 {
		return "", errors.Wrap(err, "unable to get file content")
	}
	if originalHttpRsp == nil {
		return "", errors.New("empty http response")
	}
	originalFilePath := path.Join(a.Config.TempFolder, "original", fileName)

	if originalHttpRsp.StatusCode != 404 {
		originalFolder := filepath.Dir(originalFilePath)
		if err := os.MkdirAll(originalFolder, 0777); err != nil {
			return "", errors.Wrap(err, "unable to create folder")
		}

		originalFile, err := os.Create(originalFilePath)
		if err != nil {
			return "", errors.Wrap(err, "")
		}
		if _, err := originalFile.WriteString(originalContent); err != nil {
			return "", errors.Wrap(err, "")
		}
		if err := originalFile.Close(); err != nil {
			return "", errors.Wrap(err, "")
		}
		defer os.RemoveAll(originalFilePath)
	}

	content, newHttpRsp, err := a.Gerrit.GoGerritClient().Changes.GetContent(changeID, currentRevision, fileName)
	if err != nil && newHttpRsp != nil && newHttpRsp.StatusCode != 404 {
		return "", errors.Wrap(err, "unable to get file content")
	}
	if newHttpRsp == nil {
		return "", errors.New("empty response")
	}
	newFilePath := path.Join(a.Config.TempFolder, "new", fileName)

	if newHttpRsp.StatusCode != 404 && content != nil {
		newFolder := filepath.Dir(newFilePath)
		if err := os.MkdirAll(newFolder, 0777); err != nil {
			return "", errors.Wrap(err, "unable to create folder")
		}

		newFile, err := os.Create(newFilePath)
		if err != nil {
			return "", errors.Wrap(err, "")
		}
		if _, err := newFile.WriteString(*content); err != nil {
			return "", errors.Wrap(err, "")
		}
		if err := newFile.Close(); err != nil {
			return "", errors.Wrap(err, "")
		}
		defer os.RemoveAll(newFilePath)
	}

	return createDiff(fileName, originalFilePath, originalHttpRsp.StatusCode == 404,
		newFilePath, newHttpRsp.StatusCode == 404), nil
}

func createDiff(fileName, originalFilePath string, newFileAdded bool, newFilePath string, fileDeleted bool) string {
	var outDiff string
	if newFileAdded {
		out, _ := exec.Command("diff", "-u", "/dev/null", newFilePath).CombinedOutput()
		//outDiff = string(out) + "new file mode 100644\n"
		outDiff = fmt.Sprintf("diff --git a/%s b/%s\n%s", fileName, fileName, string(out))
		//outDiff = string(out)
	} else if fileDeleted {
		out, _ := exec.Command("diff", "-u", originalFilePath, "/dev/null").CombinedOutput()
		//outDiff = string(out) + "deleted file mode 100644\n"
		//outDiff = string(out)
		outDiff = fmt.Sprintf("diff --git a/%s b/%s\n%s", fileName, fileName, string(out))
	} else {
		out, _ := exec.Command("diff", "-u", originalFilePath, newFilePath).CombinedOutput()
		//outDiff = string(out)
		outDiff = fmt.Sprintf("diff --git a/%s b/%s\n%s", fileName, fileName, string(out))
		//outDiff = string(out)
	}

	outDiff = strings.ReplaceAll(outDiff, newFilePath, fileName)
	outDiff = strings.ReplaceAll(outDiff, originalFilePath, fileName)

	return outDiff
}
