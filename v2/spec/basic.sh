#!/bin/sh

_TRUE() {
    echo 1
    return 0
}
_FALSE() {
    echo 0
    return 1
}

NAME="Alexis"
echo "hello $NAME" >&2

A=10
B=10
C=$(( A + B ))
echo "$A $B $C" >&2

T=$(_TRUE)
F=$(_FALSE)
echo "$T $F $((T == F))" >&2

ADD() {
    echo $(( $1 + $2 )) >&2
}

RESULT=$( ADD 10 20 )
echo "result was: "$RESULT >&2

D=$(( 10 / 2 * (3 + 8) - 7 ))
echo $D >&2

A=$(( 5 + 10 ))
echo $A >&2
A=$(( 5 - 10 ))
echo $A >&2
A=$(( 5 * 10 ))
echo $A >&2
A=$(( 10 / 5 ))
echo $A >&2
A=$(( 10 % 5 ))
echo $A >&2
A=$(( ( 1 + 2 ) * 3 ))
echo $A >&2

A=$(( A + 1 ))
echo $A >&2
A=$(( A - 1 ))
echo $A >&2
A=$(( -A ))
echo $A >&2

A=$(( 5 & 0 ))
echo $A >&2
A=$(( 5 | 0 ))
echo $A >&2
A=$(( 5 ^ 0 ))
echo $A >&2
A=$(( ~5 ))
echo $A >&2
A=$(( 10 << 1 ))
echo $A >&2
A=$(( 5 >> 1 ))
echo $A >&2
