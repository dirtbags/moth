FROM golang:1.13 AS builder
COPY go.* /src/
COPY pkg /src/pkg/
COPY cmd /src/cmd/
COPY theme /target/theme/
COPY example-puzzles /target/puzzles/
COPY LICENSE.md /target/
WORKDIR /src/
RUN CGO_ENABLED=0 GOOS=linux go install -i -a -ldflags '-extldflags "-static"' ./...
RUN mkdir -p /target/bin
RUN cp /go/bin/* /target/bin/

FROM builder AS tester
RUN go test ./...

FROM scratch
COPY --from=builder /target /

ENTRYPOINT [ "/bin/mothd" ]
