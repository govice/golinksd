FROM golang:1.12-alpine AS daemon_base
RUN apk add --no-cache git bash
WORKDIR /go/src/github.com/govice/golinks-daemon

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

FROM daemon_base AS daemon_build
COPY . .
RUN go install

FROM alpine
COPY --from=daemon_build /go/bin/golinks-daemon /bin/golinks-daemon
ENV TEMPLATES_HOME="/templates"
COPY templates /templates

ENTRYPOINT ["/bin/golinks-daemon"]