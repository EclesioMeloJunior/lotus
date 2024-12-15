#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Print commands and their arguments as they are executed
set -x

llc -filetype=obj output.ll -o output.o
clang output.o -o output
