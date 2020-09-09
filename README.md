# golinks-daemon
[![govice](https://circleci.com/gh/govice/golinks-daemon.svg?style=svg)](https://circleci.com/gh/govice/golinks-daemon)

This is the daemon for the GoVice project. This is currently under development, and you should expect breaking changes. The goal of this project is to produce a blockchain-backed merkle tree used to track the integrity of filesystem(s) over time.

## Usage
```
go install
golinks-daemon
```

## Configuration

See [config.json](/etc/config.json) for an example.

## Docker
```
docker build -t golinksd:latest
docker run -it golinksd -e GOLINKSD_EMAIL=****** -e GOLINKSD_PASSWORD=******
```