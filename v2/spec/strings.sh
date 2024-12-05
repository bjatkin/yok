FIRST_NAME="Alexis"
LAST_NAME="Smith"
FULL_NAME=$FIRST_NAME$LAST_NAME
echo "full name: " $FULL_NAME >&2

NAME_LEN=${#FULL_NAME}
echo "name len: " $NAME_LEN >&2

GREET="Hello"
SUB_STRING=${GREET:0:4}

GREET_WORLD="$GREET World"
echo $GREET_WORLD >&2

GREET_LEX=${GREET_WORLD/World/Lex}
echo $GREET_LEX >&2

CHEER="Jay Jay Jay!"
echo $CHEER >&2
NEW_CHEER=${CHEER//Jay/Lex}
echo $NEW_CHEER >&2

if [ "$FIRST_NAME" = "Alexis" ]; then
    echo "Name is Lex" >&2
fi

TEST=10
if [ $TEST = 10 ]; then
    echo "10 is 10" >&2
fi

if [ -z "$LAST_NAME" ]; then
    echo "No last name" >&2
fi

case "$GREET_WORLD" in
    (*"lo"*) echo "Contains 'lo'" >&2;;
esac
