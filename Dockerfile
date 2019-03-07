FROM golang:1.12-alpine
RUN apk add --no-cache git
WORKDIR /go/src/golinks-daemon
ENV GO111MODULE=on
COPY . .
RUN go mod download
RUN go install -v ./...
CMD ["golinks-daemon"]