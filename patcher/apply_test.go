package patcher_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/knit/patcher"
	"github.com/pivotal-cf-experimental/knit/patcher/fakes"
)

var _ = Describe("Apply", func() {
	var repo *fakes.Repository
	var apply patcher.Apply
	var checkpoint patcher.Checkpoint

	BeforeEach(func() {
		repo = &fakes.Repository{}
		apply = patcher.NewApply(repo)
		checkpoint = patcher.Checkpoint{
			Changes: []patcher.Changeset{
				{
					Patches: []string{"patch-1"},
					Bumps: map[string]string{
						"src/some-path": "some-other-sha",
					},
					SubmodulePatches: map[string][]string{
						"src/sub/path": []string{"path/to/other.patch"},
					},
					SubmoduleAdditions: map[string]patcher.SubmoduleAddition{
						"src/fake/sub": patcher.SubmoduleAddition{
							URL:    "fake-url",
							Ref:    "fake-ref",
							Branch: "fake-branch",
						},
					},
					SubmoduleRemovals: []string{
						"src/some-old-submodule",
						"src/other-unneeded-submodule",
					},
				},
				{
					Patches: []string{"patch-2"},
					Bumps: map[string]string{
						"src/some-other-path": "a-sha",
					},
					SubmodulePatches: map[string][]string{
						"src/some-other-sub/path": []string{"path/to/different.patch"},
					},
					SubmoduleAdditions: map[string]patcher.SubmoduleAddition{},
				},
			},
			CheckoutRef: "abcde12345",
			FinalBranch: "1.9.2",
		}
	})

	Describe("Checkpoint", func() {
		It("checkouts the initial ref defined by the checkpoint", func() {
			err := apply.Checkpoint(checkpoint)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.CheckoutCall.Receives.Ref).To(Equal("abcde12345"))
		})

		It("checks out a new branch from the initial ref", func() {
			err := apply.Checkpoint(checkpoint)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.CheckoutBranchCall.Receives.Name).To(Equal("1.9.2"))
		})

		It("applies the top-level patches", func() {
			err := apply.Checkpoint(checkpoint)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.ApplyPatchCall.Receives.Patches).To(Equal([]string{"patch-1", "patch-2"}))
		})

		It("add the new submodules", func() {
			err := apply.Checkpoint(checkpoint)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.AddSubmoduleCall.Receives.Submodules).To(Equal(map[string]patcher.SubmoduleAddition{
				"src/fake/sub": patcher.SubmoduleAddition{
					URL:    "fake-url",
					Ref:    "fake-ref",
					Branch: "fake-branch",
				},
			}))
		})

		It("removes the specified submodules", func() {
			err := apply.Checkpoint(checkpoint)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.RemoveSubmoduleCall.Receives.Paths).To(Equal([]string{"src/some-old-submodule", "src/other-unneeded-submodule"}))
		})

		It("bumps the submodules", func() {
			err := apply.Checkpoint(checkpoint)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.BumpSubmoduleCall.Receives.Submodules).To(Equal(map[string]string{
				"src/some-other-path": "a-sha",
				"src/some-path":       "some-other-sha",
			}))
		})

		It("patches individual submodules", func() {
			err := apply.Checkpoint(checkpoint)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.PatchSubmoduleCall.Receives.Paths).To(Equal([]string{"src/sub/path", "src/some-other-sub/path"}))
			Expect(repo.PatchSubmoduleCall.Receives.Patches).To(Equal([]string{"path/to/other.patch", "path/to/different.patch"}))
		})

		Context("when an error occurs", func() {
			Context("when checkout fails", func() {
				It("returns an error", func() {
					repo.CheckoutCall.Returns.Error = errors.New("meow")

					err := apply.Checkpoint(checkpoint)
					Expect(err).To(MatchError("meow"))

					Expect(repo.CheckoutBranchCall.Receives.Name).To(BeEmpty())
					Expect(repo.ApplyPatchCall.Receives.Patches).To(BeEmpty())
					Expect(repo.AddSubmoduleCall.Receives.Submodules).To(BeEmpty())
					Expect(repo.BumpSubmoduleCall.Receives.Submodules).To(BeEmpty())
					Expect(repo.PatchSubmoduleCall.Receives.Paths).To(BeEmpty())
				})
			})

			Context("when checkout branch fails", func() {
				It("returns an error", func() {
					repo.CheckoutBranchCall.Returns.Error = errors.New("meow")

					err := apply.Checkpoint(checkpoint)
					Expect(err).To(MatchError("meow"))

					Expect(repo.ApplyPatchCall.Receives.Patches).To(BeEmpty())
					Expect(repo.AddSubmoduleCall.Receives.Submodules).To(BeEmpty())
					Expect(repo.BumpSubmoduleCall.Receives.Submodules).To(BeEmpty())
					Expect(repo.PatchSubmoduleCall.Receives.Paths).To(BeEmpty())
				})
			})

			Context("when applying a patch fails", func() {
				It("returns an error", func() {
					repo.ApplyPatchCall.Returns.Error = errors.New("meow")

					err := apply.Checkpoint(checkpoint)
					Expect(err).To(MatchError("meow"))

					Expect(repo.AddSubmoduleCall.Receives.Submodules).To(BeEmpty())
					Expect(repo.BumpSubmoduleCall.Receives.Submodules).To(BeEmpty())
					Expect(repo.PatchSubmoduleCall.Receives.Paths).To(BeEmpty())
				})
			})

			Context("when adding a submodule fails", func() {
				It("returns an error", func() {
					repo.AddSubmoduleCall.Returns.Error = errors.New("meow")

					err := apply.Checkpoint(checkpoint)
					Expect(err).To(MatchError("meow"))
					Expect(repo.BumpSubmoduleCall.Receives.Submodules).To(BeEmpty())
					Expect(repo.PatchSubmoduleCall.Receives.Paths).To(BeEmpty())
					Expect(repo.RemoveSubmoduleCall.Receives.Paths).To(BeEmpty())
				})
			})

			Context("when removing a submodule fails", func() {
				It("returns an error", func() {
					repo.RemoveSubmoduleCall.Returns.Error = errors.New("meow")

					err := apply.Checkpoint(checkpoint)
					Expect(err).To(MatchError("meow"))
					Expect(repo.BumpSubmoduleCall.Receives.Submodules).To(BeEmpty())
					Expect(repo.PatchSubmoduleCall.Receives.Paths).To(BeEmpty())
				})
			})

			Context("when bumping the submodule fails", func() {
				It("returns an error", func() {
					repo.BumpSubmoduleCall.Returns.Error = errors.New("meow")

					err := apply.Checkpoint(checkpoint)
					Expect(err).To(MatchError("meow"))
					Expect(repo.PatchSubmoduleCall.Receives.Paths).To(BeEmpty())
				})
			})

			Context("when patching the submodule fails", func() {
				It("returns an error", func() {
					repo.PatchSubmoduleCall.Returns.Error = errors.New("meow")

					err := apply.Checkpoint(checkpoint)
					Expect(err).To(MatchError("meow"))
				})
			})
		})
	})
})
