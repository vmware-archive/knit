# knit
![Learning to knit](http://66.media.tumblr.com/tumblr_mcza9u6hux1rtgmslo1_500.gif)

A tool that allows you to apply a series of git patches / submodule bumps to a specified repository

## knit flags
Currently, knit has only four flags.

All of these flags are required:

- `--repository-to-patch - path to the repository you would like to apply patches to`
- `--patch-repository - path to the directory that contains your patches`
- `--version - the version you would like to jump to`

Optionally you can specify:

- `--quiet - suppress all of the ouput of the git commands that are being run`

## Running the command
Run knit like so:

```
knit --repository-to-patch /some/repository/cf-release --patch-repository /some/patches/repository/cf-release --version 1.7.2
```

pointing at a sub-directory in your patches repo that is an exact match for the repository-to-patch is VERY important

## Directory structure
knit relies on a very specific directory structure for the patches repository you supply. It has to look something like this:

```
cf-release - name of the top-level component
    └── 1.7 - start of your release versioning scheme
        ├── 2 - the patch level you are going to
        │   ├── another-example.patch - a top-level patch
        │   └── src
        │       └── loggregator
        │           └── example.patch - a submodule level patch
        └── starting-version.yml - lists starting version of each patch-level + submodule patch SHAs
```

## starting-versions.yml
The starting versions file has a section for each patch version (even if there are no associated patches) and looks like this:
```
---
starting_versions:
- version: 0
  ref: "v235"
- version: 1
  ref: "v235"
- version: 2
  ref: "v235"
  submodules:
    "src/uaa-release":
      ref: "jas4374357afasdfkgasfkdga890989080989"
    "src/capi-release/src/cloud_controller_ng":
      ref: "57afasdfkgasfkddsjfghj888328748723874"
```
