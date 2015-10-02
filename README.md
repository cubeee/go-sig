# go-sig

[![Build Status][build-status-img]][build-status]

Dynamically generated and updated skill goal signatures for RuneScape players

## Installation
    # Install gb, the build tool
    $ go get github.com/constabulary/gb/...

    # Install dependencies from dependencies.txt
    $ make deps

    # Build go-sig
    $ make build

    # Run go-sig
    $ bin/signature

## Deploying
    # Build .deb package
    $ make package

    # Transfer build/go-sig.deb to your server
    ...

    # Install .deb on your server
    remote> sudo dpkg -i go-sig.deb

## Environment variables
Name | Action | Default
:---: | --- | --- |
IMG_PATH | Path to the directory where you want to store the generated images | images/
PROCS | Number of operating system threads you want to give for `go-sig` | `runtime.NumCPU()`
DISABLE_LOGGING | Use `true` or `1` to disable output from `log` | false
ENABLE_DEBUG | Use `true` or `1` to map routes to `pprof` urls | false

[build-status-img]: https://travis-ci.org/cubeee/go-sig.svg
[build-status]: https://travis-ci.org/cubeee/go-sig
[gb-site]: http://getgb.io/
