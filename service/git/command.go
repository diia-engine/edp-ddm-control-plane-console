package git

import "os/exec"

type Command interface {
	CombinedOutput() ([]byte, error)
	SetEnv(env []string)
	SetDir(dir string)
}

func (s *Service) commandCreate(name string, arg ...string) Command {
	if s.commandCreator == nil {
		s.commandCreator = func(name string, arg ...string) Command {
			cmd := exec.Command(name, arg...)
			return &command{cmd: cmd}
		}
	}

	return s.commandCreator(name, arg...)
}

type command struct {
	cmd *exec.Cmd
}

func (c *command) CombinedOutput() ([]byte, error) {
	return c.cmd.CombinedOutput()
}

func (c *command) SetEnv(env []string) {
	c.cmd.Env = env
}

func (c *command) SetDir(dir string) {
	c.cmd.Dir = dir
}
