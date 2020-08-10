FROM golang:alpine AS build
RUN apk add --no-cache git bash
WORKDIR /go/src/github.com/govice/golinks-daemon
ENV GOPRIVATE="github.com/libp2p/*"
COPY go.mod .
COPY go.sum .
RUN go mod download -x

FROM golang:alpine
WORKDIR /go/src/github.com/govice/golinks-daemon
COPY --from=build /go/pkg /go/pkg
ENV TEMPLATES_HOME="/templates"
COPY . .
RUN go install -v .

ENTRYPOINT ["golinks-daemon"]