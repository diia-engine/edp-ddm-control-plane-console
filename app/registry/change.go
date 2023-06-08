package registry

import (
	"context"
	"ddm-admin-console/router"
	"ddm-admin-console/service/gerrit"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	goGerrit "github.com/andygrunwald/go-gerrit"
	"github.com/gin-gonic/gin"
)

const (
	currentRevision    = "current"
	mergeList          = "/MERGE_LIST"
	updateMRRetryCount = 5
)

func (a *App) abandonChange(ctx *gin.Context) (response router.Response, retErr error) {
	changeID := ctx.Param("change")

	if _, _, err := a.Gerrit.GoGerritClient().Changes.AbandonChange(changeID, &goGerrit.AbandonInput{
		Message: fmt.Sprintf("Abandoned by %s [%s]", ctx.GetString(router.UserNameSessionKey),
			ctx.GetString(router.UserEmailSessionKey)),
	}); err != nil {
		return nil, fmt.Errorf("unable to abandon change, %w", err)
	}

	mr, err := a.updateMRStatus(ctx, changeID, "ABANDONED")
	if err != nil {
		return nil, fmt.Errorf("unable to change MR status, %w", err)
	}

	if err := ClearRepoFiles(mr.Spec.ProjectName, a.Cache); err != nil {
		return nil, fmt.Errorf("unable to clear cached files")
	}

	return router.MakeHTMLResponse(200, "registry/change-abandoned.html", gin.H{}), nil
}

func (a *App) submitChange(ctx *gin.Context) (response router.Response, retErr error) {
	changeID := ctx.Param("change")

	if err := a.Gerrit.ApproveAndSubmitChange(changeID, ctx.GetString(router.UserNameSessionKey),
		ctx.GetString(router.UserEmailSessionKey)); err != nil {
		return nil, fmt.Errorf("unable to approve change, %w", err)
	}

	if _, err := a.updateMRStatus(ctx, changeID, "MERGED"); err != nil {
		return nil, fmt.Errorf("unable to change MR status, %w", err)
	}

	return router.MakeHTMLResponse(200, "registry/change-submitted.html", gin.H{}), nil
}

func (a *App) updateMRStatus(ctx context.Context, changeID, status string) (*gerrit.GerritMergeRequest, error) {
	var mr *gerrit.GerritMergeRequest

	for i := 0; i < updateMRRetryCount; i++ {
		var err error
		mr, err = a.Gerrit.GetMergeRequestByChangeID(ctx, changeID)
		if err != nil {
			return nil, fmt.Errorf("unable to get MR, %w", err)
		}

		mr.Status.Value = status

		if err := a.Gerrit.UpdateMergeRequestStatus(ctx, mr); err != nil {
			if strings.Contains(err.Error(), "modified") {
				continue
			}

			return nil, fmt.Errorf("unable to update MR status, %w", err)
		}

		break
	}

	return mr, nil
}

func (a *App) viewChange(ctx *gin.Context) (response router.Response, retErr error) {
	changeID := ctx.Param("change")

	changeInfo, _, err := a.Gerrit.GoGerritClient().Changes.GetChangeDetail(changeID, &goGerrit.ChangeOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get gerrit change details, %w", err)
	}

	changes, err := a.getChangeContents(ctx, changeInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to get changes, %w", err)
	}

	rspParams := gin.H{
		"changes":  changes,
		"change":   changeInfo,
		"changeID": changeID,
	}

	templateArgs, err := json.Marshal(rspParams)
	if err != nil {
		return nil, errors.New("unable to encode template arguments")
	}

	return router.MakeHTMLResponse(200, "registry/change.html", gin.H{
		"page":         "registry",
		"templateArgs": string(templateArgs),
	}), nil
}

