# golang:1.25-alpine — pinned 2026-03-21
FROM golang:1.25-alpine@sha256:450ce2460f20b2f581cf1ac4f36606f88817e63f907f489d9a3b1c14ba821979 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN go build -ldflags "-X github.com/Deepzima/forgia/cmd/forgia/cmd.Version=${VERSION}" -o /forgia ./cmd/forgia

# gcr.io/distroless/static:nonroot — pinned 2026-03-21
FROM gcr.io/distroless/static:nonroot@sha256:64c43684e6d2b581d1eb362ea47b6a4defee6a9cac5f7ebbda3daa67e8c9b8e6
COPY --from=builder /forgia /forgia
ENTRYPOINT ["/forgia"]
