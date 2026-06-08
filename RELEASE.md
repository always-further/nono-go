# Release Procedure

`nono-go` releases are Go module releases. The canonical release artifact is a
semver Git tag such as `v0.57.0`; there is no separate package registry publish
step for Go.

Pushing a release tag triggers the Release GitHub Action. The workflow validates
that the tag points at `origin/main`, runs the release test and lint matrix, then
creates the GitHub Release with generated notes.

## Prerequisites

- The release commit is merged to `main`.
- `main` is clean and up to date with `origin/main`.
- CI is green for the commit being released.
- The tag follows Go module semver: `vMAJOR.MINOR.PATCH`, for example
  `v0.57.0`. Prerelease tags such as `v0.58.0-rc.1` are allowed and become
  GitHub prereleases.

## Create a Release

From a clean `main` checkout:

```bash
git checkout main
git pull --ff-only
./scripts/release.sh v0.57.0
```

The script:

- validates the tag format
- ensures the working tree is clean
- ensures `HEAD` matches `origin/main`
- checks the tag does not already exist locally or remotely
- runs local Go checks
- creates an annotated tag
- pushes the tag to `origin`

After the tag is pushed, GitHub Actions runs `.github/workflows/release.yml`.
If validation passes, it creates the GitHub Release. Go users can then install
the release with:

```bash
go get github.com/always-further/nono-go@v0.57.0
```

The Go proxy and pkg.go.dev discover the version from the Git tag when the
module is requested.

## Dry Run

Use `--dry-run` to validate local state and checks without creating or pushing a
tag:

```bash
./scripts/release.sh --dry-run v0.57.0
```

Use `--skip-checks` only when the exact commit has already passed equivalent
checks:

```bash
./scripts/release.sh --skip-checks v0.57.0
```

## Recovery

If the tag exists but the GitHub Release was not created, rerun the Release
workflow manually with the existing tag:

1. Open **Actions**.
2. Select **Release**.
3. Choose **Run workflow**.
4. Enter the tag, for example `v0.57.0`.

The release creation step is idempotent. If the GitHub Release already exists,
the workflow reports that and exits successfully.

If the wrong tag was pushed, do not move it casually. Go module tags are cached
by proxies and consumers. Prefer publishing a new patch version.
