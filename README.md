# knit
![Learning to knit](http://66.media.tumblr.com/tumblr_mcza9u6hux1rtgmslo1_500.gif)

A tool that allows you to apply a series of git patches / submodule bumps to a specified repository

## knit flags
Currently, knit has only four flags.

All of these flags are required:

- `--repository-to-patch - path to the original repository you would like to apply patches to`
- `--patch-repository - path to the directory that contains all your patches for that repository`
- `--version - the version you would like to jump to`

Optionally you can specify:

- `--quiet - suppress all of the ouput of the git commands that are being run`

## Running the command
Run knit like so:

```
knit --repository-to-patch /my/original/repository/cf-release --patch-repository /my/patches/repository/cf-release --version 1.7.2
```

Pointing at the directory whose name is an exact match for the repository-to-patch is VERY important

## Directory structure
knit relies on a very specific directory structure for the patches repository you supply. It has to look something like this:

```
cf-release - Name of the original repository onto which you'll apply patches
    └── 1.7 - Major/minor of your semver-compliant release versioning scheme
          └─ starting-versions.yml - YAML file listing initial Github ref of each patch-version along with its patches, submodule patch SHAs, and submodule patches
          └─ another-example.patch - A top-level patch
          └─ src
            └── loggregator
                └── example.patch - A submodule level patch
```

## starting-versions.yml
The starting versions file has a section for each patch version and looks like this:

```
---
starting_versions:
- version: 0
  ref: "v235"
- version: 1
  ref: "v235"
  patches:
  - "path/to/patch/in/pcf-patches/under/minor-release-dir"
- version: 2
  ref: "v235"
  submodules:
    "path/to/submodule/from/root/of/original/repo":
      patches:
      - "path/to/patch/in/pcf-patches/under/minor-release-dir"
    "path/to/another/submodule/from/root/of/original/repo":
      ref: "57afasdfkgasfkddsjfghj888328748723874"
    "path/to/newly/added/submodule/from/root/of/original/repo":
      add:
        url: https://example.com/someuser/repo.git
        ref: 7c013a3cd565e0b5541014338b353cde45d5c2a7
- version: 3
  ref: "v235"
  submodules:
    "path/to/another/submodule/from/root/of/original/repo":
      remove: true
```
