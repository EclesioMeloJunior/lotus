#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Print commands and their arguments as they are executed
set -x

CPPFLAGS=$(`which llvm-config` --cppflags)
LDFLAGS=$(`which llvm-config` --ldflags --libs --system-libs all | tr '\n' ',')
go env -w CGO_CPPFLAGS="$CPPFLAGS"
go env -w CGO_LDFLAGS="$LDFLAGS"
