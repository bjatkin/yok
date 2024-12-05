curl -X POST localhost:8080/get_name 2> /dev/null

cat examples/data/test.txt > examples/data/new_test.txt

sort < examples/data/test.txt
echo "---" >&2

sort << EOF
a
c
d
b
EOF

sort < examples/data/test.txt > examples/data/sorted.txt
