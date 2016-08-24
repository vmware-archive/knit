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
    "src/fake-new-sub":
      add:
        ref: fake-sha-3
        url: fake-url
        branch: fake-branch
  hotfixes:
    "something.else":
      patches:
      - Top-88.patch
      submodules:
        "src/magic-fake-sub":
          ref: magic-fake-sha-1
          patches:
          - Sub-Magic.patch
- version: 3
  ref: 'v124'
  hotfixes:
    "urgent":
      submodules:
        "src/magic-fake-sub":
          ref: magic-fake-sha-1
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
			filepath.Join(patchesRepo, "1.9", "Top-88.patch"),
			filepath.Join(patchesRepo, "1.9", "Sub-Magic.patch"),
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
					SubmoduleAdditions: map[string]patcher.SubmoduleAddition{
						"src/fake-new-sub": patcher.SubmoduleAddition{
							URL:    "fake-url",
							Ref:    "fake-sha-3",
							Branch: "fake-branch",
						},
					},
				},
			}))
		})

		Context("when the specified version is not listed", func() {
			It("returns a valid version list", func() {
				versions, err := ps.VersionsToApplyFor("1.9.1")
				Expect(err).NotTo(HaveOccurred())

				Expect(versions).To(Equal([]patcher.Version{
					{
						Major:              1,
						Minor:              9,
						Patch:              0,
						Ref:                "v122",
						SubmoduleBumps:     map[string]string{},
						SubmodulePatches:   map[string][]string{},
						SubmoduleAdditions: map[string]patcher.SubmoduleAddition{},
					},
				}))
			})
		})

		Context("when a hotfix version is requested", func() {
			Context("when there are existing patches for the vanilla patch release", func() {
				It("includes the hotfix patches in the response", func() {
					versions, err := ps.VersionsToApplyFor("1.9.2+something.else")
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
								filepath.Join(patchesRepo, "1.9", "Top-88.patch"),
							},
							SubmoduleBumps: map[string]string{
								"src/fake-sub-1":     "fake-sha-1",
								"src/magic-fake-sub": "magic-fake-sha-1",
							},
							SubmodulePatches: map[string][]string{
								"src/fake-sub-1": {
									filepath.Join(patchesRepo, "1.9", "Sub-1.patch"),
								},
								"src/fake-sub-2": {
									filepath.Join(patchesRepo, "1.9", "Sub-2.patch"),
								},
								"src/magic-fake-sub": {
									filepath.Join(patchesRepo, "1.9", "Sub-Magic.patch"),
								},
							},
							SubmoduleAdditions: map[string]patcher.SubmoduleAddition{
								"src/fake-new-sub": patcher.SubmoduleAddition{
									URL:    "fake-url",
									Ref:    "fake-sha-3",
									Branch: "fake-branch",
								},
							},
						},
					}))
				})
			})

			Context("when there are no existing patches for the vanilla patch release", func() {
				It("includes the hotfix patches in the response", func() {
					Expect(func() {
						ps.VersionsToApplyFor("1.9.3+urgent")
					}).NotTo(Panic())
				})
			})
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

			Context("when the hotfix version does not exist", func() {
				It("returns an error", func() {
					_, err := ps.VersionsToApplyFor("1.9.2+does.not.exist")
					Expect(err).To(MatchError(`Hotfix not found: "does.not.exist"`))
				})
			})

			Context("when a new submodule is added without a ref", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(startingVersionsYAML, []byte(`
---					
starting_versions:
- version: 2
  ref: 'v124'
  submodules:
    "src/fake-new-sub":
      add:
        url: fake-url,
        branch: fake-branch`), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := ps.VersionsToApplyFor("1.9.2")
					Expect(err).To(MatchError(`Missing ref for new submodule: "src/fake-new-sub"`))
				})
			})
		})
	})
})
