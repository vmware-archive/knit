package patcher

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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
	Add     SubmoduleAddition
	Remove  bool
}

type SubmoduleAddition struct {
	URL    string
	Ref    string
	Branch string
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
	Major              int
	Minor              int
	Patch              int
	Ref                string
	Patches            []string
	SubmoduleBumps     map[string]string
	SubmodulePatches   map[string][]string
	SubmoduleAdditions map[string]SubmoduleAddition
	SubmoduleRemovals  []string
}

func (ps PatchSet) VersionsToApplyFor(version string) ([]Version, error) {
	majorVersion, minorVersion, patchVersion, hotfixVersion, err := ps.parseVersion(version)
	if err != nil {
		return nil, err
	}

	releaseDirName, err := ps.releaseDirName(majorVersion, minorVersion)
	if err != nil {
		return nil, err
	}

	startingVersions, err := ps.parseStartingVersionsFile(releaseDirName)
	if err != nil {
		return nil, err
	}

	var versions []Version
	var effectiveVersion Version
	for _, v := range startingVersions.Versions {
		vers := Version{
			Major:              majorVersion,
			Minor:              minorVersion,
			Patch:              v.Version,
			Ref:                v.Ref,
			SubmoduleBumps:     map[string]string{},
			SubmodulePatches:   map[string][]string{},
			SubmoduleAdditions: map[string]SubmoduleAddition{},
			SubmoduleRemovals:  []string{},
		}

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

			if len(submodulePatches) > 0 {
				vers.SubmodulePatches[path] = submodulePatches
			}

			if submodule.Add.URL != "" {
				if submodule.Add.Ref == "" {
					return nil, fmt.Errorf("Missing ref for new submodule: %q", path)
				}

				vers.SubmoduleAdditions[path] = submodule.Add
			}

			if submodule.Remove {
				vers.SubmoduleRemovals = append(vers.SubmoduleRemovals, path)
			}
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

func (ps PatchSet) releaseDirName(majorVersion, minorVersion int) (string, error) {
	path := filepath.Join(ps.path, fmt.Sprintf("%d.%d", majorVersion, minorVersion))
	_, err := os.Stat(path)
	if err == nil {
		return fmt.Sprintf("%d.%d", majorVersion, minorVersion), nil
	}

	path = filepath.Join(ps.path, fmt.Sprintf("%d", majorVersion), fmt.Sprintf("%d", minorVersion))
	_, err = os.Stat(path)
	if err == nil {
		return filepath.Join(fmt.Sprintf("%d", majorVersion), fmt.Sprintf("%d", minorVersion)), nil
	}
	return "", errors.New("please provide either major.minor or major/minor for directory structure")
}

func (ps PatchSet) parseStartingVersionsFile(releaseDirName string) (StartingVersions, error) {
	startingVersionsYAML, err := ioutil.ReadFile(filepath.Join(ps.path, releaseDirName, "starting-versions.yml"))
	if err != nil {
		return StartingVersions{}, errors.New("please provide a starting-versions.yml file")
	}

	var startingVersions StartingVersions
	err = yaml.Unmarshal(startingVersionsYAML, &startingVersions)
	if err != nil {
		return StartingVersions{}, err
	}

	return startingVersions, nil
}
