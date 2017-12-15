package goja

import (
	"context"
	"fmt"

	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
)

// InlineRequires attempts to generate new source code that replaces
// top-level require() calls with code that those calls reference.
//
// Originally, this code processed and emitted Goja abstract syntax
// trees (ASTs) (see github.com/dop251/goja/ast).  An emitted tree was
// a modified version of the input tree.  Each library was either
// pre-parsed, pre-compiled, or plain source. But Goja can't currently
// support (easily) modification of ASTs or Programs; therefore, the
// current implementation rewrites the given source based on
// processing of the given source's AST.
//
// An alternative approach is simply to define a require() function
// and put that function in the runtime's environment. However, that
// approach would require the use of eval at runtime, which would
// prevent precompilation. With the current implementation, actions
// and guards can be precompiled when a machine specification is
// compiled. That precompilation can given substantial runtime
// performance benefits when libraries are large.
func InlineRequires(ctx context.Context, src string, provider func(context.Context, string) (string, error)) (string, error) {

	p, err := parser.ParseFile(nil, "", src, 0)
	if err != nil {
		return "", err
	}

	type Required struct {
		Idx0 int
		Idx1 int
		Name string
	}

	requires := make([]Required, 0, 8)

	for _, s := range p.Body {
		exps, is := s.(*ast.ExpressionStatement)
		if !is {
			continue
		}

		call, is := exps.Expression.(*ast.CallExpression)
		if !is {
			continue
		}

		id, is := call.Callee.(*ast.Identifier)
		if !is {
			continue
		}
		if id.Name != "require" {
			continue
		}
		if len(call.ArgumentList) != 1 {
			return "", fmt.Errorf("bad require args: %#v", call.ArgumentList)
		}

		arg := call.ArgumentList[0]
		lit, is := arg.(*ast.StringLiteral)
		if !is {
			return "", fmt.Errorf("bad require arg: %#v", arg)
		}

		s := lit.Value
		requires = append(requires, Required{
			Idx0: int(exps.Idx0()),
			Idx1: int(exps.Idx1()),
			Name: s,
		})
	}

	var inlined string
	switch len(requires) {
	case 0:
		inlined = src
	default:
		inlined = src[0 : requires[0].Idx0-1]
		for i, r := range requires {
			lib, err := provider(ctx, r.Name)
			if err != nil {
				return "", err
			}

			inlined += lib

			to := len(src)
			if i < len(requires)-1 {
				to = requires[len(requires)-1].Idx0 - 1
			}
			inlined += src[requires[i].Idx1:to]
		}
	}

	return inlined, nil
}
