#!/bin/bash

# Usage: VARNAME INFILENAME OUTFILENAME
#
# Writes OUTFILE that is a C file that initializes char* VARNAME as
# the source of INFILENAME.

set -e

VARNAME="$1"
INFILENAME="$2"
OUTFILENAME="$3"

[ ! -e "$VARNAME" ] || (echo "File $VARNAME exists" >&2; exit 1)

cp "$INFILENAME" "$VARNAME"src
xxd -i "$VARNAME"src "$OUTFILENAME"
rm "$VARNAME"src

echo "unsigned char* $VARNAME() { return $VARNAME"src"; } " >> $OUTFILENAME
