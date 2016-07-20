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
}

func (p *PatchSet) VersionsToApplyFor(version string) ([]patcher.Version, error) {
	p.VersionsToApplyForCall.Receives.Version = version

	return p.VersionsToApplyForCall.Returns.Versions, p.VersionsToApplyForCall.Returns.Error
}
