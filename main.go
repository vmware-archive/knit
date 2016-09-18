package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pivotal-cf-experimental/knit/patcher"
)

var buildVersion string

func main() {
	var (
		releaseRepository string
		patchesRepository string
		version           string
		quiet             bool
		showBuildVersion  bool
	)

	flag.StringVar(&releaseRepository, "repository-to-patch", "", "")
	flag.StringVar(&patchesRepository, "patch-repository", "", "")
	flag.StringVar(&version, "version", "", "")
	flag.BoolVar(&quiet, "quiet", false, "")
	flag.BoolVar(&showBuildVersion, "v", false, "")
	flag.Parse()

	if showBuildVersion {
		if buildVersion == "" {
			buildVersion = "dev"
		}

		fmt.Printf("Knit version: %s\n", buildVersion)
		os.Exit(0)
	}

	var missingFlag string
	switch {
	case releaseRepository == "":
		missingFlag = "repository-to-patch is a required flag"
	case patchesRepository == "":
		missingFlag = "patch-repository is a required flag"
	case version == "":
		missingFlag = "version is a required flag"
	}

	if missingFlag != "" {
		log.Fatal(missingFlag)
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		log.Fatal(err)
	}

	versionsParser := patcher.NewVersionsParser(version, patcher.NewPatchSet(patchesRepository))
	runner, err := patcher.NewCommandRunner(gitPath, quiet)
	if err != nil {
		log.Fatal(err)
	}

	err = checkGitVersion(runner)
	if err != nil {
		log.Fatal(err)
	}

	repo := patcher.NewRepo(runner, releaseRepository, "bot", "witchcraft@example.com")
	apply := patcher.NewApply(repo)

	initialCheckpoint, err := versionsParser.GetCheckpoint()
	if err != nil {
		log.Fatal(err)
	}

	err = apply.Checkpoint(initialCheckpoint)
	if err != nil {
		log.Fatal(err)
	}
}

func checkGitVersion(runner patcher.CommandRunner) error {
	out, err := runner.CombinedOutput(patcher.Command{
		Args: []string{"--version"},
	})
	matches := regexp.MustCompile(`.*(\d\.\d\.\d).*`).FindStringSubmatch(string(out))
	if len(matches) < 2 {
		return errors.New("could not determine `git` version")
	}

	parts := strings.Split(matches[1], ".")
	if len(parts) < 3 {
		return errors.New("could not determine `git` version")
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("could not determine `git` version: %s", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("could not determine `git` version: %s", err)
	}

	if major < 2 {
		return errors.New("knit requires a version of git >= 2.9.0")
	}

	if major == 2 && minor < 9 {
		return errors.New("knit requires a version of git >= 2.9.0")
	}

	return nil
}
