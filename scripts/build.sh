#!/bin/bash
#
# This script is used to build .deb & .rpm packages using docker and a docker image
#
# Requirements:
#   - Working docker command
#   - Internet access to pull docker images
#

NAME="vindalu"
SCM_PATH="github.com/vindalu/${NAME}"

VERSION=$(cat etc/vindalu.json.sample | grep version | cut -d ' ' -f 6 | sed "s/\"//g")
# Image being created
DOCKER_NAME="vindalu/${NAME}:${VERSION}"
# Image to build the code
BLD_DOCKER_IMAGE="golang:1.4.3"
# Image to build package 
PKG_DOCKER_IMAGE="euforia/fpm"

# Check for docker binary
which docker > /dev/null || {
    echo "'docker' command not found!";
    exit 2;
}

# Build phase
cat <<BLDCODE

Building
--------

BLDCODE
docker run --rm -v "$PWD":/go/src/${SCM_PATH} \
    -w /go/src/${SCM_PATH} ${BLD_DOCKER_IMAGE}  make all || {
    
    echo "Build failed"
    exit 3
}

# Packaging phase
cat <<PKG

Packaging
---------

PKG
# Build packages
docker run --rm -v "$PWD":/usr/src/${NAME} -w /usr/src/${NAME} $PKG_DOCKER_IMAGE make .packages || {
    echo "Package build failed"
    exit 4
}

# Dockerizing phase
cat <<DKRIMG

Dockerizing
-----------

DKRIMG
# Build docker image
docker build --rm --force-rm -t $DOCKER_NAME . || {
    echo "Docker image build failed"
    exit 5
}

# Final check
cat <<SMR

Summary
-------

SMR
docker run --rm -t $DOCKER_NAME /opt/${NAME}/bin/${NAME} version && echo -e "\nSuccess!\n" || {
    echo -e "\nFailed!\n"
    exit 5
}
