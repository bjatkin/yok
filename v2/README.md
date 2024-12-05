# Yо̄k

**WARNING:** **Yо̄k** is currently under heavy development and *everything* is subject to change.

**Yо̄k** aims to be the golden goodness hidden inside the shell.
It is a programming language that compiles to **sh**, offering a more modern and convenient way to write and manage shell scripts.
**Yо̄k** takes inspiration from many modern languages like [Go](https://go.dev/), [Python](https://www.python.org/), [Elixir](https://elixir-lang.org/), and [Zig](https://ziglang.org/) all while providing access to the unique language features of **sh**.
In fact, programmers familiar with **sh** will likely be able to intuit how any given **Yо̄k** programs will be represented once transpiled. 

## Quick Start

1. Installation:

Clone the repository

```sh
$ git clone https://github.com/your-username/yok
```

Build the transpiler

```sh
$ go build .
```

2. Create a new Yо̄k file

```yok
# say hello
print("Hello, world!")
```

3. Transpile your script

```sh
$ yok build hello.yok
```

This will generate a `hello.sh` file.

```sh
#!/bin/sh

# say hello
echo "Hello, world!" >&2
```

4. You can now execute the generated sh script

```sh
$ ./hello.sh
```

## Goals and Philosophy

* **Readability and Maintainability**: **Yо̄k** aims to provide a more modern and intuitive syntax than traditional shell scripting. 
    This includes reducing shell scripts reliance on operator foo and making common patterns (e.g. error checking) more obvious.

* **Compatibility**: despite it's rough edges **sh** is still a ubiquitous tool.
    Replacing shell scripts has never been the goal of this project.
    Rather we intend to make shipping **sh** faster, easier, and less bug prone.
    That's why **Yо̄k** complies to clear, idiomatic **sh** code.

* **Developer Friendly**: **Yо̄k** intends to be a modern scripting language and so supports features that have become standard for modern languages.
    This includes a built in `fmt` command for standard language formatting,
    a `test` command with first class support for testing,
    and a modern `macro` system.

## Yо̄k Features and Syntax

### Values

**Yо̄k** primarily treats values as strings, just like in **sh**.
However it introduces the concept of `atoms` as an alternative way to represent literals.
**Yо̄k** supports normal strings just like you would expect:

```yok
let normal = "normal string"
```

Additionally **Yо̄k** `atoms` can be defined by prefixing a string with a `:`.
Importantly these strings can not contain spaces.
In **Yо̄k** `atoms` are often used to represent integer literals, though any string is valid
It is important to know that the type of an `atoms` is `string`.

```yok
let a = :10
let b = :20
let status = :ok
```

As long as the file paths do not contain spaces you can also use atoms to represent file paths.

```yok
let my_file = :/my/file.txt
let my_dir = :my/relative/dir
```

In **Yо̄k** you can also use triple quotes (""") to specify multiline string values.

```yok
let his_name = """John
Jacob
Jingleheimer
Schmidt
"""
let my_name = his_name
```

**Yо̄k** also supports single quoted strings.
However these are only ever in switch statements for string pattern matching 
You can learn more in the [Control Flow](#control-flow) section.

**BUT WHY?:** It may seem odd for **Yо̄k** to eschew integer or boolean types.
This is because in **sh** support for integer literals is actually illusory.
Integer literals are actual strings.

```sh
a=10 # <- this is actually the string "10"
```

In fact in **sh** pretty much everything can be thought of as being a string.

```sh
a=hello # <- this is a string
b=42    # <- so is this
c=true  # <- and this
d=3.14  # <- this is a string too

echo a "world" # <- this prints `a world` instead of `hello world` because `a` is a string, not the variable `a`
echo $a "world" # <- this is how to actually print `hello world`
```

This is surprising for many developers and can be a source of unexpected behavior.
**Yо̄k** attempts to make this behavior more explicit and clear.
This is of course at the expense of some slightly clunky syntax but we feel this is a reasonable tradeoff.

### Variables

Variables must be declared with `let` before they can be used.
This prevents some bugs including instances where misspelled variables silently resolve to empty values.

```yok
let x = :42
let y = "hello"
y = :world

let verbs = "run, jump, skip"
# this is a compile time error error because `verb` is not declare and should actually be `verbs`
print("I like to do the following:", verb)
```

Variables can also be set in the parent environment (i.e. exported), using the `super` keyword

```yok
# this will be set in the parent environment
super home = :/usr/me
```

### Integer Math

While **Yо̄k** does not support typed integers, it has several operators that can be used to do integer calculations.
These tools take strings as input, convert those strings to integers, and then return strings as output.

```yok
# simple mathematical operations
a = :5 + :10
a = :5 - :10
a = :5 * :10
a = :10 / :5
a = :10 % :5
a = ( :1 + :2 ) * :3

# unary add, minus, and negation 
a++
a--
a = -a
```

**Note:** Floating point math is not supported natively in **Yо̄k** (yet :D), but you can leverage tools like `bc` or `awk` to make it possible.

### String Operations

Substrings can be created using string slices.

```yok
let hello = "hello world"
let greet = hello[:5]
let place = hello[6:]
let mid = hello[2:5]
```

**Yо̄k** does not support string concatenations like some languages.
Instead all string combinations should be done using format strings.

```yok
let fiz = "fiz"
let buzz = "buzz"
let fiz_buzz = "{fiz}{buzz}"
```

You can also get the length of a string by using the `len` builtin function

```yok
let dog_breed = "Dalmatian"
let breed_len = len(dog)
```

There are also the `replace` and `replace_all` builtin functions to replace substrings in a larger string.

```yok
let cheer = "hip hip hooray"
print(cheer) # prints "hip hip hooray"

cheer = replace(cheer, "hooray", "hoora")
print(cheer) # print "hip hip hoora"

cheer = replace_all(cheer, "hip", "hoop")
print(cheer, "hoop hoop hoora")
```

### Control Flow

**Yо̄k** supports all the same control flow constructs that **sh** provides.
This includes all the expected `if` variants:

```yok
if x > 0 {
    print("x is positive")
}

if x == 0 {
    print("x is zero")
} else {
    print("x is not zero")
}

if x < 0 {
    print("x is negative)
} else if x > 0 {
    print("x is positive")
} else {
    print("x is zero")
}

if x > 0 and y > 0 {
    print("x and y are positive")
}

if x < 0 or y < 0 {
    print("x or y is negative)
}
```

`switch` statements are also supported and include support for **sh** style pattern matching:

```yok
switch "hello" {
    "hello"   { print("hello how are you") }
    "goodbye" { print("see you later") }
}

let a = "hello world"
switch a {
    # sh style string pattern matching is supported
    '*friend' { print("hello to a friend") }
    '*world'  { print("hello to the world") }
}
```

as well as `for` and `while` loops:

```yok
for i in range(:1, :10) {
    print("i is ", i)
}

# only the value :true is truthy, all other values are falsy
while :true {
    print("loop forever")
}
```

### Comparison Operators

Control flow relies on the use of comparison operators.
**Yо̄k** supports all the basic comparison operators you would expect.

```yok
let x = :10
let y = :20

if x == y {
    print("x == y")
}

if x != y {
    print("x != y")
}

if x > y {
    print("x > y")
}

if x < y {
    print("x < y")
}

if x >= y {
    print("x >= y")
}

if x <= y {
    print("x <= y")
}
```

**Warning:** Comparison operations are *statements* in **Yо̄k**, not expressions.
This means they do *not* return a value.
Instead they work by setting the `error code`.
Trying to use a comparison as a value will result in a compile time error.

```yok
let age = 19

# this fails because `age > 16` does not return a value and so can not be assigned to a variable
let can_drive = age > 16
```

### Functions

**Yо̄k** functions are declared with the `fn` keyword.
They can take input parameters and return a value.

```yok
fn add(a, b) {
    return a + b
}
```

Functions behave like commands so they can also read from `stdin` and set the `error code`

```yok
fn div(a, b) {
    if b == 0 {
        # set the error code to 1 on return
        return :0, :1
    }
    # no status code is specified so it defaults to 0
    return a / b
}
```

### Commands

**Yо̄k** treats commands and function calls in the same way.
Commands from the environment must be explicitly imported with `use` at the top of your script.
These commands can then be called just like functions.

```yok
use {
    curl
}

curl("-X=POST", "localhost:8000/")
```

The content that these commands send to `stdout` can be "captured" and placed in a variable.

```yok
use {
    seq
}

let sequence = seq(:1, :10)
print(sequence) # this will print the numbers from 1 to 10
```

### Stdout, Stdin and Stderr

`stdin`, `stdout`, `stderr` can be manipulated just like in `sh`.

For example, you can send data from a file into a command using the named `stdin` argument in a command or function.

```yok
# take the test.txt file descriptor and set it to greps `stdin` file descriptor
grep("test", stdin=:test.txt)
```

You can also pipe a string directly into `stdin` with the `<=` syntax.

```yok
# create a temporary file from the given string and use the file descriptor for greps `stdin`
grep("test, stdin<="testing\ntesting\n1 2 3")
```

`stdout` and `stderr` can also be set for either commands or functions.

```yok
# silence all output from `cat` using the special `/dev/null` file descriptor
# also remap stderr to stdout, notice stdout here is a keyword, not a string
cat(:my_file.txt, stdout=:/dev/null, stderr=stdout)
```

using the `=>` syntax a file can be appended to, rather than overwritten.

```yok
cat(:my_file.txt, stdout=>:my_log.txt)
```

### Pipelines

**Yо̄k** supports classic **sh** pipelines.
The language treats commands and functions the same, meaning they can be used interchangeably in the pipeline.
In order to use a function in a pipeline it must use the `read` keyword to get input from `stdin` and `yield` a value.
Functions which do not read from `stdin` and `yield` a value will cause a compile time error if they are used in a pipeline.

```yok
use {
    cat
    grep
}

fn say_hello(greet) {
    let name = ""
    while read(name) {
        yield "{greet} {name}"
    }
}

let lex_greeting = cat(:names.txt) | say_hello("xin chao") | grep(:lex)
```

### Error Handling

**sh** relies on `error codes` and the special `$?` variable for handling errors.
**Yо̄k** cleans up the syntax around using these tools for error handling.
You can use the `catch` syntax to explicitly handle any non-zero error codes.
This is not required but can be useful to provide better error messages to your user, exit your script gracefully, and perform any necessary cleanup when your code fails.

```yok
let result = curl("localhost:8000/") catch(e) {
    print("failed to curl localhost, error_code:{e}")
    do_cleanup()
    result = :none
}
```

The `or` keyword can be used quickly set a default value when something fails.

```yok
let result = curl("localhost:8000/") or "request failed!"
```

Functions can also return error codes by returning a second value.
This value must be a string literal for a value between :1 and :255

```yok
fn div(a, b) {
    if b == 0 {
        # return the error code :1 here
        return :0, :1
    }
    # no error code is specified so the error code is set to :0
    return a / b
}
```

Error code returns can even be used from the top level of a script to exit with an error code

```yok
let password = ""
# read the password in from the user
read(password)

if password != "password" {
    return "invalid password", :1
}
```

### Yо̄k Builtins

**Yо̄k** comes with several useful builtins

* `print` is used to write output to the terminal.
    It writes to `stderr` rather than `stdout`.
    This is because `stdout` is used when piping commands together.
    `stdout` is also used to return values from functions (remember functions are essentially small shell commands).
    All of this makes writing user facing messages to `stderr` a much better default choice.
* `read` can be used to read strings from `stdin`.
    This is especially useful in pipelines and can be used in conjunction with the `while` and `yield` keywords to great effect.
* `len` can be used to get the length of a string in bytes.
* `replace` and `replace_all` can be used to replace substrings in a larger string.

### Inline Sh

In the case that direct used of `sh` script is required, it can be accessed using an `sh` block.
This code will not be validated by the **Yо̄k** compiler and breaks all guarantees that the **Yо̄k** language makes.
Use this feature with caution.

```yok
let greeting = "Hello"

sh {
    echo $GREETING
}
```

### Testing

Testing support is built directly into **Yо̄k**.
You can define a test anywhere in a **Yо̄k** script to test functionality.

```yok
fn div(a, b) {
    if b == 0 {
        return :0, :1
    }
    return a / b
}

test "div works as expected" {
    let got = div(:10, :2)
    assert got == :5, "div returned {got}, but wanted 5"

    let got_err = :false
    div(:10, :0) catch(e) {
        got_err = :true
    }
    assert got_err == :true, "wanted err but did not get one"
}
```

If you want to test your entire script, rather than a simple function, you can do so by calling `self()`.
This will execute the script, replacing all command and function with those defined in the test environment.

```yok
ls("-l") | wc("-l")

test "test full script" {
    # this function overwrites the `ls` command so it can be mocked
    fn ls() {
        return """total 0
file 1
file 2
file 3
"""
    }

    # 'got' here is populated with the contents of stdout after running the given script
    let got = self()
    assert got == :4, "the script returned {got}, but wanted :4"
}
```

running these tests is as simple as running `yok test [your yok file]`

### Macros

**sh** is a simple language and **Yо̄k** was designed to reflect this simplicity.
In order to help support this simplicity, **Yо̄k** includes a macro system.
Macros are implemented using the `mx`, `quote`, `unquote` and `body` keywords.

```yok
# unless is the opposite of 'if' and runs code only if the function check value is not true 
mx unless(check) {
    if unquote{check} {} else {
        # 'body' is a macro keyword representing the body passed to the macro in
        # in curly braces
        unquote{body}
    }
}

# when calling unless here`a == :true` is passed as `check` and everything 
# between the {} is passed as `body`
unless(a == :true) {
    print("a is false")
}

# the macro call expands to the following 
# if (a == :true) {} else {
#     print("a is false")
# }
```

### Data Structures

Data structures, like arrays and dicts, are not supported natively in **Yо̄k** (yet :D).
Instead use `jq` and JSON strings to represent these data structures.
