#!/bin/bash -ex

docker build --pull -t proget.local.lubar.me/dfhack-docker/build-env:gcc-4.8 gcc48
docker build --pull -t proget.local.lubar.me/dfhack-docker/build-env:latest gcc7
