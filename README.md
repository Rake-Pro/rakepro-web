# rakepro-web

The landing page for [rake.pro](https://rake.pro) - a small, self-contained Go
web server serving a single branded page with social links. Static assets and
HTML templates are embedded into the binary, so the whole site ships as a single
artifact with no runtime file dependencies.

## Stack

- **Go 1.23**, standard-library `net/http` with method-based routing (no web framework)
- **[zerolog](https://github.com/rs/zerolog)** for structured logging (the only dependency)
- Assets embedded via `embed.FS`
- Container: multi-stage build to a distroless, non-root, static image
- Deploy: plain Docker or Kubernetes (reference Kustomize manifests under
  `deploy/k8s` - adapt to your own cluster/deploy tooling)

## Layout

```
cmd/server/         main entrypoint (config load, signal handling, run)
internal/config/    env-driven configuration
internal/server/    routing, middleware, zerolog setup, lifecycle
web/                embedded templates + static assets (css/js/img)
deploy/k8s/         Deployment, Service, Ingress, Kustomization
Dockerfile          multi-stage build -> distroless
Makefile            run / build / test / docker / k8s targets
```

## Run locally

```
make run            # development mode, console logging, :8080
# or
go run ./cmd/server
```

Then open http://localhost:8080.

First build only:

```
make tidy           # resolves go.sum (no network-pinned sums are committed yet)
```

## Configuration

All settings come from the environment (see `.env.example`):

| Variable                   | Default        | Purpose                                  |
|----------------------------|----------------|------------------------------------------|
| `RAKEPRO_ADDR`             | `:8080`        | Bind address                             |
| `RAKEPRO_ENV`              | `development`  | `development` uses console logs; else JSON |
| `RAKEPRO_LOG_LEVEL`        | `info`         | `trace`..`error`                         |
| `RAKEPRO_READ_TIMEOUT`     | `5s`           | Request read timeout                     |
| `RAKEPRO_WRITE_TIMEOUT`    | `10s`          | Response write timeout                   |
| `RAKEPRO_IDLE_TIMEOUT`     | `120s`         | Keep-alive idle timeout                  |
| `RAKEPRO_SHUTDOWN_TIMEOUT` | `15s`          | Graceful shutdown budget                 |

## Endpoints

- `GET /` - homepage
- `GET /static/*` - embedded CSS/JS/images
- `GET /healthz` - liveness (JSON)
- `GET /readyz` - readiness (JSON)

## Docker

```
make docker-build           # rakepro/rakepro-web:<version>
make docker-run             # -> http://localhost:8080
```

## Kubernetes

```
make k8s-deploy             # kubectl apply -k deploy/k8s
```

Edit `deploy/k8s/ingress.yaml` (host, ingressClassName, TLS issuer) and pin the
image tag in `deploy/k8s/kustomization.yaml` for your cluster.

## Notes

- The Go module path is `github.com/rakepro/rakepro-web`. If you host it
  elsewhere, update the path in `go.mod` and the imports.
- Branding assets in `web/static/img` are the circuit-R icon, avatar, and
  wordmark PNGs. No webfonts - the type stack is system-ui/Inter (OS default).

## License

Code is licensed under the [MIT License](LICENSE). The Rake-Pro name, logo,
and brand assets under `web/static/img/` are identity marks and are not
covered by the MIT license or offered for reuse.
