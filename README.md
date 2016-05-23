# knit

A tool that allows you to apply a series of git patches / submodule bumps to a specified repo

## knit flags

Currently, knit has only four flags.

All of these flags are required:

- `--repository-to-patch - the repository you would like to apply patches to`
- `--patches-repository - the repository that contains your patches`
- `--version - the version you would like to jump to`

Optionally you can specify:

- `--debug - show all of the ouput of the git commands that are being run`

## running the command

run knit like so:

```
knit --repository-to-patch /some/repository/my-release --patches-repository /some/patches/repository/my-release --version 2.0.0
```

pointing at a sub-directory in your patches repo that is an exact match for the repository-to-patch is VERY important

## directory structure

knit relies on a very specific directory structure for the patches repository you supply. It has to look something like this:

```
some-component - name of the top-level component
    └── 2.0 - start of your release versioning scheme
        ├── 8 - the patch level you are going to
        │   ├── another-example.patch - a top-level patch
        │   └── src
        │       └── github.com
        │           └── a-repo
        │               └── some-component
        │                   └── example.patch - a submodule level patch
        └── starting-version.yml - lists starting version of each patch-level + submodule patch SHAs
```

## starting-versions.yml

the starting versions file has a section for each patch version (even if there are no associated patches) and looks like this:
```
---
starting_versions:
- version: 0
  ref: "v42"
  submodules:
    "src/github.com/path-to-submodule": jas4374357afasdfkgasfkdga890989080989
    "src/github.com/another-patch/some-submodule": jas4374357afasdfkgasfkddsjfghj888328748723874
```
