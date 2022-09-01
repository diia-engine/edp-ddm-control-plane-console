package registry

import (
	"ddm-admin-console/router"
	"fmt"
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

func (a *App) viewChange(ctx *gin.Context) (response *router.Response, retErr error) {
	changeID := ctx.Param("change")

	changes, err := a.getChangeContents(changeID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get changes")
	}

	return router.MakeResponse(200, "registry/change.html", gin.H{
		"page":    "registry",
		"changes": changes,
	}), nil
}

func (a *App) getChangeContents(changeID string) (string, error) {
	changeInfo, _, err := a.Gerrit.GoGerritClient().Changes.GetChangeDetail(changeID, &goGerrit.ChangeOptions{})
	if err != nil {
		return "", errors.Wrap(err, "unable to get gerrit change details")
	}

	files, _, err := a.Gerrit.GoGerritClient().Changes.ListFiles(changeID, currentRevision, &goGerrit.FilesOptions{})
	if err != nil {
		return "", errors.Wrap(err, "unable to get change files")
	}

	changes := make([]string, 0, len(files)-1)
	for fileName := range files {
		if fileName == "/COMMIT_MSG" {
			continue
		}

		changesContent, err := a.getFileChanges(changeID, fileName, changeInfo.Project, changeInfo.Branch)
		if err != nil {
			return "", errors.Wrap(err, "unable to get file changes")
		}

		changes = append(changes, changesContent)
	}

	return strings.Join(changes, ""), nil
}

func (a *App) getFileChanges(changeID, fileName, projectName, branchName string) (string, error) {
	originalContent, originalHttpRsp, err := a.Gerrit.GoGerritClient().Projects.GetBranchContent(projectName, branchName, url.PathEscape(fileName))
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
