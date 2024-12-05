NAME=$(curl localhost:8080/get_name)
echo "name is '$NAME'" >&2

NAME=$(curl localhost:8080/get_name)
E=$?
if [ E != 0 ]; then
    echo "command failed with code $E" >&2
    NAME="Jay"
fi
echo "now name is" $NAME >&2

NAME=$(curl localhost:8080/get_name || echo "Lex")
echo "default name is" $NAME >&2
