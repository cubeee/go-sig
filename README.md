# go-sig

[![Build Status][build-status-img]][build-status]

Dynamically generated and updated skill goal signatures for RuneScape players.

## Building the Docker image
```
docker build -t go-sig .
```

## Running locally
You can easily run go-sig locally with [Docker compose][docker-compose], you just have to make your own copy of ``docker-compose.yml`` from the provided template file and then use the following commands.
```
# Run
docker-compose up

# Destroy
docker-compose down
```

## Environment variables
Name | Action | Default
:---: | --- | --- |
IMG_PATH | Path to the directory where you want to store the generated images | images/
PROCS | Number of operating system threads you want to give for `go-sig` | `runtime.NumCPU()`
DISABLE_LOGGING | Use `true` or `1` to disable output from `log` | false
ENABLE_DEBUG | Use `true` or `1` to map routes to `pprof` urls | false
AES_KEY | The key used to encrypt and decrypt hidden usernames in signature urls | ""
VIRTUAL_HOST | The url displayed on generated signature result page | sig.scapelog.com
SECURE | Use `true` for `https` and `false` for `http` to be used in links | true

[build-status-img]: https://travis-ci.org/cubeee/go-sig.svg
[build-status]: https://travis-ci.org/cubeee/go-sig
[docker-compose]: https://docs.docker.com/compose/