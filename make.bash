#!/bin/bash -ex

tag=proget.lubar.me/dfhack-docker/build-env
docker build --pull -t $tag:gcc-4.8 gcc48
docker push $tag:gcc-4.8
docker build --pull -t $tag:latest gcc7
docker push $tag:latest
docker build --pull -t $tag:msvc msvc
docker push $tag:msvc
