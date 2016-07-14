package main

import (
	"flag"
	"log"
	"os/exec"

	"github.com/pivotal-cf-experimental/knit/patcher"
)

func main() {
	var (
		releaseRepository string
		patchesRepository string
		version           string
		debug             bool
	)

	flag.StringVar(&releaseRepository, "repository-to-patch", "", "")
	flag.StringVar(&patchesRepository, "patch-repository", "", "")
	flag.StringVar(&version, "version", "", "")
	flag.BoolVar(&debug, "debug", false, "")
	flag.Parse()

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
		log.Fatalln(missingFlag)
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		log.Fatalln(err)
	}

	versionsParser := patcher.NewVersionsParser(version, patcher.NewPatchSet(patchesRepository))
	runner, err := patcher.NewCommandRunner(gitPath, debug)
	if err != nil {
		log.Fatalln(err)
	}

	repo := patcher.NewRepo(runner, releaseRepository, "bot", "witchcraft@example.com")
	apply := patcher.NewApply(repo)

	initialCheckpoint, err := versionsParser.GetCheckpoint()
	if err != nil {
		log.Fatalln(err)
	}

	err = apply.Checkpoint(initialCheckpoint)
	if err != nil {
		log.Fatalln(err)
	}
}
