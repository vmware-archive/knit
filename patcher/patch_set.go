package patcher

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type StartingVersions struct {
	Versions []struct {
		Version    int
		Ref        string
		Submodules map[string]Submodule
		Patches    []string
		Hotfixes   map[string]Hotfix
	} `yaml:"starting_versions"`
}

type Submodule struct {
	Ref     string
	Patches []string
}

type Hotfix struct {
	Patches    []string
	Submodules map[string]Submodule
}

type PatchSet struct {
	path string
}

func NewPatchSet(path string) PatchSet {
	return PatchSet{path}
}

type Version struct {
	Major            int
	Minor            int
	Patch            int
	Ref              string
	Patches          []string
	SubmoduleBumps   map[string]string
	SubmodulePatches map[string][]string
}

func (ps PatchSet) VersionsToApplyFor(version string) ([]Version, error) {
	majorVersion, minorVersion, patchVersion, hotfixVersion, err := ps.parseVersion(version)
	if err != nil {
		return nil, err
	}

	startingVersions, err := ps.parseStartingVersionsFile(majorVersion, minorVersion)
	if err != nil {
		return nil, err
	}

	var versions []Version
	var effectiveVersion Version
	for _, v := range startingVersions.Versions {
		vers := Version{
			Major:            majorVersion,
			Minor:            minorVersion,
			Patch:            v.Version,
			Ref:              v.Ref,
			SubmoduleBumps:   map[string]string{},
			SubmodulePatches: map[string][]string{},
		}

		releaseDirName := fmt.Sprintf("%d.%d", majorVersion, minorVersion)

		if v.Version == patchVersion && hotfixVersion != "" {
			if _, ok := v.Hotfixes[hotfixVersion]; ok {
				v.Patches = append(v.Patches, v.Hotfixes[hotfixVersion].Patches...)

				for path, submodule := range v.Hotfixes[hotfixVersion].Submodules {
					if v.Submodules == nil {
						v.Submodules = make(map[string]Submodule)
					}

					if _, ok := v.Submodules[path]; !ok {
						v.Submodules[path] = Submodule{}
					}

					hotfixRef := v.Submodules[path].Ref
					if submodule.Ref != "" {
						hotfixRef = submodule.Ref
					}

					v.Submodules[path] = Submodule{
						Ref:     hotfixRef,
						Patches: append(v.Submodules[path].Patches, submodule.Patches...),
					}
				}
			} else {
				return nil, fmt.Errorf("Hotfix not found: %q", hotfixVersion)
			}
		}

		for _, patch := range v.Patches {
			vers.Patches = append(vers.Patches, filepath.Join(ps.path, releaseDirName, patch))
		}

		for path, submodule := range v.Submodules {
			if submodule.Ref != "" {
				vers.SubmoduleBumps[path] = submodule.Ref
			}

			submodulePatches := []string{}

			for _, patch := range submodule.Patches {
				submodulePatches = append(submodulePatches, filepath.Join(ps.path, releaseDirName, patch))
			}

			vers.SubmodulePatches[path] = submodulePatches
		}

		if v.Version <= patchVersion {
			effectiveVersion = vers
		}

		versions = append(versions, vers)
	}

	var versionsToApply []Version
	for _, v := range versions {
		if v.Ref == effectiveVersion.Ref && v.Patch <= effectiveVersion.Patch {
			versionsToApply = append(versionsToApply, v)
		}
	}

	return versionsToApply, nil
}

func (ps PatchSet) parseVersion(version string) (int, int, int, string, error) {
	hotfixParts := strings.Split(version, "+")

	versionParts := strings.Split(hotfixParts[0], ".")
	majorVersion, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return -1, -1, -1, "", err
	}

	minorVersion, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return -1, -1, -1, "", err
	}

	patchVersion, err := strconv.Atoi(versionParts[2])
	if err != nil {
		return -1, -1, -1, "", err
	}

	var hotfixVersion string
	if len(hotfixParts) > 1 {
		hotfixVersion = hotfixParts[1]
	}

	return majorVersion, minorVersion, patchVersion, hotfixVersion, nil
}

func (ps PatchSet) parseStartingVersionsFile(majorVersion, minorVersion int) (StartingVersions, error) {
	startingVersionsYAML, err := ioutil.ReadFile(filepath.Join(ps.path, fmt.Sprintf("%d.%d", majorVersion, minorVersion), "starting-versions.yml"))
	if err != nil {
		return StartingVersions{}, err
	}

	var startingVersions StartingVersions
	err = yaml.Unmarshal(startingVersionsYAML, &startingVersions)
	if err != nil {
		return StartingVersions{}, err
	}

	return startingVersions, nil
}
