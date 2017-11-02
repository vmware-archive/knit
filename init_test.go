package main_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var (
	patcher          string
	cfReleaseRepo    string
	cfPatchesDir     string
	diegoReleaseRepo string
	diegoPatchesDir  string
)

var _ = BeforeSuite(func() {
	var err error
	patcher, err = gexec.Build("github.com/pivotal-cf/knit")
	Expect(err).NotTo(HaveOccurred())

	cfReleaseRepo = os.Getenv("CF_RELEASE_DIR")
	cfPatchesDir = os.Getenv("CF_PATCHES_DIR")

	diegoReleaseRepo = os.Getenv("DIEGO_RELEASE_DIR")
	diegoPatchesDir = os.Getenv("DIEGO_PATCHES_DIR")

	if cfReleaseRepo == "" {
		Fail("CF_RELEASE_DIR is a required env var")
	}

	if cfPatchesDir == "" {
		Fail("CF_PATCHES_DIR is a required env var")
	}

	if diegoReleaseRepo == "" {
		Fail("DIEGO_RELEASE_DIR is a required env var")
	}

	if diegoPatchesDir == "" {
		Fail("DIEGO_PATCHES_DIR is a required env var")
	}
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestApplyPatches(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ApplyPatches Suite")
}
