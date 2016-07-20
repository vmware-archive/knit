package patcher

import "fmt"

type StartingVersions struct {
	Versions []struct {
		Version    int
		Ref        string
		Submodules map[string]Submodule
		Patches    []string
	} `yaml:"starting_versions"`
}

type Checkpoint struct {
	Changes     []Changeset
	CheckoutRef string
	FinalBranch string
}

type Changeset struct {
	Patches          []string
	Bumps            map[string]string
	SubmodulePatches map[string][]string
}

type patchSet interface {
	VersionsToApplyFor(version string) ([]Version, error)
	SubmodulePatchesFor(Version) (submodulePatches map[string][]string, err error)
}

type Submodule struct {
	Ref string
}

type VersionsParser struct {
	version  string
	patchSet patchSet
}

func NewVersionsParser(version string, patchSet patchSet) VersionsParser {
	return VersionsParser{
		version:  version,
		patchSet: patchSet,
	}
}

func (p VersionsParser) GetCheckpoint() (Checkpoint, error) {
	var checkpoint Checkpoint

	versionsToApply, err := p.patchSet.VersionsToApplyFor(p.version)
	if err != nil {
		return Checkpoint{}, err
	}

	if len(versionsToApply) == 0 {
		return Checkpoint{}, fmt.Errorf("Missing starting version %q in starting-versions.yml", p.version)
	}

	for _, version := range versionsToApply {
		submodulePatches, err := p.patchSet.SubmodulePatchesFor(version)
		if err != nil {
			return Checkpoint{}, err
		}

		checkpoint.Changes = append(checkpoint.Changes, Changeset{
			Patches:          version.Patches,
			Bumps:            version.SubmoduleBumps,
			SubmodulePatches: submodulePatches,
		})
	}

	checkpoint.CheckoutRef = versionsToApply[0].Ref
	checkpoint.FinalBranch = p.version

	return checkpoint, nil
}
