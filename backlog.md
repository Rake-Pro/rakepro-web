# Backlog

Outstanding / future work for this repo. Started tracking 2026-06-15.
Conventions: see [CLAUDE.md](CLAUDE.md). Shipped items move to
[changelog.md](changelog.md).

Priority key: P1 = should do soon, P2 = planned, P3 = nice-to-have.

## Deploy / ops

- [ ] **First production rollout.** Image `0.1.2` is built and pinned;
  bringing it live is gated on the GHCR pull secret. Confirm the cluster's
  image pull secret covers this image, deploy via the private GitOps repo,
  then point the reverse proxy at the ingress. (P1)
- [ ] **Stand up a test/staging site** so changes can be eyeballed off a real URL
  before production (none exists today; previews are ad-hoc). (P2)

## App / frontend

- [ ] **Discord link.** Currently click-to-copy of the handle `rake` (Discord has
  no canonical profile URL). If/when there is a server, swap to a real invite
  link. (P3)
- [ ] Add `apple-touch-icon` + a small `favicon.ico` fallback for older clients. (P3)
- [ ] Decide the long-term direction for rake.pro - this landing is a placeholder. (P2)

## Engineering

- [ ] **No tests yet.** Add at least a handler smoke test (homepage 200, probes,
  static asset served) and wire `go test ./...` into CI before the image build. (P2)
- [ ] Resolve the module-path vs repo-name mismatch (`github.com/rakepro/...` vs
  `Rake-Pro/rakepro-web`) if the repo is ever consumed via `go get`. (P3)
- [ ] Renovate/dep updates for the single `zerolog` dependency and the Go base
  image (optional - surface area is tiny). (P3)
