package llvm

import "tinygo.org/x/go-llvm"

func stringType() llvm.Type {
	stringType := llvm.GlobalContext().StructCreateNamed("String")
	stringType.StructSetBody([]llvm.Type{
		llvm.GlobalContext().Int32Type(),
		llvm.PointerType(llvm.GlobalContext().Int8Type(), 0)},
		false)

	return stringType
}
