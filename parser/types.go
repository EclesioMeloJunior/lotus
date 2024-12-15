package parser

import (
	"errors"
	"fmt"
	"strings"
)

// Types defines behaviors for a certain instance/raw value.
// e.g Int32 is type that defines arithmetic behaviors for a raw value.
var ErrWrongTypeAssigment = errors.New("wrong type assignment")

type Type int

const (
	Void Type = iota
	Int32
	String
	Float32
)

func (t *Type) Verify(st Expression) error {
	switch *t {
	case Int32:
		return verifyInt32(st)
	case String:
		return verifyString(st)
	case Void:
		return nil
	default:
		return ErrWrongTypeAssigment
	}
}

func verifyInt32(st Expression) error {
	switch inner := st.(type) {
	case *Identifier:
		if inner.Type == Int32 {
			return nil
		}
	case *IntegerLiteral:
		return nil
	case *InfixExpression:
		if err := verifyInt32(inner.Left); err != nil {
			return err
		}

		if !strings.ContainsAny(inner.Operator, "+-*/") {
			return fmt.Errorf("int32 allowed infix operators: + - * /")
		}

		if err := verifyInt32(inner.Right); err != nil {
			return err
		}

		return nil
	case *PrefixExpression:
		if inner.Operator == "~" || inner.Operator == "++" || inner.Operator == "--" {
			return fmt.Errorf("int32 allowed prefix operators: ~ ++ --")
		}

		if err := verifyInt32(inner.Right); err != nil {
			return err
		}

		return nil
	}

	return ErrWrongTypeAssigment
}

func verifyString(st Expression) error {
	switch inner := st.(type) {
	case *Identifier:
		if inner.Type == String {
			return nil
		}
	case *StringLiteral:
		return nil
	case *InfixExpression:
		if err := verifyString(inner.Left); err != nil {
			return err
		}

		if inner.Operator != "+" {
			return fmt.Errorf("string allowed infix operators: + (concatenation)")
		}

		if err := verifyString(inner.Right); err != nil {
			return err
		}

		return nil
	}

	return ErrWrongTypeAssigment
}
