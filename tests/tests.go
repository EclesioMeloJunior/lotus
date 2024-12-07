package tests

import (
	"testing"

	"github.com/EclesioMeloJunior/lotus/internal/source"
	"github.com/EclesioMeloJunior/lotus/ir/llvm"
	"github.com/stretchr/testify/require"
	gollvm "tinygo.org/x/go-llvm"
)

func readInput(t *testing.T, input string) *source.SourceFile {
	t.Helper()

	src, err := source.FromFile(input)
	require.NoError(t, err)
	return src
}

func runMainFn(t *testing.T, irGen *llvm.IRGenerator,
	execResult func(gollvm.ExecutionEngine, gollvm.GenericValue)) {
	t.Helper()

	gollvm.LinkInMCJIT()
	gollvm.InitializeNativeTarget()
	gollvm.InitializeNativeAsmPrinter()

	options := gollvm.NewMCJITCompilerOptions()
	options.SetMCJITOptimizationLevel(2)
	options.SetMCJITEnableFastISel(true)
	options.SetMCJITNoFramePointerElim(true)
	options.SetMCJITCodeModel(gollvm.CodeModelJITDefault)

	engine, err := gollvm.NewMCJITCompiler(irGen.Module, options)
	require.NoError(t, err)
	defer engine.Dispose()

	output := engine.RunFunction(irGen.Module.NamedFunction("main"), nil)
	execResult(engine, output)
}
