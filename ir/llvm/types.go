package llvm

import (
	"tinygo.org/x/go-llvm"
)

var stringType llvm.Type

func init() {
	stringType = llvm.GlobalContext().StructCreateNamed("String")
	stringType.StructSetBody([]llvm.Type{
		llvm.GlobalContext().Int32Type(),
		llvm.PointerType(llvm.GlobalContext().Int8Type(), 0)},
		false)
}
