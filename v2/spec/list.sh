#!/bin/sh

_true=1
_false=0

str_list=("a" "b" "c")
int_list=(0 1 2 3 4)
bool_list=($_false $_true $_false)

mixed_list=("one" 2 $_false "last elem")

# lists have the type 'list'
empty_list=()

# loop over all the elements in the mixed_list
for (( i=0; i<${#mixed_list[@]}; i++ )); do
    echo "elem: $i is ${mixed_list[$i]}" >&2
done
