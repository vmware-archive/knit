package patcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const modulePrefix = "path = "

type Repo struct {
	debug          bool
	runner         runner
	repo           string
	gitPath        string
	committerName  string
	committerEmail string
}

type runner interface {
	Run(command Executor) error
	CombinedOutput(command Executor) ([]byte, error)
}

func NewRepo(runner runner, gitPath string, repo string, debug bool, committerName, committerEmail string) Repo {
	return Repo{
		debug:          debug,
		runner:         runner,
		repo:           repo,
		gitPath:        gitPath,
		committerName:  committerName,
		committerEmail: committerEmail,
	}
}

func (r Repo) ConfigureCommitter() error {
	commands := []*exec.Cmd{
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "config", "--global", "user.name", r.committerName}, Dir: r.repo},
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "config", "--global", "user.email", r.committerEmail}, Dir: r.repo},
	}

	for _, command := range commands {
		if r.debug {
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
		}

		if err := r.runner.Run(command); err != nil {
			return err
		}
	}

	return nil
}

func (r Repo) Checkout(checkoutRef string) error {
	commands := []*exec.Cmd{
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "checkout", checkoutRef}, Dir: r.repo},
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "clean", "-ffd"}, Dir: r.repo},
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "submodule", "update", "--init", "--recursive", "--force"}, Dir: r.repo},
	}

	for _, command := range commands {
		if r.debug {
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
		}

		if err := r.runner.Run(command); err != nil {
			return err
		}
	}

	return nil
}

func (r Repo) CleanSubmodules() error {
	submodules, err := r.submodules()
	if err != nil {
		return err
	}

	var commands []*exec.Cmd
	for _, submodule := range submodules {
		command := &exec.Cmd{Path: r.gitPath, Args: []string{"git", "clean", "-ffd"}, Dir: submodule}
		commands = append(commands, command)
	}

	for _, command := range commands {
		if r.debug {
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
		}

		if err := r.runner.Run(command); err != nil {
			return err
		}
	}

	return nil
}

func (r Repo) ApplyPatch(patch string) error {
	command := &exec.Cmd{Path: r.gitPath, Args: []string{"git", "am", patch}, Dir: r.repo}

	if r.debug {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	if err := r.runner.Run(command); err != nil {
		return err
	}

	return nil
}

func (r Repo) BumpSubmodule(path, sha string) error {
	commands := []*exec.Cmd{
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "checkout", sha}, Dir: filepath.Join(r.repo, path)},
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "submodule", "update", "--init", "--recursive", "--force"}, Dir: filepath.Join(r.repo, path)},
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "clean", "-ffd"}, Dir: filepath.Join(r.repo, path)},
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "add", "-A", path}, Dir: r.repo},
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "commit", "-m", fmt.Sprintf("Knit bump of %s", path)}, Dir: r.repo},
	}

	for _, command := range commands {
		if r.debug {
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
		}

		if err := r.runner.Run(command); err != nil {
			return err
		}
	}

	return nil
}

func (r Repo) PatchSubmodule(path, fullPathToPatch string) error {
	applyCommand := &exec.Cmd{Path: r.gitPath, Args: []string{"git", "am", fullPathToPatch}, Dir: filepath.Join(r.repo, path)}
	if r.debug {
		applyCommand.Stdout = os.Stdout
		applyCommand.Stderr = os.Stderr
	}

	if err := r.runner.Run(applyCommand); err != nil {
		return err
	}

	addCommand := &exec.Cmd{Path: r.gitPath, Args: []string{"git", "add", "-A", path}, Dir: r.repo}

	if output, err := r.runner.CombinedOutput(addCommand); err != nil {
		re := regexp.MustCompile(`^.*is in submodule '(.*)'`)
		submodulePath := re.FindStringSubmatch(string(output))[1]

		commands := []*exec.Cmd{
			&exec.Cmd{Path: r.gitPath, Args: []string{"git", "add", "-A", "."}, Dir: filepath.Join(r.repo, submodulePath)},
			&exec.Cmd{Path: r.gitPath, Args: []string{"git", "commit", "-m", fmt.Sprintf("Knit submodule patch of %s", submodulePath)}, Dir: filepath.Join(r.repo, submodulePath)},
		}

		for _, command := range commands {
			if r.debug {
				command.Stdout = os.Stdout
				command.Stderr = os.Stderr
			}

			if err := r.runner.Run(command); err != nil {
				return err
			}
		}
	}

	commitCommands := []*exec.Cmd{
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "add", "-A", "."}, Dir: r.repo},
		&exec.Cmd{Path: r.gitPath, Args: []string{"git", "commit", "-m", fmt.Sprintf("Knit patch of %s", path)}, Dir: r.repo},
	}

	for _, command := range commitCommands {
		if r.debug {
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
		}

		if err := r.runner.Run(command); err != nil {
			return err
		}
	}

	return nil
}

func (r Repo) CheckoutBranch(name string) error {
	command := &exec.Cmd{Path: r.gitPath, Args: []string{"git", "rev-parse", "--verify", name}, Dir: r.repo}
	if r.debug {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	if err := r.runner.Run(command); err == nil {
		return fmt.Errorf("Branch %q already exists. Please delete it before trying again", name)
	}

	command = &exec.Cmd{Path: r.gitPath, Args: []string{"git", "checkout", "-b", name}, Dir: r.repo}

	if r.debug {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	if err := r.runner.Run(command); err != nil {
		return err
	}
	return nil
}

func (r Repo) submodules() ([]string, error) {
	modules, err := ioutil.ReadFile(filepath.Join(r.repo, ".gitmodules"))
	if err != nil {
		return nil, err
	}

	var modulePaths []string
	lines := strings.Split(string(modules), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, modulePrefix) {
			modulePaths = append(modulePaths, strings.TrimPrefix(line, modulePrefix))
		}
	}

	var paths []string
	for _, modulePath := range modulePaths {
		fullModulePath := filepath.Join(r.repo, modulePath)
		_, err := os.Stat(fullModulePath)
		if os.IsNotExist(err) {
			continue
		}

		paths = append(paths, fullModulePath)
	}

	return paths, nil
}
