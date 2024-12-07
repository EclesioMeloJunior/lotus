# Lotus

This is a language that uses LLVM to generate the intermediary representation before compiling to machine code


### Getting started

- First you need to install LLVM 16+
- Run the `./setup.sh` script to changes Golang compilation flags to use LLVM
- You can execute a example lotus code: `go run ./... tests/simple_main.lt`
- A file `output.ll` you be generated, then you can use:

```sh
llc -filetype=obj output.ll -o output.o
clang output.o -o output
./output
echo $?
```
