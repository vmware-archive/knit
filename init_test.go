package main_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var (
	patcher     string
	releaseRepo string
	patchesRepo string
)

var _ = BeforeSuite(func() {
	var err error
	patcher, err = gexec.Build("github.com/pivotal-cf-experimental/knit")

	releaseRepo = os.Getenv("CF_RELEASE_DIR")
	patchesRepo = os.Getenv("PCF_PATCHES_DIR")

	if releaseRepo == "" {
		Fail("CF_RELEASE_DIR is a required env var")
	}

	if patchesRepo == "" {
		Fail("PCF_PATCHES_DIR is a required env var")
	}
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestApplyPatches(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ApplyPatches Suite")
}
