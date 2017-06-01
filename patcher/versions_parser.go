package patcher

import "fmt"

type Checkpoint struct {
	Changes          []Changeset
	CheckoutRef      string
	FinalBranch      string
	ResultingVersion string
}

type Changeset struct {
	Patches            []string
	Bumps              map[string]string
	SubmodulePatches   map[string][]string
	SubmoduleAdditions map[string]SubmoduleAddition
	SubmoduleRemovals  []string
}

type patchSet interface {
	VersionsToApplyFor(version string) ([]Version, error)
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
		checkpoint.Changes = append(checkpoint.Changes, Changeset{
			Patches:            version.Patches,
			Bumps:              version.SubmoduleBumps,
			SubmodulePatches:   version.SubmodulePatches,
			SubmoduleAdditions: version.SubmoduleAdditions,
			SubmoduleRemovals:  version.SubmoduleRemovals,
		})
	}

	checkpoint.CheckoutRef = versionsToApply[0].Ref
	checkpoint.FinalBranch = p.version
	checkpoint.ResultingVersion = fmt.Sprintf("%s-%s", versionsToApply[0].Ref, p.version)

	return checkpoint, nil
}
