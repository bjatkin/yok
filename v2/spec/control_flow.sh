#!/bin/sh

if [ 10 -lt 100 ]; then
    echo "all is well" >&2
fi

if [ 10 -lt 100 ]; then
    echo "all is well" >&2
else
    echo "something is wrong" >&2
fi

a=20
b=30
if [ $a -gt 10 ] && [ $b -gt 5 ]; then
    echo "both checks are true" >&2
else
    echo "a or b is false" >&2
fi

case $a in
    (0)
        echo "a is 0" >&2
        ;;
    (1)
        echo "a is 1" >&2
        ;;
    (2)
        echo "a is 2" >&2
        ;;
    (5|10|15|20)
        echo "a is 5, 10, 15 or 20" >&2
        ;;
    (*)
        echo "a is unknown" >&2
        ;;
esac

c="test"
case $c in
    (ok)
        echo "c is ok" >&2
        ;;
    (test)
        echo "c is test" >&2
        ;;
esac

for i in $(seq 0 10); do
    echo "i: " $i >&2
done

for i in $(seq 2 2 12); do
    echo "even: " $i >&2
done

j=0
while [ $j -le 10 ]; do
    echo "j: " $j >&2
    j=$((j + 1))
done
