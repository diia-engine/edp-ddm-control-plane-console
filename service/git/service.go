package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
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

func Make(path, user, key string) *Service {
	return &Service{
		path:    path,
		TempDir: tempDir,
		user:    user,
		key:     key,
	}
}

func (s *Service) Clone(url string) error {
	//cmd := s.commandCreate("git", "clone", url, s.path)
	//bts, err := cmd.CombinedOutput()
	//if err != nil {
	//	return errors.Wrapf(err, "clone execute error: %s", string(bts))
	//}

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

	keyFile, err := os.Create(fmt.Sprintf("%s/sshkey_%d", s.TempDir, time.Now().Unix()))
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

func (s *Service) AddRemote(remoteName, url string) error {
	cmd := s.commandCreate("git", "remote", "add", remoteName, url)
	cmd.SetDir(s.path)

	if bts, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "unable to add remote, err: %s", string(bts))
	}

	return nil
}
