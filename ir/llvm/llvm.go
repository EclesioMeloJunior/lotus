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
	gen.generate(program.Statements, "")
}

func (gen *IRGenerator) generate(stmts []parser.Node, fnName string) {
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *parser.VarStatement:
			gen.generateVarStatement(stmt, fnName)
		case *parser.ReassignVarStatement:
			gen.generateVarStatement(stmt.ToVarAssignment(), fnName)
		case *parser.FnStatement:
			if fnName != "" {
				panic("functions cannot be inside other functions")
			}

			gen.generateFnStatement(stmt)
		case *parser.ReturnStatement:
			gen.generateReturnStatement(stmt, fnName)
		}
	}
}

// generateVarStatement generates LLVM IR for a variable declaration.
func (gen *IRGenerator) generateVarStatement(stmt *parser.VarStatement, fnName string) {
	var alloca llvm.Value
	if stmt.Type != parser.Void {
		alloca = gen.builder.CreateAlloca(gen.fromRawTypeToLLVMType(stmt.Type), stmt.Name)
	} else {
		alloca = gen.builder.CreateAlloca(gen.context.Int8Type(), stmt.Name)
	}

	if stmt.Value != nil {
		varValue := gen.generateExpression(stmt.Value, fnName)
		gen.builder.CreateStore(varValue, alloca)
	}

	if fnName != "" {
		gen.locals[fnName][stmt.Name] = alloca
	} else {
		gen.globals[stmt.Name] = alloca
	}
}

func (gen *IRGenerator) fromRawTypeToLLVMType(rawType parser.Type) llvm.Type {
	switch rawType {
	case parser.Int32:
		return gen.context.Int32Type()
	case parser.String:
		return stringType
	case parser.Void:
		return gen.context.VoidType()
	case parser.Float32:
		return gen.context.FloatType()
	default:
		panic(fmt.Sprintf("type %v not supported", rawType))
	}
}

func (gen *IRGenerator) getFnSignatureType(stmt *parser.FnStatement) llvm.Type {
	returnType := gen.fromRawTypeToLLVMType(stmt.ReturnType)

	var paramsTypes []llvm.Type
	for _, p := range stmt.Args {
		paramsTypes = append(paramsTypes, gen.fromRawTypeToLLVMType(p.Type))
	}

	return llvm.FunctionType(returnType, paramsTypes, false)
}

// generateFnStatement generates LLVM IR for a function declaration.
func (gen *IRGenerator) generateFnStatement(stmt *parser.FnStatement) {
	fnType := gen.getFnSignatureType(stmt)
	fn := llvm.AddFunction(gen.Module, stmt.Name, fnType)

	fn.SetFunctionCallConv(llvm.CCallConv)
	entry := llvm.AddBasicBlock(fn, "entry")
	gen.builder.SetInsertPointAtEnd(entry)

	gen.locals = make(map[string]map[string]llvm.Value)
	gen.locals[stmt.Name] = make(map[string]llvm.Value)

	if len(stmt.Body) == 0 {
		gen.builder.CreateRetVoid()
		return
	}

	gen.generate(stmt.Body, stmt.Name)
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
	case *parser.StringLiteral:
		return llvm.ConstString(expr.Value, true)
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
