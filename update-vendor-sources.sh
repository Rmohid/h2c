#!/bin/bash

##########################################################
# download all dependencies to tmp dir
##########################################################

TMP_DIR=$(mktemp -d)
export ORIG_GOPATH="$GOPATH"
export GOPATH="$TMP_DIR"
unset GO15VENDOREXPERIMENT

go get github.com/rmohid/h2d
# go get golang.org/x/net/http2/hpack
# go get github.com/fatih/color

export GOPATH="$ORIG_GOPATH"
export GO15VENDOREXPERIMENT=1

##########################################################
# replace content of vendor dir with download
##########################################################

find "$TMP_DIR/src" -name '.git' | while read dir ; do rm -rf "$dir" ; done
rm -rf "$TMP_DIR/src/github.com/rmohid/h2d"
rm -rf "$GOPATH/src/github.com/rmohid/h2d/vendor"
mkdir "$GOPATH/src/github.com/rmohid/h2d/vendor"
mv "$TMP_DIR"/src/* "$GOPATH/src/github.com/rmohid/h2d/vendor"

echo LAST UPDATE: `date` > "$GOPATH/LAST_UPDATE.txt"

##########################################################
# build in Docker container w/o network connection
# to make sure all dependencies are included.
##########################################################

BUILD_SCRIPT="
    mkdir -p /go/src/github.com/fstab &&
    mv /tmp/h2d /go/src/github.com/fstab &&
    export GO15VENDOREXPERIMENT=1 &&
    go install github.com/rmohid/h2d &&
    echo build successful &&
    /go/bin/h2d version
"

container_id=$(docker create --net=none -i -t golang bash -c "$BUILD_SCRIPT")
docker cp $GOPATH/src/github.com/rmohid/h2d $container_id:/tmp
docker start -a $container_id
docker rm $container_id > /dev/null
