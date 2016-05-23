package patcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const modulePrefix = "path = "

type commandRunner interface {
	Run(command Command) (err error)
	CombinedOutput(command Command) ([]byte, error)
}

type Repo struct {
	debug          bool
	runner         commandRunner
	repo           string
	gitPath        string
	committerName  string
	committerEmail string
}

func NewRepo(commandRunner commandRunner, gitPath string, repo string, debug bool, committerName, committerEmail string) Repo {
	return Repo{
		debug:          debug,
		runner:         commandRunner,
		repo:           repo,
		gitPath:        gitPath,
		committerName:  committerName,
		committerEmail: committerEmail,
	}
}

func (r Repo) ConfigureCommitter() error {

	commands := []Command{
		Command{
			Executable: r.gitPath,
			Args:       []string{"config", "--global", "user.name", r.committerName},
			Dir:        r.repo,
		},
		Command{
			Executable: r.gitPath,
			Args:       []string{"config", "--global", "user.email", r.committerEmail},
			Dir:        r.repo,
		},
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
	commands := []Command{
		Command{
			Executable: r.gitPath,
			Args:       []string{"checkout", checkoutRef},
			Dir:        r.repo,
		},
		Command{
			Executable: r.gitPath,
			Args:       []string{"clean", "-ffd"},
			Dir:        r.repo,
		},
		Command{
			Executable: r.gitPath,
			Args:       []string{"submodule", "update", "--init", "--recursive", "--force"},
			Dir:        r.repo,
		},
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

	var commands = []Command{}
	for _, submodule := range submodules {
		command := Command{
			Executable: r.gitPath,
			Args:       []string{"clean", "-ffd"},
			Dir:        submodule,
		}
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
	command := Command{
		Executable: r.gitPath,
		Args:       []string{"am", patch},
		Dir:        r.repo,
	}

	if r.debug {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	err := r.runner.Run(command)
	if err != nil {
		return err
	}

	return nil
}

func (r Repo) BumpSubmodule(path, sha string) error {

	pathToSubmodule := filepath.Join(r.repo, path)

	commands := []Command{
		Command{
			Executable: r.gitPath,
			Args:       []string{"checkout", sha},
			Dir:        pathToSubmodule,
		},
		Command{
			Executable: r.gitPath,
			Args:       []string{"submodule", "update", "--init", "--recursive", "--force"},
			Dir:        pathToSubmodule,
		},
		Command{
			Executable: r.gitPath,
			Args:       []string{"clean", "-ffd"},
			Dir:        pathToSubmodule,
		},
		Command{
			Executable: r.gitPath,
			Args:       []string{"add", "-A", path},
			Dir:        r.repo,
		},
		Command{
			Executable: r.gitPath,
			Args:       []string{"commit", "-m", fmt.Sprintf("Knit bump of %s", path), "--no-verify"},
			Dir:        r.repo,
		},
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
	applyCommand := Command{
		Executable: r.gitPath,
		Args:       []string{"am", fullPathToPatch},
		Dir:        filepath.Join(r.repo, path),
	}

	if r.debug {
		applyCommand.Stdout = os.Stdout
		applyCommand.Stderr = os.Stderr
	}

	if err := r.runner.Run(applyCommand); err != nil {
		return err
	}

	addCommand := Command{
		Executable: r.gitPath,
		Args:       []string{"add", "-A", path},
		Dir:        r.repo,
	}

	if output, err := r.runner.CombinedOutput(addCommand); err != nil {
		//TODO take this one out as a constant
		re := regexp.MustCompile(`^.*is in submodule '(.*)'`)
		submodulePath := re.FindStringSubmatch(string(output))[1]
		absoluteSubmodulePath := filepath.Join(r.repo, submodulePath)

		commands := []Command{
			Command{
				Executable: r.gitPath,
				Args:       []string{"add", "-A", "."},
				Dir:        absoluteSubmodulePath,
			},
			Command{
				Executable: r.gitPath,
				Args:       []string{"commit", "-m", fmt.Sprintf("Knit submodule patch of %s", submodulePath), "--no-verify"},
				Dir:        absoluteSubmodulePath,
			},
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

	commitCommands := []Command{
		Command{
			Executable: r.gitPath,
			Args:       []string{"add", "-A", "."},
			Dir:        r.repo,
		},
		Command{
			Executable: r.gitPath,
			Args:       []string{"commit", "-m", fmt.Sprintf("Knit patch of %s", path), "--no-verify"},
			Dir:        r.repo,
		},
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
	command := Command{
		Executable: r.gitPath,
		Args:       []string{"rev-parse", "--verify", name},
		Dir:        r.repo,
	}

	if r.debug {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	if err := r.runner.Run(command); err == nil {
		return fmt.Errorf("Branch %q already exists. Please delete it before trying again", name)
	}

	command = Command{
		Executable: r.gitPath,
		Args:       []string{"checkout", "-b", name},
		Dir:        r.repo,
	}

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
