#!/bin/sh

X=10
Y=20
Z=30

if [ "$X" -gt 0 ]; then
    echo "x is positive" >&2
    if [ "$Y" = 20 ]; then
        echo "y is 20" >&2
        if [ "$Z" != "$X" ]; then
            echo "z does not equal x" >&2
        fi
    fi
else
    echo "x is negative or zero" >&2
    if [ "$Y" != 20 ]; then
        echo "y is not 20" >&2
    else
        echo "y is still 20" >&2
    fi
fi

if [ "$X" -lt 0 ]; then
    echo "x is negative" >&2
elif [ "$X" -gt 1 ]; then
    echo "x is negative" >&2
elif [ "$X" = 1 ]; then
    echo "x is negative" >&2
else
    echo "x is zero" >&2
fi