#!/bin/bash -ex

#nc=--no-cache

tag=proget.lubar.me/dfhack-docker/build-env
docker build $nc --pull -t $tag:gcc-4.8 gcc48
docker push $tag:gcc-4.8
docker build $nc --pull -t $tag:latest gcc7
docker push $tag:latest
docker build $nc --pull -t $tag:msvc msvc
docker push $tag:msvc
