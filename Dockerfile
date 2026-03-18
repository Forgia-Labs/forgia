FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN go build -ldflags "-X github.com/Deepzima/forgia/cmd/forgia/cmd.Version=${VERSION}" -o /forgia ./cmd/forgia

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /forgia /forgia
ENTRYPOINT ["/forgia"]
