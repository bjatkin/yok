cat ./examples/data/test.txt | grep "a" | sort

ls -l | wc -l

greet() {
    NAME=""
    while read NAME; do
        echo "$1 $NAME" >&2
    done
}

echo -e "John\nJacob\nJingleheimer\nSchmidt" | greet "Hello"
echo "---" >&2
echo -e "John\nJacob\nJingleheimer\nSchmidt" | greet "Guten Tag"
