package registry

import (
	"context"
	"ddm-admin-console/router"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	goGerrit "github.com/andygrunwald/go-gerrit"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	currentRevision = "current"
	mergeList       = "/MERGE_LIST"
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

	if err := a.Gerrit.ApproveAndSubmitChange(changeID, ctx.GetString(router.UserNameSessionKey),
		ctx.GetString(router.UserEmailSessionKey)); err != nil {
		return nil, errors.Wrap(err, "unable to approve change")
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

	changes, err := a.getChangeContents(ctx, changeInfo)
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

func (a *App) getChangeContents(ctx context.Context, changeInfo *goGerrit.ChangeInfo) (string, error) {
	files, _, err := a.Gerrit.GoGerritClient().Changes.ListFiles(changeInfo.ID, currentRevision, &goGerrit.FilesOptions{})
	if err != nil {
		return "", errors.Wrap(err, "unable to get change files")
	}

	changes := make([]string, 0, len(files)-1)
	for fileName := range files {
		if fileName == "/COMMIT_MSG" {
			continue
		}

		if fileName == mergeList {
			changesContent, err := a.getMergeListFileChanges(ctx, changeInfo.ID, changeInfo.Project)
			if err != nil {
				return "", errors.Wrap(err, "unable to load merge list")
			}

			changes = append(changes, changesContent...)
			continue
		}

		changesContent, err := a.getChangeFileChanges(changeInfo.ID, fileName, changeInfo.Project)
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

func (a *App) getMergeListFileChanges(ctx context.Context, changeID, projectName string) ([]string, error) {
	//rq, _ := a.Gerrit.GoGerritClient().NewRequest("GET",
	//	fmt.Sprintf("changes/%s/revisions/%s/mergeable", changeID, currentRevision), nil)
	//
	//a.Gerrit.GoGerritClient().Do(rq)
	////a.Gerrit.GoGerritClient().Changes.GetMergeable("")

	commits, err := a.Gerrit.GetMergeListCommits(ctx, changeID, currentRevision)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get merge list commits")
	}

	log.Println(commits)

	content, _, err := a.Gerrit.GoGerritClient().Changes.GetContent(changeID, currentRevision, mergeList)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get file content")
	}

	if err := a.decodeMergeListContent(*content, projectName); err != nil {
		return nil, errors.Wrap(err, "")
	}

	return []string{}, nil
}

func (a *App) decodeMergeListContent(content, projectName string) error {
	r := regexp.MustCompile("\\* ([a-f0-9]+) ")
	matches := r.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) == 2 {
			commitHash := m[1]

			commitInfo, _, err := a.Gerrit.GoGerritClient().Projects.GetCommit(projectName, commitHash)
			if err != nil {
				return errors.Wrapf(err, "unable to get commit info")
			}

			files, _, err := a.Gerrit.GoGerritClient().Changes.ListFiles(commitInfo.Commit, currentRevision, &goGerrit.FilesOptions{})
			if err != nil {
				return errors.Wrap(err, "")
			}

			log.Println(files)
		}
	}

	return nil
}

func (a *App) getChangeFileChanges(changeID, fileName, projectName string) (string, error) {
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
