FROM golang:1.23 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o healer ./cmd/healer

# hadolint ignore=DL3007
FROM gcr.io/distroless/static-debian12:latest
COPY --from=builder /src/healer /healer
ENTRYPOINT ["/healer"]
