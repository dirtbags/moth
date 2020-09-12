FROM golang:1.13 AS builder
COPY go.* /src/
COPY pkg /src/pkg/
COPY cmd /src/cmd/
WORKDIR /src/
#RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go install -i -a -ldflags '-extldflags "-static"' ./...

FROM scratch
COPY --from=builder /go/bin/* /usr/bin/
COPY theme /theme/
COPY example-puzzles /puzzles/
COPY LICENSE.md /LICENSE.md

ENTRYPOINT [ "/usr/bin/mothd" ]
