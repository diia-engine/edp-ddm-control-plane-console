package git

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
	cssh "golang.org/x/crypto/ssh"
)

const (
	tempDir = "/tmp"
)

type Service struct {
	path           string
	key            string
	_keyFilePath   string
	user           string
	commandCreator func(name string, arg ...string) Command
	TempDir        string
}

type User struct {
	Name  string
	Email string
}

func Make(path, user, key string) *Service {
	return &Service{
		path:    path,
		TempDir: tempDir,
		user:    user,
		key:     key,
	}
}

func (s *Service) GenerateChangeID() (string, error) {
	h := sha1.New()
	if _, err := h.Write([]byte(time.Now().Format(time.RFC3339))); err != nil {
		return "", errors.Wrap(err, "unable to write hash")
	}
	return fmt.Sprintf("I%x", h.Sum(nil)), nil
}

func (s *Service) Clean() error {
	if s._keyFilePath != "" {
		if err := os.RemoveAll(s._keyFilePath); err != nil {
			return errors.Wrap(err, "unable to clear key file")
		}
	}

	if err := os.RemoveAll(s.path); err != nil {
		return errors.Wrap(err, "unable to clear repo path")
	}

	return nil
}

func (s *Service) Clone(url string) error {
	keyPath, err := s.keyFilePath()
	if err != nil {
		return errors.Wrap(err, "unable to create key file")
	}

	cloneCMD := s.commandCreate("git", "clone", "--mirror", url, s.path)
	cloneCMD.SetEnv(s.authEnv(keyPath))

	if bts, err := cloneCMD.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "unable to clone repo: %s", string(bts))
	}

	if err := s.bareToNormal(s.path); err != nil {
		return errors.Wrap(err, "unable to covert bare repo to normal")
	}

	fetchCMD := s.commandCreate("git", "--git-dir", path.Join(s.path, ".git"), "pull", "origin", "master",
		"--unshallow", "--no-rebase")
	fetchCMD.SetEnv(s.authEnv(keyPath))
	bts, err := fetchCMD.CombinedOutput()
	if err != nil && !strings.Contains(string(bts), "does not make sense") {
		return errors.Wrapf(err, "unable to pull unshallow repo: %s", string(bts))
	}

	return nil
}

func (s *Service) SetAuthor(user *User) error {
	cmd := s.commandCreate("git", "config", "user.email", user.Email)
	cmd.SetDir(s.path)

	bts, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "unable to commit: %s", string(bts))
	}

	cmd = s.commandCreate("git", "config", "user.name", user.Name)
	cmd.SetDir(s.path)

	bts, err = cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "unable to commit: %s", string(bts))
	}

	return nil
}

func (s *Service) RawCommit(message string /*user *User,*/, params ...string) error {
	baseParams := []string{"commit" /* fmt.Sprintf("--author=\"%s <%s>\"", user.Name, user.Email),*/, "-m", message}
	baseParams = append(baseParams, params...)

	cmd := s.commandCreate("git", baseParams...)
	cmd.SetDir(s.path)
	//cmd.SetEnv([]string{fmt.Sprintf("GIT_AUTHOR_EMAIL=\"%s\"", user.Email),
	//	fmt.Sprintf("GIT_AUTHOR_NAME=\"%s\"", user.Name)})

	//GIT_AUTHOR_EMAIL="you@email.com" && GIT_AUTHOR_NAME="Your Name"

	msg, err := cmd.StrCombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "unable to commit: %s", msg)
	}

	return nil
}

func (s *Service) Commit(message string, files []string, user *User) error {
	_, w, err := s.worktree()
	if err != nil {
		return errors.Wrap(err, "unable to get worktree")
	}

	for _, f := range files {
		if _, err := w.Add(f); err != nil {
			return errors.Wrapf(err, "unable to add file: %s", f)
		}
	}

	if _, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{Name: user.Name, Email: user.Email, When: time.Now()},
	}); err != nil {
		return errors.Wrap(err, "unable to perform git commit")
	}

	return nil
}

