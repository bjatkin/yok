#!/bin/sh

GREET="hello world"
GREET_LEN=${#GREET}
GREET_FRIEND=${GREET/world/friend}
STRANGE_GREETING=${GREET//l/i}
PLACE=${GREET##"hello "}
SHORT_GREET=${GREET%%" world"}