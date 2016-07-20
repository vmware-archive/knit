package patcher_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/knit/patcher"
	"github.com/pivotal-cf-experimental/knit/patcher/fakes"
)

var _ = Describe("VersionsParser", func() {
	Describe("Checkpoint", func() {
		var (
			patchSet *fakes.PatchSet
			vp       patcher.VersionsParser
		)

		BeforeEach(func() {
			patchSet = &fakes.PatchSet{}
			vp = patcher.NewVersionsParser("1.9.2", patchSet)
		})

		It("returns the checkpoint of the patches repository", func() {
			patchSet.VersionsToApplyForCall.Returns.Versions = []patcher.Version{
				{
					Major: 1,
					Minor: 9,
					Patch: 2,
					Ref:   "v124",
					SubmoduleBumps: map[string]string{
						"src/foo": "ref-1",
						"src/bar": "ref-2",
					},
					Patches: []string{
						"patch-1",
						"patch-2",
						"patch-3",
					},
				},
			}
			patchSet.SubmodulePatchesForCall.Returns.SubmodulePatches = map[string][]string{
				"src/submodule1": []string{
					"patch-repo/release/1.6/2/src/submodule1/foo.patch",
					"patch-repo/release/1.6//2/src/submodule1/foo2.patch",
				},
			}

			checkpoint, err := vp.GetCheckpoint()
			Expect(err).NotTo(HaveOccurred())

			Expect(checkpoint).To(Equal(patcher.Checkpoint{
				Changes: []patcher.Changeset{
					{
						Patches: []string{"patch-1", "patch-2", "patch-3"},
						Bumps: map[string]string{
							"src/foo": "ref-1",
							"src/bar": "ref-2",
						},
						SubmodulePatches: map[string][]string{
							"src/submodule1": []string{
								"patch-repo/release/1.6/2/src/submodule1/foo.patch",
								"patch-repo/release/1.6//2/src/submodule1/foo2.patch",
							},
						},
					},
				},
				CheckoutRef: "v124",
				FinalBranch: "1.9.2",
			}))

			Expect(patchSet.VersionsToApplyForCall.Receives.Version).To(Equal("1.9.2"))
			Expect(patchSet.SubmodulePatchesForCall.Receives.Version).To(Equal(patcher.Version{
				Major: 1,
				Minor: 9,
				Patch: 2,
				Ref:   "v124",
				Patches: []string{
					"patch-1",
					"patch-2",
					"patch-3",
				},
				SubmoduleBumps: map[string]string{
					"src/foo": "ref-1",
					"src/bar": "ref-2",
				},
			}))
		})

		Context("when an error occurs", func() {
			Context("when the patchset fails to find versions", func() {
				It("returns an error", func() {
					patchSet.VersionsToApplyForCall.Returns.Error = errors.New("failed to find versions")

					_, err := vp.GetCheckpoint()
					Expect(err).To(MatchError(ContainSubstring("failed to find versions")))
				})
			})

			Context("when the patchset finds no versions to apply", func() {
				It("returns an error", func() {
					patchSet.VersionsToApplyForCall.Returns.Versions = []patcher.Version{}

					_, err := vp.GetCheckpoint()
					Expect(err).To(MatchError(ContainSubstring(`Missing starting version "1.9.2" in starting-versions.yml`)))
				})
			})

			Context("when the patchset fails to find submodule patches", func() {
				It("returns an error", func() {
					patchSet.VersionsToApplyForCall.Returns.Versions = []patcher.Version{
						{Major: 1, Minor: 9, Patch: 2, Ref: "v124"},
					}
					patchSet.SubmodulePatchesForCall.Returns.Error = errors.New("failed to find any submodule patches")

					_, err := vp.GetCheckpoint()
					Expect(err).To(MatchError(ContainSubstring("failed to find any submodule patches")))
				})
			})
		})
	})
})
