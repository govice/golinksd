FROM golang:alpine AS build
WORKDIR /go/src/github.com/govice/golinksd
COPY go.mod .
COPY go.sum .
RUN go mod download -x

FROM golang:alpine
RUN apk add --no-cache git bash
WORKDIR /go/src/github.com/govice/golinksd
EXPOSE 8082
COPY --from=build /go/pkg /go/pkg
ENV GOLINKSD_TEMPLATES_HOME="/root/.golinksd/templates"
ENV GIN_MODE=release
COPY ./templates /root/.golinksd/templates
COPY ./etc/config.json /root/.golinksd/
COPY *.go ./
COPY go.mod .
COPY go.sum .
COPY ./bin/entrypoint.sh /usr/local/bin
RUN chmod +x /usr/local/bin/entrypoint.sh
RUN go install -v .
CMD ["golinks-daemon"]
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
