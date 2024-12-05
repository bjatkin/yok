add () {
    echo $(( $1 + $2 ))
}

div () {
    if [ $2 -eq 0 ]; then
        return 1
    fi

    echo $(( $1 / $2 ))
}

print_multi () {
    for i in $(seq 0 $2); do
        echo -n "$1" >&2
    done

    # it's important that all other commands in the function that might write to stdout
    # instead write to stderr or are silenced entirely, if not the return value will be corrupted
    echo "ok"
}

echo "$(add 10 20)" >&2

echo "$(div 10 5 || echo 0)" >&2
echo "$(div 10 0 || echo 0)" >&2

echo $(print_multi "hello " 3) >&2
