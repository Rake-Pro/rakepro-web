# Changelog

Notable changes to this repo. Newest first. Outstanding work is in
[backlog.md](backlog.md). Format loosely follows Keep a Changelog. Dates are
YYYY-MM-DD. Versions match the git tag / GHCR image tag.

## [Unreleased]

### Security
- CI: every action in the workflows is pinned to a commit SHA (version in a trailing comment; Dependabot keeps the pins current).
- Release: a manual run releases only on `main`; a pushed tag releases only if it is `vX.Y.Z` on a commit reachable from `main`. The image is pushed by digest, scanned by digest, and only then tagged `X.Y.Z`, `X.Y`, `sha-<short>` and `latest` (new job "Publish tags"), so a failed scan leaves `latest` on the previous release. Release versioning reads only plain `vX.Y.Z` tags; other `v*` tags are ignored.
- Trivy rescan: scans the digest behind `latest` (what production runs) instead of the newest version tag, with a warning when the two differ.
- Stream status poller: decodes at most 64 KiB of the status response and follows redirects only on the same host (at most 3).
- Static file server: directory paths under `/static/` return 404 instead of a file listing.
- Reference manifest and Makefile no longer name an image in an unrelated registry namespace: `deploy/k8s` uses a `ghcr.io/OWNER/rakepro-web` placeholder (set the image in `kustomization.yaml`), and `make docker-build` tags a local `rakepro-web` image unless `IMAGE` is set.

### Changed
- CI: release flow renamed from `main -> prod` to `dev -> main`. `dev` is the default working branch, `main` the protected release branch; the bot PR is now "Merge dev to main". Mechanics unchanged. Pushes to `dev` now also publish `:dev` and `:dev-<sha>` images.
- Release model -> `master`/`prod` branches. `master` is dev/default; merging a PR
  `master` -> `prod` builds and pushes `:latest` (+ `sha-<short>`), which
  deployment automation digest-pins into production. PRs targeting `prod` build without
  pushing (validation). `prod` is protected by a ruleset (PR + 1 approval +
  `build` check). `build-image.yml` triggers moved from `main` to `prod`.

## [0.1.2] - 2026-06-15

### Changed
- Swapped to the clean high-res branding assets: `icon-circuit-r` (icon) and
  `wordmark` (wordmark), replacing earlier brand-iteration assets. Area-downscaled
  for the web so they stay crisp at render size without being multi-MB.

## [0.1.1] - 2026-06-15

### Changed
- Simplified from a multi-section marketing homepage to a single-screen
  link-in-bio landing: animated rings around the circuit-R icon, the wordmark,
  and social chips: Twitch (`rakectl`), GitHub (`rake-pro`), Discord (`rake`,
  click-to-copy since Discord has no public profile URL), and email
  (`admin@rake.pro`). Removed the hero/capability-cards/CTAs.

## [0.1.0] - 2026-06-14

### Added
- Initial scaffold: self-contained Go (1.23) web server for the rake.pro
  homepage. stdlib `net/http` with method-based routing, zerolog structured
  logging, `embed.FS` for templates + static assets.
- `/healthz` + `/readyz` probes, env-driven config (`RAKEPRO_*`), graceful
  shutdown on SIGINT/SIGTERM.
- Multi-stage `Dockerfile` to `distroless/static:nonroot`; `Makefile` for
  run/build/test/docker/k8s; reference Kustomize manifests under `deploy/k8s/`.
- `.github/workflows/build-image.yml`: amd64 image build to private GHCR
  `ghcr.io/rake-pro/rakepro-web` on push to `main` and `v*` tags.
