package gox

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/goplus/gox/internal"
)

// ----------------------------------------------------------------------------

// A Variable represents a function parameters and results.
type Param = types.Var

// NewParam returns a new variable representing a function parameter.
func (p *Package) NewParam(name string, typ types.Type) *Param {
	return types.NewParam(token.NoPos, p.Types, name, typ)
}

// ----------------------------------------------------------------------------

// A Tuple represents an ordered list of variables; a nil *Tuple is a valid (empty) tuple.
// Tuples are used as components of signatures and to represent the type of multiple
// assignments; they are not first class types of Go.
type Tuple = types.Tuple

// NewTuple returns a new tuple for the given parameters.
func NewTuple(x ...*Param) *Tuple {
	return types.NewTuple(x...)
}

// ----------------------------------------------------------------------------

// Func type
type Func struct {
	*types.Func
	decl *ast.FuncDecl
	old  codeBlockCtx
}

// BodyStart func
func (p *Func) BodyStart(pkg *Package) *CodeBuilder {
	return pkg.cb.startCodeBlock(p, &p.old)
}

// End is for internal use.
func (p *Func) End(cb *CodeBuilder) {
	pkg := cb.pkg
	stmts := cb.endCodeBlock(p.old)
	body := &ast.BlockStmt{List: stmts}
	t := p.Type().(*types.Signature)
	if fn := p.decl; fn == nil { // is closure
		expr := &ast.FuncLit{Type: toFuncType(pkg, t), Body: body}
		cb.stk.Push(internal.Elem{Val: expr, Type: t})
	} else {
		fn.Name, fn.Type, fn.Body = ident(p.Name()), toFuncType(pkg, t), body
		if recv := t.Recv(); recv != nil {
			fn.Recv = toRecv(pkg, recv)
		}
	}
}

// NewFunc func
func (p *Package) NewFunc(recv *Param, name string, params, results *Tuple, variadic bool) *Func {
	sig := types.NewSignature(recv, params, results, variadic)
	return p.NewFuncWith(name, sig)
}

// NewFuncWith func
func (p *Package) NewFuncWith(name string, sig *types.Signature) *Func {
	if name == "" {
		panic("TODO: no func name")
	}
	p.endImport()

	fn := types.NewFunc(token.NoPos, p.Types, name, sig)
	p.Types.Scope().Insert(fn)

	decl := &ast.FuncDecl{}
	p.decls = append(p.decls, decl)
	return &Func{Func: fn, decl: decl}
}

// ----------------------------------------------------------------------------