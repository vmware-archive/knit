package patcher_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/knit/patcher"
)

const startingVersionsContent = `---
starting_versions:
- version: 0
  ref: 'v122'
- version: 1
  ref: 'v123'
- version: 2
  ref: 'v124'
  patches:
  - Top-1.patch
  - Top-2.patch
  submodules:
    "src/fake-sub-1":
      ref: fake-sha-1
      patches:
      - Sub-1.patch
    "src/fake-sub-2":
      patches:
      - Sub-2.patch
`

var _ = Describe("PatchSet", func() {
	var (
		ps                   patcher.PatchSet
		patchesRepo          string
		startingVersionsYAML string
		files                []string
	)

	BeforeEach(func() {
		tmpDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		startingVersionsYAML = filepath.Join(tmpDir, "some-release", "1.9", "starting-versions.yml")
		err = os.MkdirAll(filepath.Dir(startingVersionsYAML), 0755)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(startingVersionsYAML, []byte(startingVersionsContent), 0644)
		Expect(err).NotTo(HaveOccurred())

		patchesRepo = filepath.Join(tmpDir, "some-release")

		err = os.MkdirAll(filepath.Join(patchesRepo, "1.9", "2", "src", "something"), 0755)
		Expect(err).NotTo(HaveOccurred())

		files = []string{
			filepath.Join(patchesRepo, "1.9", "Top-1.patch"),
			filepath.Join(patchesRepo, "1.9", "Top-2.patch"),
			filepath.Join(patchesRepo, "1.9", "Sub-1.patch"),
			filepath.Join(patchesRepo, "1.9", "Sub-2.patch"),
		}

		for _, file := range files {
			f, err := ioutil.TempFile(filepath.Dir(file), "")
			Expect(err).NotTo(HaveOccurred())

			err = os.Rename(f.Name(), file)
			Expect(err).NotTo(HaveOccurred())
		}

		ps = patcher.NewPatchSet(patchesRepo)
	})

	AfterEach(func() {
		err := os.RemoveAll(patchesRepo)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("VersionsToApplyFor", func() {
		It("returns the versions to apply based on the specified version", func() {
			versions, err := ps.VersionsToApplyFor("1.9.2")
			Expect(err).NotTo(HaveOccurred())

			Expect(versions).To(Equal([]patcher.Version{
				{
					Major: 1,
					Minor: 9,
					Patch: 2,
					Ref:   "v124",
					Patches: []string{
						filepath.Join(patchesRepo, "1.9", "Top-1.patch"),
						filepath.Join(patchesRepo, "1.9", "Top-2.patch"),
					},
					SubmoduleBumps: map[string]string{
						"src/fake-sub-1": "fake-sha-1",
					},
					SubmodulePatches: map[string][]string{
						"src/fake-sub-1": {
							filepath.Join(patchesRepo, "1.9", "Sub-1.patch"),
						},
						"src/fake-sub-2": {
							filepath.Join(patchesRepo, "1.9", "Sub-2.patch"),
						},
					},
				},
			}))
		})

		Context("when an error occurs", func() {
			Context("when the starting-versions file does not exist", func() {
				BeforeEach(func() {
					ps = patcher.NewPatchSet("/some/broken/patch")
				})

				It("returns an error", func() {
					_, err := ps.VersionsToApplyFor("1.9.2")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})

				Context("when the version can't be parsed", func() {
					It("returns an error", func() {
						_, err := ps.VersionsToApplyFor("%$#")
						Expect(err).To(MatchError(ContainSubstring("invalid syntax")))
					})
				})
			})

			Context("when the starting versions yaml cannot be parsed", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(startingVersionsYAML, []byte("%%%"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := ps.VersionsToApplyFor("1.9.2")
					Expect(err).To(MatchError("yaml: could not find expected directive name"))
				})
			})
		})
	})
})
