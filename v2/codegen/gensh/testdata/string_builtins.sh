#!/bin/sh

GREET="hello world"
GREET_LEN=${#GREET}
PLACE=${GREET##"hello "}
echo $PLACE >&2
SHORT_GREET=${GREET%%" world"}
echo $SHORT_GREET >&2

# use literal instead of identifier
_TMP1="new york"
STATE_LEN=${#_TMP1}

# use call instead of identifier
_TMP2=$(echo "new mexico")
STATE_LEN=${#_TMP2}

# use identifiers for remove fix
T=test
I=ing
_TMP3=testing
echo ${_TMP3##$(echo -n $T)} >&2
_TMP4=testing
echo ${_TMP4%%$(echo -n $I)} >&2