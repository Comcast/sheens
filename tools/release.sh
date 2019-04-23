#!/bin/bash

# Build a release
#
# Usage: VERSION=0.1.2 ./release.sh [GOOS] [GOARCH]
#
# ToDo: Use Docker.

set -e

export GOOS=${1:-linux}
export GOARCH=${2:-amd64}

[ -n "$VERSION" ] || (echo "Need VERSION" >&2; exit 1)

echo $VERSION

REL=sheens-$GOOS-$GOARCH-$VERSION
TARGET=`pwd`/rel/$REL
rm -rf $TARGET
mkdir -p $TARGET/bin
for D in cmd/mexpect cmd/patmatch cmd/spectool cmd/mqclient cmd/sio; do
    echo $D
    (cd $D && go build -o $TARGET/bin/$(basename $D) -ldflags="-s -w -X main.GitCommit=$(git rev-list -1 HEAD) -X main.Version=${VERSION:-NA}" )
done

go get -u github.com/bronze1man/yaml2json

(cd $GOPATH/src/github.com/bronze1man/yaml2json && go build -o $TARGET/bin/yaml2json)

if [ "$GOOS" = "linux" -a "$GOARCH" = "amd64" ]; then
    (cd $TARGET/bin && /usr/bin/upx *)
fi

mkdir $TARGET/js
cp js/*.js $TARGET/js

cp -R specs $TARGET/
cp LICENSE $TARGET/LICENSE.txt

echo "sheens $GOOS $GOARCH $VERSION $(git rev-list -1 HEAD)" > $TARGET/VERSION.txt
date --rfc-3339=ns -u >> $TARGET/VERSION.txt
echo "https://github.com/Comcast/sheens" > $TARGET/VERSION.txt

echo "https://github.com/Comcast/sheens" > $TARGET/README.txt

if [ "$GOOS" = "windows" ]; then
    cd $TARGET/.. && zip -r $REL.zip $REL && ls -l $REL.zip
else
    cd $TARGET/.. && tar zcf $REL.tgz $REL && ls -l $REL.tgz
fi
