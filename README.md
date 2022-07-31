# aqua-docker

Install tools in Docker image

## :warning: Deprecated

Please see the document.

https://aquaproj.github.io/docs/tutorial-extras/build-container-image

## Overview

aqua-docker is a CLI installing CLIs in Docker images.

## :warning: Restriction

aqua-docker assumes that installed CLIs work as a single executable file.
So some tools such as [tfenv](https://github.com/tfutils/tfenv) can't be supported.

## Requirements

* Docker

## How to use

Please run aqua-docker by `go run` command in Dockerfile.

e.g. Install actionlint and reviewdog with aqua-docker.

aqua.yaml

```yaml
registries:
- type: standard
  ref: v3.18.0 # renovate: depName=aquaproj/aqua-registry
packages:
- name: rhysd/actionlint@v1.6.15
- name: reviewdog/reviewdog@v0.14.1
```

Dockerfile

```dockerfile
FROM golang:1.18.4 AS aqua
COPY aqua.yaml /aqua.yaml
RUN go run github.com/aquaproj/aqua-docker@v0.1.0 --aqua-version v1.17.1 --config /aqua.yaml --dest /dist actionlint reviewdog

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=aqua /dist/* /usr/local/bin/
```

Build an image.

```console
$ docker build -t foo .
```

Let's confirm if tools are installed.

```console
$ docker run --rm -ti foo sh
/ # reviewdog -version
0.14.1
/ # actionlint -version
1.6.15
installed by downloading from release page
built with go1.18.3 compiler for linux/arm64
/ # which reviewdog
/usr/local/bin/reviewdog
/ # which actionlint
/usr/local/bin/actionlint
/ # which aqua # aqua isn't installed
/ # 
```

## How does it works?

1. Install aqua
1. Run `aqua i`
1. Create a directory (By default, the directory path is `dist`, but you can change this by `-dest` option) and copy files at the directory

## LICENSE

[MIT](LICENSE)
