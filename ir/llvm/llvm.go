package llvm

import (
	"fmt"

	"github.com/EclesioMeloJunior/lotus/parser"
	"tinygo.org/x/go-llvm"
)

// IRGenerator represents an LLVM IR generator.
type IRGenerator struct {
	Module  llvm.Module
	builder llvm.Builder
	context llvm.Context

	globals map[string]llvm.Value
	locals  map[string]map[string]llvm.Value
}

// NewIRGenerator creates a new instance of IRGenerator.
func NewIRGenerator() *IRGenerator {
	context := llvm.NewContext()
	module := context.NewModule("main")
	builder := context.NewBuilder()
	return &IRGenerator{
		Module:  module,
		builder: builder,
		context: context,
		globals: make(map[string]llvm.Value),
		locals:  make(map[string]map[string]llvm.Value),
	}
}

// GenerateIR generates LLVM IR from the given AST.
func (gen *IRGenerator) GenerateIR(program *parser.Program) {
	for _, stmt := range program.Statements {
		switch stmt := stmt.(type) {
		case *parser.VarStatement:
			gen.generateVarStatement(stmt, "")
		case *parser.FnStatement:
			gen.generateFnStatement(stmt)
		}
	}
}

// generateVarStatement generates LLVM IR for a variable declaration.
func (gen *IRGenerator) generateVarStatement(stmt *parser.VarStatement, fnName string) {
	alloca := gen.builder.CreateAlloca(gen.context.Int32Type(), stmt.Name)
	varValue := gen.generateExpression(stmt.Value, fnName)

	gen.builder.CreateStore(varValue, alloca)

	if fnName != "" {
		gen.locals[fnName][stmt.Name] = alloca
	} else {
		gen.globals[stmt.Name] = alloca
	}
}

// generateFnStatement generates LLVM IR for a function declaration.
func (gen *IRGenerator) generateFnStatement(stmt *parser.FnStatement) {
	fnType := llvm.FunctionType(gen.context.Int32Type(), nil, false)
	fn := llvm.AddFunction(gen.Module, stmt.Name, fnType)
	fn.SetFunctionCallConv(llvm.CCallConv)
	entry := llvm.AddBasicBlock(fn, "entry")
	gen.builder.SetInsertPointAtEnd(entry)

	gen.locals = make(map[string]map[string]llvm.Value)
	gen.locals[stmt.Name] = make(map[string]llvm.Value)

	for _, bodyStmt := range stmt.Body {
		switch bodyStmt := bodyStmt.(type) {
		case *parser.VarStatement:
			gen.generateVarStatement(bodyStmt, stmt.Name)
		case *parser.ReturnStatement:
			gen.generateReturnStatement(bodyStmt, stmt.Name)
		}
	}
}

// generateReturnStatement generates LLVM IR for a return statement.
func (gen *IRGenerator) generateReturnStatement(stmt *parser.ReturnStatement, fnName string) {
	returnValue := gen.generateExpression(stmt.Value, fnName)
	gen.builder.CreateRet(returnValue)
}

// generateExpression generates LLVM IR for an expression.
func (gen *IRGenerator) generateExpression(expr parser.Expression, fnName string) llvm.Value {
	switch expr := expr.(type) {
	case *parser.IntegerLiteral:
		return llvm.ConstInt(gen.context.Int32Type(), uint64(expr.Value), false)
	case *parser.FloatLiteral:
		return llvm.ConstFloat(gen.context.FloatType(), expr.Value)
	case *parser.Identifier:
		return gen.builder.CreateLoad(gen.context.Int32Type(), gen.locals[fnName][expr.Value], expr.Value)
	case *parser.InfixExpression:
		left := gen.generateExpression(expr.Left, fnName)
		right := gen.generateExpression(expr.Right, fnName)

		switch expr.Operator {
		case "+":
			return gen.builder.CreateAdd(left, right, "addtmp")
		case "-":
			return gen.builder.CreateSub(left, right, "subtmp")
		case "*":
			return gen.builder.CreateMul(left, right, "multmp")
		case "/":
			return gen.builder.CreateSDiv(left, right, "divtmp")
		default:
			panic(fmt.Sprintf("unknown operator: %s", expr.Operator))
		}
	default:
		panic(fmt.Sprintf("unknown expression type: %T", expr))
	}
}

// Dump prints the generated LLVM IR.
func (gen *IRGenerator) Dump() {
	gen.Module.Dump()
}
