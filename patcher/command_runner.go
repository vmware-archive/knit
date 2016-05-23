package patcher

import (
	"io"
	"os/exec"
)

type Command struct {
	Executable string
	Args       []string
	Dir        string
	Stdout     io.Writer
	Stderr     io.Writer
}

type CommandRunner struct{}

func NewCommandRunner() CommandRunner {
	return CommandRunner{}
}

func (r CommandRunner) CombinedOutput(command Command) ([]byte, error) {
	path, err := exec.LookPath(command.Executable)
	if err != nil {
		return []byte{}, err
	}

	cmd := &exec.Cmd{
		Path:   path,
		Args:   append([]string{path}, command.Args...),
		Dir:    command.Dir,
		Stdout: command.Stdout,
		Stderr: command.Stderr,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	return output, nil
}

func (r CommandRunner) Run(command Command) error {
	path, err := exec.LookPath(command.Executable)
	if err != nil {
		return err
	}

	cmd := &exec.Cmd{
		Path:   path,
		Args:   append([]string{path}, command.Args...),
		Dir:    command.Dir,
		Stdout: command.Stdout,
		Stderr: command.Stderr,
	}

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
