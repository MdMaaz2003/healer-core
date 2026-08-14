FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod ./
COPY cmd/ cmd/
RUN go build -o healer ./cmd/healer

FROM gcr.io/distroless/static-debian12
COPY --from=build /src/healer /healer
CMD ["/healer"]