func (s *Service) authEnv(keyPath string) []string {
	return []string{fmt.Sprintf(`GIT_SSH_COMMAND=ssh -i %s -l %s -o StrictHostKeyChecking=no`, keyPath,
		s.user), "GIT_SSH_VARIANT=ssh"}
}

func (s *Service) bareToNormal(path string) error {
	if err := os.MkdirAll(fmt.Sprintf("%s/.git", path), 0777); err != nil {
		return errors.Wrap(err, "unable to create .git folder")
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return errors.Wrap(err, "unable to list dir")
	}

	for _, f := range files {
		if f.Name() == ".git" {
			continue
		}

		if err := os.Rename(fmt.Sprintf("%s/%s", path, f.Name()),
			fmt.Sprintf("%s/.git/%s", path, f.Name())); err != nil {
			return errors.Wrap(err, "unable to rename file")
		}
	}

	gitDir := fmt.Sprintf("%s/.git", path)
	cmd := s.commandCreate("git", "--git-dir", gitDir, "config", "--local",
		"--bool", "core.bare", "false")
	if bts, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, string(bts))
	}

	cmd = s.commandCreate("git", "--git-dir", gitDir, "config", "--local",
		"--bool", "remote.origin.mirror", "false")
	if bts, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, string(bts))
	}

	cmd = s.commandCreate("git", "--git-dir", gitDir, "reset", "--hard")
	cmd.SetDir(path)
	if bts, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, string(bts))
	}

	return nil
}

func (s *Service) keyFilePath() (string, error) {
	if s._keyFilePath != "" {
		return s._keyFilePath, nil
	}

	keyFile, err := os.Create(fmt.Sprintf("%s/sshkey_%d", s.TempDir, time.Now().UnixNano()))
	if err != nil {
		return "", errors.Wrap(err, "unable to create temp file for ssh key")
	}
	keyFileInfo, _ := keyFile.Stat()
	keyFilePath := fmt.Sprintf("%s/%s", s.TempDir, keyFileInfo.Name())

	if _, err = keyFile.WriteString(s.key); err != nil {
		return "", errors.Wrap(err, "unable to write ssh key")
	}

	if err = keyFile.Close(); err != nil {
		return "", errors.Wrap(err, "unable to close file")
	}

	if err := os.Chmod(keyFilePath, 0400); err != nil {
		return "", errors.Wrap(err, "unable to chmod ssh key file")
	}

	s._keyFilePath = keyFilePath

	return keyFilePath, nil
}

func IsErrReferenceNotFound(err error) bool {
	if err == nil {
		return false
	}

	return errors.Cause(err).Error() == "reference not found"
}

func IsErrNonFastForwardUpdate(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "non-fast-forward update")
}

func (s *Service) Pull(remoteName string) (*object.Commit, error) {
	r, w, err := s.worktree()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get worktree")
	}

	keyPath, err := s.keyFilePath()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create key file")
	}

	publicKeys, err := ssh.NewPublicKeysFromFile(s.user, keyPath, "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create public keys")
	}
	publicKeys.HostKeyCallback = cssh.InsecureIgnoreHostKey()

	if err := w.Pull(&git.PullOptions{RemoteName: remoteName, Auth: publicKeys}); err != nil {
		return nil, errors.Wrap(err, "unable to pull")
	}

	ref, err := r.Head()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ref")
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return nil, errors.Wrap(err, "unable to get last commit")
	}

	return commit, nil
}

func (s *Service) Push(remoteName string, pushParams ...string) error {
	keyPath, err := s.keyFilePath()
	if err != nil {
		return errors.Wrap(err, "unable to init auth")
	}

	basePushParams := []string{"--git-dir", path.Join(s.path, ".git"),
		"push", remoteName}
	basePushParams = append(basePushParams, pushParams...)

	pushCMD := s.commandCreate("git", basePushParams...)
	pushCMD.SetEnv(s.authEnv(keyPath))
	pushCMD.SetDir(s.path)

	if bts, err := pushCMD.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "unable to push changes, err: %s", string(bts))
	}

	return nil
}

