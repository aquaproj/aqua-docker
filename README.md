# aqua-docker

Install tools in Docker image

## Overview

aqua-docker is a CLI installing CLIs in Docker images.

## :warning: Restriction

aqua-docker assumes that installed CLIs work as a single executable file.
So some tools such as [tfenv](https://github.com/tfutils/tfenv) can be supported.

## Requirements

* Docker

## How to use

Please run aqua-docker by `go run` command in Dockerfile.

e.g.

```dockerfile
FROM golang:1.18.4 AS aqua
COPY aqua.yaml /aqua.yaml
RUN go run github.com/aquaproj/aqua-docker@v0.1.0 --aqua-version v1.17.1 --config /aqua.yaml --dest /dist golangci-lint actionlint reviewdog

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=aqua /dist/* /usr/local/bin/
```

## How does it works?

1. Install aqua
1. Run `aqua i`
1. Create a directory (By default, the directory path is `dist`, but you can change this by `-dest` option) and copy files at the directory

## LICENSE

[MIT](LICENSE)
