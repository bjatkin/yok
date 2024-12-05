#!/bin/sh

# these use commands will be checked before the script runs
if test -z "$(command -v git)"; then
    echo "ERROR: failed to find required command 'git'" >&2
    exit 255
fi

if test -z "$(command -v foo)"; then
    echo "ERROR: failed to find required command 'foo'" >&2
    exit 255
fi