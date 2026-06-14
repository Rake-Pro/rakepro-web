# --- build stage ---------------------------------------------------------
FROM golang:1.23-alpine AS build

WORKDIR /src

# Cache module downloads separately from the source for faster rebuilds.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

ARG VERSION=dev
# Static, stripped binary. CGO is off so it runs in a scratch/distroless image.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X github.com/rakepro/rakepro-web/internal/server.version=${VERSION}" \
    -o /out/rakepro-web ./cmd/server

# --- runtime stage -------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=build /out/rakepro-web /rakepro-web

EXPOSE 8080
USER nonroot:nonroot

ENV RAKEPRO_ADDR=":8080" \
    RAKEPRO_ENV="production" \
    RAKEPRO_LOG_LEVEL="info"

ENTRYPOINT ["/rakepro-web"]
