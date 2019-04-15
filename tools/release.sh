#!/bin/bash

# Build a release
#
# Usage: VERSION=0.1.2 ./rel.sh [GOOS] [GOARCH]

set -e

export GOOS=${1:-linux}
export GOARCH=${2:-amd64}

[ -n "$VERSION" ] || (echo "Need VERSION" >&2; exit 1)

echo $VERSION

REL=sheens-$GOOS-$GOARCH-$VERSION
TARGET=`pwd`/rel/$REL
rm -rf $TARGET
mkdir -p $TARGET
for D in cmd/mexpect cmd/patmatch cmd/spectool cmd/mqclient cmd/siomq cmd/siostd; do
    echo $D
    # -ldflags "-X main.GitCommit=$$(git rev-list -1 HEAD) -X main.Version=$${VERSION:-NA}"
    (cd $D && go build -o $TARGET/$(basename $D) -ldflags="-s -w -X main.GitCommit=$(git rev-list -1 HEAD) -X main.Version=${VERSION:-NA}" )
done

go get -u github.com/bronze1man/yaml2json

(cd $GOPATH/src/github.com/bronze1man/yaml2json && go build -o $TARGET/yaml2json)

if [ "$GOOS" = "linux" -a "$GOARCH" = "amd64" ]; then
    (cd $TARGET && /usr/bin/upx *)
fi

mkdir $TARGET/js
cp js/*.js $TARGET/js

cp -R specs $TARGET/
cp LICENSE $TARGET/LICENSE.txt

echo "sheens $GOOS $GOARCH $VERSION $(git rev-list -1 HEAD)" > $TARGET/VERSION.txt
date --rfc-3339=ns -u >> $TARGET/VERSION.txt

echo "https://github.com/Comcast/sheens" > $TARGET/README.txt

if [ "$GOOS" = "windows" ]; then
    cd $TARGET/.. && zip -r $REL.zip $REL && ls -l $REL.zip
else
    cd $TARGET/.. && tar zcf $REL.tgz $REL && ls -l $REL.tgz
fi
