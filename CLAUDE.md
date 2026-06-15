# CLAUDE.md

Working memory for this repo. Read this first. Track outstanding work in
[backlog.md](backlog.md) and record shipped changes in [changelog.md](changelog.md).

## What this repo is

The **rake.pro landing page**: a single self-contained Go binary that serves one
branded, halo-themed page with social links (Twitch, GitHub, Discord, email). No
database, no backend, no marketing funnel - it is a link-in-bio landing while we
decide what rake.pro becomes long term. Static assets and HTML templates are
embedded into the binary (`embed.FS`), so the whole site ships as one artifact
with no runtime file dependencies.

## Stack

- **Go 1.23**, standard-library `net/http` with method-based routing. No web framework.
- **[zerolog](https://github.com/rs/zerolog)** for structured logging - the only dependency.
- `embed.FS` for templates + static assets (`web/embed.go`).
- Container: multi-stage build to `gcr.io/distroless/static-debian12:nonroot` (static, non-root, read-only rootfs).

## Layout

| Path | What it holds |
| --- | --- |
| `cmd/server/` | `main` - config load, signal handling, run loop |
| `internal/config/` | env-driven configuration (`RAKEPRO_*`) |
| `internal/server/` | routing, request-logging middleware, zerolog setup, server lifecycle |
| `web/` | `embed.go` + `templates/index.html` + `static/{css,js,img}` |
| `deploy/k8s/` | reference Kustomize manifests (NOT the live deploy - see Deployment) |
| `Dockerfile` | multi-stage -> distroless |
| `Makefile` | `run` / `build` / `test` / `docker-build` / `k8s-deploy` |
| `.github/workflows/build-image.yml` | GHCR image CI |

## Config (env vars)

All settings come from the environment (see `.env.example`). Key ones:
`RAKEPRO_ADDR` (`:8080`), `RAKEPRO_ENV` (`development` -> console logs, else JSON),
`RAKEPRO_LOG_LEVEL` (`info`). Timeouts: `RAKEPRO_{READ,WRITE,IDLE,SHUTDOWN}_TIMEOUT`.

## Endpoints

`GET /` (page), `GET /static/*` (embedded assets), `GET /healthz` + `GET /readyz`
(JSON probes for Docker/k8s).

## Build, run, release

- Local: `make run` (dev mode, console logs, :8080) or `go run ./cmd/server`.
- Image CI: `.github/workflows/build-image.yml` builds **amd64** and pushes to
  **private GHCR** `ghcr.io/rake-pro/rakepro-web` via the built-in `GITHUB_TOKEN`
  (no external registry secrets). Triggers: push to `main` (-> `latest` + `sha-`)
  and `v*` tags (-> semver). `VERSION` is injected into the binary via build-arg.
- **Release flow:** edit -> `go build` + run locally to verify -> commit -> tag
  `vX.Y.Z` (builds `ghcr.io/rake-pro/rakepro-web:X.Y.Z`) -> bump the chart's
  `tag`/`appVersion` in the GitOps repo (below).

## Deployment (lives in another repo)

The Helm chart that actually deploys this is in the **`Rake-Pro/GitOps-ArgoCD`**
repo at `cluster-apps/rakepro-web` - NOT here. It wraps the k8s-at-home `common`
library chart, pins the GHCR image tag, and pulls the private image via an
ExternalSecret (`ghcr-rakepro-web`, sourced from GSM key `ghcr-rakepro`). To ship
a new image: bump `image.tag` in that chart's `values.yaml` and `version`/
`appVersion` in its `Chart.yaml`, then sync in ArgoCD. The `deploy/k8s/` manifests
here are a standalone reference only.

## Branding assets

Sourced from the `branding` repo (`assets/png/`). Use the **clean** variants:
`icon-circuit-r` (icon) and `wordmark` (wordmark) - area-downscaled for web in
`web/static/img/`. Avoid the `*-unsc` (washed out) and `wordmark-transparent`
(grunge speckle) variants - they look low quality at render size.

## Conventions

- zerolog everywhere; no `fmt.Print`/`log` in request paths.
- Pin image tags downstream; the build also publishes `latest` for convenience only.
- amd64-only (the target cluster is amd64).
- Frontend is progressive-enhancement: the page works fully without JS (JS only
  adds halo parallax and Discord click-to-copy).
- Default git behavior: stage only; commit/push when explicitly asked.

## Known notes

- The Go **module path** is `github.com/rakepro/rakepro-web` while the GitHub repo
  is `Rake-Pro/rakepro-web`. Harmless for a private app binary (module path is
  internal, used in `-ldflags -X`), but update both if the repo is ever `go get`-ed.
- No test/staging site yet - production is the only target.