func (a *App) getChangeContents(ctx context.Context, changeInfo *goGerrit.ChangeInfo) (string, error) {
	files, _, err := a.Gerrit.GoGerritClient().Changes.ListFiles(changeInfo.ID, currentRevision, &goGerrit.FilesOptions{
		Parent: 1,
	})
	if err != nil {
		return "", fmt.Errorf("unable to get change files, %w", err)
	}

	changes := make([]string, 0, len(files)-1)
	for fileName := range files {
		if fileName == "/COMMIT_MSG" || fileName == mergeList {
			continue
		}

		changesContent, err := a.getChangeFileChanges(changeInfo.ID, fileName, changeInfo.Project)
		if err != nil {
			return "", fmt.Errorf("unable to get file changes, %w", err)
		}

		changes = append(changes, changesContent)
	}

	out := strings.Join(changes, "")
	bts, err := json.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("unable to encode changes, %w", err)
	}

	return string(bts), nil
}

func (a *App) getChangeContentData(changeID, fileName, projectName string) (string, error) {
	content, newHttpRsp, err := a.Gerrit.GoGerritClient().Changes.GetContent(changeID, currentRevision, fileName)
	if err != nil && newHttpRsp != nil && newHttpRsp.StatusCode != 404 {
		return "", errors.New("unable to get file content")
	}
	if newHttpRsp == nil || content == nil {
		return "", errors.New("empty response")
	}

	return *content, nil
}

// TODO: split function
func (a *App) getChangeFileChanges(changeID, fileName, projectName string) (string, error) {
	commitInfo, _, err := a.Gerrit.GoGerritClient().Changes.GetCommit(changeID, currentRevision, &goGerrit.CommitOptions{})
	if err != nil {
		return "", fmt.Errorf("unable to get change commit, %w", err)
	}
	if len(commitInfo.Parents) == 0 {
		return "", errors.New("no parent commit for change found")
	}

	originalContent, originalHttpRsp, err := a.Gerrit.GoGerritClient().Projects.GetCommitContent(projectName,
		commitInfo.Parents[0].Commit, url.PathEscape(fileName))
	if err != nil && originalHttpRsp != nil && originalHttpRsp.StatusCode != 404 {
		return "", fmt.Errorf("unable to get file content, %w", err)
	}
	if originalHttpRsp == nil {
		return "", errors.New("empty http response")
	}
	originalFilePath := path.Join(a.Config.TempFolder, "original", fileName)

	if originalHttpRsp.StatusCode != 404 {
		originalFolder := filepath.Dir(originalFilePath)
		if err := os.MkdirAll(originalFolder, 0777); err != nil {
			return "", fmt.Errorf("unable to create folder, %w", err)
		}

		originalFile, err := os.Create(originalFilePath)
		if err != nil {
			return "", fmt.Errorf("unable to creaete file, %w", err)
		}
		if _, err := originalFile.WriteString(originalContent); err != nil {
			return "", fmt.Errorf("unable to write string, %w", err)
		}
		if err := originalFile.Close(); err != nil {
			return "", fmt.Errorf("unable to close file, %w", err)
		}
		defer os.RemoveAll(originalFilePath)
	}

	content, newHttpRsp, err := a.Gerrit.GoGerritClient().Changes.GetContent(changeID, currentRevision, fileName)
	if err != nil && newHttpRsp != nil && newHttpRsp.StatusCode != 404 {
		return "", fmt.Errorf("unable to get file content, %w", err)
	}
	if newHttpRsp == nil {
		return "", errors.New("empty response")
	}
	newFilePath := path.Join(a.Config.TempFolder, "new", fileName)

	if newHttpRsp.StatusCode != 404 && content != nil {
		newFolder := filepath.Dir(newFilePath)
		if err := os.MkdirAll(newFolder, 0777); err != nil {
			return "", fmt.Errorf("unable to create folder, %w", err)
		}

		newFile, err := os.Create(newFilePath)
		if err != nil {
			return "", fmt.Errorf("unable to create file, %w", err)
		}
		if _, err := newFile.WriteString(*content); err != nil {
			return "", fmt.Errorf("unable to write string, %w", err)
		}
		if err := newFile.Close(); err != nil {
			return "", fmt.Errorf("unable to close file, %w", err)
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
