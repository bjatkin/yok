#!/bin/sh

# valid nesting
NAME=Jacob
echo "Length of name is: " ${#NAME} >&2

# requires command substitution
echo hello $(echo Alexis) >&2

# requires many levels of command substitution
echo "go bin:" $(echo $(ls -la $(which go)))

# nested paramater expansions
_TMP1=hello
_TMP2=${#_TMP1}
_TMP3=${#_TMP2}
echo ${#_TMP3} >&2
EXCLAIM="¿What! What! are you doing!?"
_TMP4=${EXCLAIM##¿}
echo ${_TMP4%%?} >&2