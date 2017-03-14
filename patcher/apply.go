package patcher

import "sort"

type Apply struct {
	repo repository
}

type repository interface {
	ConfigureCommitter() error
	Checkout(checkoutRef string) error
	CheckoutBranch(name string) error
	ApplyPatch(patch string) error
	AddSubmodule(path, url, ref, branch string) error
	RemoveSubmodule(path string) error
	BumpSubmodule(path, sha string) error
	PatchSubmodule(path string, patch string) error
}

func NewApply(repo repository) Apply {
	return Apply{
		repo: repo,
	}
}

func (a Apply) Checkpoint(checkpoint Checkpoint) error {
	err := a.repo.ConfigureCommitter()
	if err != nil {
		return err
	}

	err = a.repo.Checkout(checkpoint.CheckoutRef)
	if err != nil {
		return err
	}

	err = a.repo.CheckoutBranch(checkpoint.FinalBranch)
	if err != nil {
		return err
	}

	for _, change := range checkpoint.Changes {
		for _, patch := range change.Patches {
			err := a.repo.ApplyPatch(patch)
			if err != nil {
				return err
			}
		}

		for path, addition := range change.SubmoduleAdditions {
			err := a.repo.AddSubmodule(path, addition.URL, addition.Ref, addition.Branch)
			if err != nil {
				return err
			}
		}

		for _, path := range change.SubmoduleRemovals {
			err := a.repo.RemoveSubmodule(path)
			if err != nil {
				return err
			}
		}

		paths := sortSubmodules(change.Bumps)

		for _, path := range paths {
			sha := change.Bumps[path]
			err := a.repo.BumpSubmodule(path, sha)
			if err != nil {
				return err
			}
		}

		submodulePaths := sortSubmodulePatches(change.SubmodulePatches)

		for _, submodulePath := range submodulePaths {
			for _, patch := range change.SubmodulePatches[submodulePath] {
				err := a.repo.PatchSubmodule(submodulePath, patch)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func sortSubmodules(submodules map[string]string) []string {
	var sortedPaths []string

	for path, _ := range submodules {
		sortedPaths = append(sortedPaths, path)
	}

	sort.Strings(sortedPaths)

	return sortedPaths
}

func sortSubmodulePatches(submodulePatches map[string][]string) []string {
	var sortedPaths []string

	for path, _ := range submodulePatches {
		sortedPaths = append(sortedPaths, path)
	}

	sort.Strings(sortedPaths)

	return sortedPaths
}
