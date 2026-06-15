# Changelog

Notable changes to this repo. Newest first. Outstanding work is in
[backlog.md](backlog.md). Format loosely follows Keep a Changelog. Dates are
YYYY-MM-DD. Versions match the git tag / GHCR image tag.

## [Unreleased]

### Changed
- Release model -> `master`/`prod` branches. `master` is dev/default; merging a PR
  `master` -> `prod` builds and pushes `:latest` (+ `sha-<short>`), which ArgoCD
  Image Updater digest-pins onto the cluster. PRs targeting `prod` build without
  pushing (validation). `prod` is protected by a ruleset (PR + 1 approval +
  `build` check). `build-image.yml` triggers moved from `main` to `prod`.

## [0.1.2] - 2026-06-15

### Changed
- Swapped to the clean high-res branding assets: `icon-circuit-r` (icon) and
  `wordmark` (wordmark), replacing the washed-out `icon-unsc` and the
  grunge-speckled `wordmark-transparent`. Area-downscaled for the web so they
  stay crisp at render size without being multi-MB.

## [0.1.1] - 2026-06-15

### Changed
- Simplified from a multi-section marketing homepage to a single-screen,
  halo-themed link-in-bio landing: animated halo rings around the circuit-R
  icon, the wordmark, and social chips - Twitch (`rakectl`), GitHub (`rake-pro`),
  Discord (`rake`, click-to-copy since Discord has no public profile URL), and
  email (`admin@rake.pro`). Removed the hero/capability-cards/CTAs.

## [0.1.0] - 2026-06-14

### Added
- Initial scaffold: self-contained Go (1.23) web server for the rake.pro
  homepage. stdlib `net/http` with method-based routing, zerolog structured
  logging, `embed.FS` for templates + static assets.
- `/healthz` + `/readyz` probes, env-driven config (`RAKEPRO_*`), graceful
  shutdown on SIGINT/SIGTERM.
- Multi-stage `Dockerfile` to `distroless/static:nonroot`; `Makefile` for
  run/build/test/docker/k8s; reference Kustomize manifests under `deploy/k8s/`.
- `.github/workflows/build-image.yml` - amd64 image build to private GHCR
  `ghcr.io/rake-pro/rakepro-web` on push to `main` and `v*` tags.
