package fakes

import "github.com/pivotal-cf-experimental/knit/patcher"

type PatchSet struct {
	VersionsToApplyForCall struct {
		Receives struct {
			Version string
		}
		Returns struct {
			Versions []patcher.Version
			Error    error
		}
	}

	PatchesForCall struct {
		Receives struct {
			Version patcher.Version
		}
		Returns struct {
			Patches []string
			Error   error
		}
	}

	BumpsForCall struct {
		Receives struct {
			Version patcher.Version
		}
		Returns struct {
			Bumps map[string]string
			Error error
		}
	}

	SubmodulePatchesForCall struct {
		Receives struct {
			Version patcher.Version
		}
		Returns struct {
			SubmodulePatches map[string][]string
			Error            error
		}
	}
}

func (p *PatchSet) VersionsToApplyFor(version string) ([]patcher.Version, error) {
	p.VersionsToApplyForCall.Receives.Version = version

	return p.VersionsToApplyForCall.Returns.Versions, p.VersionsToApplyForCall.Returns.Error
}

func (p *PatchSet) PatchesFor(version patcher.Version) ([]string, error) {
	p.PatchesForCall.Receives.Version = version

	return p.PatchesForCall.Returns.Patches, p.PatchesForCall.Returns.Error
}

func (p *PatchSet) BumpsFor(version patcher.Version) (map[string]string, error) {
	p.BumpsForCall.Receives.Version = version

	return p.BumpsForCall.Returns.Bumps, p.BumpsForCall.Returns.Error
}

func (p *PatchSet) SubmodulePatchesFor(version patcher.Version) (map[string][]string, error) {
	p.SubmodulePatchesForCall.Receives.Version = version

	return p.SubmodulePatchesForCall.Returns.SubmodulePatches, p.SubmodulePatchesForCall.Returns.Error
}