func (s *Service) SetFileContents(filePath, contents string) error {
	filePath = path.Join(s.path, filePath)

	fp, err := os.Create(filePath)
	if err != nil {
		return errors.Wrapf(err, "unable to create file: %s", filePath)
	}

	if _, err := fp.WriteString(contents); err != nil {
		return errors.Wrapf(err, "unable to put file contents, file: %s", filePath)
	}

	if err := fp.Close(); err != nil {
		return errors.Wrapf(err, "unable to close file: %s", filePath)
	}

	return nil
}

func (s *Service) GetFileContents(filePath string) (string, error) {
	filePath = path.Join(s.path, filePath)
	fp, err := os.Open(filePath)
	if err != nil {
		return "", errors.Wrapf(err, "unable to open file: %s", filePath)
	}

	bts, err := ioutil.ReadAll(fp)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read file: %s", filePath)
	}

	return string(bts), nil
}

func (s *Service) AddRemote(remoteName, url string) error {
	cmd := s.commandCreate("git", "remote", "add", remoteName, url)
	cmd.SetDir(s.path)

	if bts, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "unable to add remote, err: %s", string(bts))
	}

	return nil
}

func (s *Service) RawCheckout(branch string, create bool) error {
	baseParams := []string{"checkout", branch}
	if create {
		baseParams = []string{"checkout", "-b", branch}
	}

	cmd := s.commandCreate("git", baseParams...)
	cmd.SetDir(s.path)

	if bts, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "unable to checkout, err: %s", string(bts))
	}

	return nil
}

func (s *Service) Checkout(branch string, create bool) error {
	_, w, err := s.worktree()
	if err != nil {
		return errors.Wrap(err, "unable to get worktree")
	}

	if err := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch), Create: create}); err != nil {
		return errors.Wrap(err, "unable to checkout branch")
	}

	return nil
}

func (s *Service) DeleteBranch(branchName string) error {
	cmd := s.commandCreate("git", "branch", "-D", branchName)
	cmd.SetDir(s.path)

	msg, err := cmd.StrCombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "unable to delete branch, msg: %s", msg)
	}

	return nil
}

func (s *Service) Add(file string) error {
	cmd := s.commandCreate("git", "add", file)
	cmd.SetDir(s.path)

	msg, err := cmd.StrCombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "unable to run add, msg: %s", msg)
	}

	return nil
}

func (s *Service) Rebase(targetBranch string, params ...string) (string, error) {
	gitArgs := []string{"rebase", targetBranch}
	gitArgs = append(gitArgs, params...)

	cmd := s.commandCreate("git", gitArgs...)
	cmd.SetDir(s.path)

	msg, err := cmd.StrCombinedOutput()
	if err != nil {
		return msg, errors.Wrap(err, "unable to run rebase")
	}

	return msg, nil
}

func (s *Service) worktree() (*git.Repository, *git.Worktree, error) {
	r, err := git.PlainOpen(s.path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to open repo")
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get worktree")
	}

	return r, w, nil
}

func (s *Service) RemoveBranch(name string) error {
	r, err := git.PlainOpen(s.path)
	if err != nil {
		return errors.Wrap(err, "unable to open repo")
	}

	headRef, err := r.Head()
	if err != nil {
		return errors.Wrap(err, "unable to get head ref")
	}

	err = r.Storer.RemoveReference(
		plumbing.NewHashReference(plumbing.NewBranchReferenceName(name), headRef.Hash()).Name())

	if err != nil {
		return errors.Wrap(err, "unable to remove branch")
	}

	return nil
}

func ExtractMrURL(pushMessage string) string {
	return regexp.MustCompile(
		`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`).
		FindString(pushMessage)
}

func CommitMessageWithChangeID(commitMessage, changeID string) string {
	return fmt.Sprintf("%s\n\nChange-Id: %s", commitMessage, changeID)
}
