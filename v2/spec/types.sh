#!/bin/sh

_true=1
_false=0

a=10      # type is int
b="str"   # type is string
c=$_false # type is bool

d=20     # type is still int
e="str"  # type is still string
f=$_true # type is still bool

g=0       # type is int and default vaue is 0
h=""      # type is string and default value is ""
i=$_false # type is bool and default vaue is false

# constants are replaced with values by the complier