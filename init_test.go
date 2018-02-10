package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var (
	pathToKnit string
)

var _ = BeforeSuite(func() {
	var err error
	pathToKnit, err = gexec.Build("github.com/pivotal-cf/knit")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestApplyPatches(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ApplyPatches Suite")
}

func initGitRepo(pathToRepo string) {
	err := ioutil.WriteFile(filepath.Join(pathToRepo, "file-in-repo.txt"), []byte("hello, world!"), os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	gitCmd := exec.Command("git", "init")
	gitCmd.Dir = pathToRepo
	output, err := gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))

	gitCmd = exec.Command("git", "add", ".")
	gitCmd.Dir = pathToRepo
	output, err = gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))

	gitCmd = exec.Command("git", "config", "user.name", "Knit Acceptance Test Committer")
	gitCmd.Dir = pathToRepo
	output, err = gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))

	gitCmd = exec.Command("git", "config", "user.email", "cf-release-engineering@pivotal.io")
	gitCmd.Dir = pathToRepo
	output, err = gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))

	gitCmd = exec.Command("git", "commit", "-m", "first commit")
	gitCmd.Dir = pathToRepo
	output, err = gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))
}

func createPatch(pathToRepo, patchesDir string) {
	f, err := os.OpenFile(filepath.Join(pathToRepo, "file-in-repo.txt"), os.O_APPEND|os.O_WRONLY, 0644)
	Expect(err).NotTo(HaveOccurred())

	defer f.Close()

	// creates patch for normal case
	_, err = f.WriteString("another change")
	Expect(err).NotTo(HaveOccurred())

	gitCmd := exec.Command("git", "add", ".")
	gitCmd.Dir = pathToRepo
	output, err := gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))

	gitCmd = exec.Command("git", "commit", "-m", "a change to the file")
	gitCmd.Dir = pathToRepo
	output, err = gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))

	gitCmd = exec.Command("git", "format-patch", "--stdout", "HEAD^")
	gitCmd.Dir = pathToRepo
	output, err = gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))

	err = ioutil.WriteFile(filepath.Join(patchesDir, "1.2", "change.patch"), []byte(output), os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	// creates patch for hotfix case
	_, err = f.WriteString("this is 2nd change")
	Expect(err).NotTo(HaveOccurred())

	gitCmd = exec.Command("git", "add", ".")
	gitCmd.Dir = pathToRepo
	output, err = gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))

	gitCmd = exec.Command("git", "commit", "-m", "a hotfix patch")
	gitCmd.Dir = pathToRepo
	output, err = gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))

	gitCmd = exec.Command("git", "format-patch", "--stdout", "HEAD^")
	gitCmd.Dir = pathToRepo
	output, err = gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))

	err = ioutil.WriteFile(filepath.Join(patchesDir, "1.2", "change2.patch"), []byte(output), os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	err = ioutil.WriteFile(filepath.Join(patchesDir, "1.2", "starting-versions.yml"), []byte(`---
starting_versions:
- version: 1
  ref: master
  patches:
  - change.patch
  hotfixes:
    "hot.fix":
      patches:
      - change2.patch
- version: 2
  ref: master
  submodules:
    "path/to/kiln":
      add:
        url: https://github.com/pivotal-cf/kiln.git
        ref: 7c018a3cd508e0b5541014362b353cde32d5c2a7
- version: 3
  ref: master
  submodules: {"path/to/kiln": { remove: true }}`), os.FileMode(int(0644)))
	Expect(err).NotTo(HaveOccurred())

	gitCmd = exec.Command("git", "reset", "HEAD^^", "--hard")
	gitCmd.Dir = pathToRepo
	output, err = gitCmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error: %s", output))
}
