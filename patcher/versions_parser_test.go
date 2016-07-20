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
					SubmodulePatches: map[string][]string{
						"src/foo": {
							"foo-1.patch",
						},
						"src/bar": {
							"bar-1.patch",
						},
					},
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
							"src/foo": {
								"foo-1.patch",
							},
							"src/bar": {
								"bar-1.patch",
							},
						},
					},
				},
				CheckoutRef: "v124",
				FinalBranch: "1.9.2",
			}))

			Expect(patchSet.VersionsToApplyForCall.Receives.Version).To(Equal("1.9.2"))
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
		})
	})
})
