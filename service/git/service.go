package git

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/uuid"
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
	changeID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid, %w", err)
	}

	h := sha1.New()

	if _, err := h.Write([]byte(changeID.String())); err != nil {
		return "", errors.Wrap(err, "failed to write hash")
	}

	return fmt.Sprintf("I%x", h.Sum(nil)), nil
}

func (s *Service) Clean() error {
	if s._keyFilePath != "" {
		if err := os.RemoveAll(s._keyFilePath); err != nil {
			return fmt.Errorf("unable to clear key file, %w", err)
		}
	}

	if err := os.RemoveAll(s.path); err != nil {
		return fmt.Errorf("unable to clear repo path, %w", err)
	}

	return nil
}

func (s *Service) Clone(url string) error {
	keyPath, err := s.keyFilePath()
	if err != nil {
		return fmt.Errorf("unable to create key file, %w", err)
	}

	cloneCMD := s.commandCreate("git", "clone", "--mirror", url, s.path)
	cloneCMD.SetEnv(s.authEnv(keyPath))

	if bts, err := cloneCMD.CombinedOutput(); err != nil {
		return fmt.Errorf("unable to clone repo: %s, %w", string(bts), err)
	}

	if err := s.bareToNormal(s.path); err != nil {
		return fmt.Errorf("unable to covert bare repo to normal, %w", err)
	}

	fetchCMD := s.commandCreate(
		"git",
		"--git-dir", path.Join(s.path, ".git"), "pull", "origin", "master", "--unshallow", "--no-rebase",
	)

	fetchCMD.SetEnv(s.authEnv(keyPath))

	if bts, err := fetchCMD.CombinedOutput(); err != nil && !strings.Contains(string(bts), "does not make sense") {
		return fmt.Errorf("unable to pull unshallow repo: %s, %w", string(bts), err)
	}

	return nil
}

func (s *Service) SetAuthor(user *User) error {
	cmd := s.commandCreate("git", "config", "user.email", user.Email)
	cmd.SetDir(s.path)

	if bts, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("unable to commit: %s, %w", string(bts), err)
	}

	cmd = s.commandCreate("git", "config", "user.name", user.Name)
	cmd.SetDir(s.path)

	if bts, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("unable to commit: %s, %w", string(bts), err)
	}

	return nil
}

func (s *Service) RawCommit(u *User, message string, params ...string) error {
	if err := s.SetAuthor(u); err != nil {
		return fmt.Errorf("unable to set author, %w", err)
	}

	baseParams := []string{"commit", "-m", message}
	baseParams = append(baseParams, params...)

	cmd := s.commandCreate("git", baseParams...)
	cmd.SetDir(s.path)

	if msg, err := cmd.StrCombinedOutput(); err != nil {
		return fmt.Errorf("unable to commit: %s, %w", msg, err)
	}

	return nil
}

func (s *Service) authEnv(keyPath string) []string {
	return []string{fmt.Sprintf(`GIT_SSH_COMMAND=ssh -i %s -l %s -o StrictHostKeyChecking=no`, keyPath,
		s.user), "GIT_SSH_VARIANT=ssh"}
}

func (s *Service) bareToNormal(path string) error {
	if err := os.MkdirAll(fmt.Sprintf("%s/.git", path), 0o777); err != nil {
		return fmt.Errorf("unable to create .git folder, %w", err)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("unable to list dir, %w", err)
	}

	for _, f := range files {
		if f.Name() == ".git" {
			continue
		}

		if err := os.Rename(fmt.Sprintf("%s/%s", path, f.Name()),
			fmt.Sprintf("%s/.git/%s", path, f.Name())); err != nil {
			return fmt.Errorf("unable to rename file, %w", err)
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
		return "", fmt.Errorf("unable to create temp file for ssh key, %w", err)
	}
	keyFileInfo, _ := keyFile.Stat()
	keyFilePath := fmt.Sprintf("%s/%s", s.TempDir, keyFileInfo.Name())

	if _, err = keyFile.WriteString(s.key); err != nil {
		return "", fmt.Errorf("unable to write ssh key, %w", err)
	}

	if err = keyFile.Close(); err != nil {
		return "", fmt.Errorf("unable to close file, %w", err)
	}

	if err := os.Chmod(keyFilePath, 0o400); err != nil {
		return "", fmt.Errorf("unable to chmod ssh key file, %w", err)
	}

	s._keyFilePath = keyFilePath

	return keyFilePath, nil
}

func IsErrReferenceNotFound(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "reference not found")
}

