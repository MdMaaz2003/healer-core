FROM golang:1.23 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN go build -o healer ./cmd/healer

FROM gcr.io/distroless/static-debian12
COPY --from=builder /src/healer /healer
ENTRYPOINT ["/healer"]
