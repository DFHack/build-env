#!/bin/bash -ex

docker build --pull -t proget.local.lubar.me/dfhack-docker/build-env:gcc-4.8 gcc48
docker push proget.local.lubar.me/dfhack-docker/build-env:gcc-4.8
docker build --pull -t proget.local.lubar.me/dfhack-docker/build-env:latest gcc7
docker push proget.local.lubar.me/dfhack-docker/build-env:latest
#docker build --pull -t proget.local.lubar.me/dfhack-docker/build-env:msvc msvc
#docker push proget.local.lubar.me/dfhack-docker/build-env:msvc