func (s *Service) RawPull(params ...string) error {
	keyPath, err := s.keyFilePath()
	if err != nil {
		return fmt.Errorf("unable to init auth, %w", err)
	}

	baseParams := []string{"pull"}
	baseParams = append(baseParams, params...)

	cmd := s.commandCreate("git", baseParams...)
	cmd.SetEnv(s.authEnv(keyPath))
	cmd.SetDir(s.path)

	msg, err := cmd.StrCombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to pull: %s, %w", msg, err)
	}

	return nil
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
		return nil, fmt.Errorf("unable to get worktree, %w", err)
	}

	keyPath, err := s.keyFilePath()
	if err != nil {
		return nil, fmt.Errorf("unable to create key file, %w", err)
	}

	publicKeys, err := ssh.NewPublicKeysFromFile(s.user, keyPath, "")
	if err != nil {
		return nil, fmt.Errorf("unable to create public keys, %w", err)
	}
	publicKeys.HostKeyCallback = cssh.InsecureIgnoreHostKey()

	if err := w.Pull(&git.PullOptions{RemoteName: remoteName, Auth: publicKeys}); err != nil {
		return nil, fmt.Errorf("unable to pull, %w", err)
	}

	ref, err := r.Head()
	if err != nil {
		return nil, fmt.Errorf("unable to get ref, %w", err)
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("unable to get last commit, %w", err)
	}

	return commit, nil
}

func (s *Service) Push(remoteName string, pushParams ...string) error {
	keyPath, err := s.keyFilePath()
	if err != nil {
		return fmt.Errorf("unable to init auth, %w", err)
	}

	basePushParams := []string{
		"--git-dir", path.Join(s.path, ".git"),
		"push", remoteName,
	}
	basePushParams = append(basePushParams, pushParams...)

	pushCMD := s.commandCreate("git", basePushParams...)
	pushCMD.SetEnv(s.authEnv(keyPath))
	pushCMD.SetDir(s.path)

	if bts, err := pushCMD.CombinedOutput(); err != nil {
		return fmt.Errorf("unable to push changes, err: %s, %w", string(bts), err)
	}

	return nil
}

func (s *Service) SetFileContents(filePath, contents string) error {
	filePath = path.Join(s.path, filePath)

	dir := path.Dir(filePath)
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return fmt.Errorf("unable to create dir, %w", err)
		}
	}

	fp, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("unable to create file: %s, %w", filePath, err)
	}

	if _, err := fp.WriteString(contents); err != nil {
		return fmt.Errorf("unable to put file contents, file: %s, %w", filePath, err)
	}

	if err := fp.Close(); err != nil {
		return fmt.Errorf("unable to close file: %s, %w", filePath, err)
	}

	return nil
}

func (s *Service) GetFileContents(filePath string) (string, error) {
	filePath = path.Join(s.path, filePath)
	fp, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %s, %w", filePath, err)
	}

	bts, err := io.ReadAll(fp)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %s, %w", filePath, err)
	}

	return string(bts), nil
}

func (s *Service) AddRemote(remoteName, url string) error {
	cmd := s.commandCreate("git", "remote", "add", remoteName, url)
	cmd.SetDir(s.path)

	if bts, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("unable to add remote, err: %s, %w", string(bts), err)
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
		return fmt.Errorf("unable to checkout, err: %s, %w", string(bts), err)
	}

	return nil
}

func (s *Service) Checkout(branch string, create bool) error {
	_, w, err := s.worktree()
	if err != nil {
		return fmt.Errorf("unable to get worktree, %w", err)
	}

	if err := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: create,
	}); err != nil {
		return fmt.Errorf("unable to checkout branch, %w", err)
	}

	return nil
}

func (s *Service) DeleteBranch(branchName string) error {
	cmd := s.commandCreate("git", "branch", "-D", branchName)
	cmd.SetDir(s.path)

	msg, err := cmd.StrCombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to delete branch, msg: %s, %w", msg, err)
	}

	return nil
}

func (s *Service) Add(file string) error {
	cmd := s.commandCreate("git", "add", file)
	cmd.SetDir(s.path)

	msg, err := cmd.StrCombinedOutput()
	if err != nil {
		return fmt.Errorf("unable to run add, msg: %s, %w", msg, err)
	}

	return nil
}

func (s *Service) worktree() (*git.Repository, *git.Worktree, error) {
	r, err := git.PlainOpen(s.path)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to open repo, %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get worktree, %w", err)
	}

	return r, w, nil
}

func (s *Service) RemoveBranch(name string) error {
	r, err := git.PlainOpen(s.path)
	if err != nil {
		return fmt.Errorf("unable to open repo, %w", err)
	}

	headRef, err := r.Head()
	if err != nil {
		return fmt.Errorf("unable to get head ref, %w", err)
	}

	err = r.Storer.RemoveReference(
		plumbing.NewHashReference(plumbing.NewBranchReferenceName(name), headRef.Hash()).Name())

	if err != nil {
		return fmt.Errorf("unable to remove branch, %w", err)
	}

	return nil
}

func CommitMessageWithChangeID(commitMessage, changeID string) string {
	return fmt.Sprintf("%s\n\nChange-Id: %s", commitMessage, changeID)
}
