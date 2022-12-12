// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package funcorder

import (
	"os"
	"sort"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/analysis"
)

const (
	Name = "funcorder"
)

var (
	Analyzer = &analysis.Analyzer{
		Name: Name,
		Doc:  "check declaration order and count of types, constants, variables and functions",
		Run:  run,
	}
)

type (
	recvFunc struct {
		Index    int
		FuncName string
	}
)

func run(pass *analysis.Pass) (interface{}, error) {
	// a decorated ast package "dst" is used to avoid free floating comment issue
	// see https://github.com/golang/go/issues/20744
	dec := decorator.NewDecorator(pass.Fset) // holds mapping between ast and dst

	for _, f := range pass.Files {
		dstF, err := dec.DecorateFile(f) // transform ast file to dst file
		if err != nil {
			panic(err)
		}

		// build recv <> []func mapping
		rFuncs := make(map[string][]recvFunc)

		for i, decl := range dstF.Decls {
			switch funcDecl := decl.(type) {
			case *dst.FuncDecl:
				recv := getRecvStructName(funcDecl)
				if recv == "" || isMock(recv) {
					continue
				}
				rFuncs[recv] = append(rFuncs[recv], recvFunc{
					Index:    i,
					FuncName: funcDecl.Name.Name,
				})
			}
		}

		// rearrange dst ordering for each recv
		for _, funcList := range rFuncs {
			sort.Slice(funcList, func(i, j int) bool {
				return strings.Compare(funcList[i].FuncName, funcList[j].FuncName) < 0
			})
			selectedIndex := make(map[int]struct{})
			for i := range funcList {
				selectedIndex[funcList[i].Index] = struct{}{}
			}

			cursor := 0
			var newDecls []dst.Decl
			for i := 0; i < len(dstF.Decls); i++ {
				index := i
				if _, ok := selectedIndex[i]; ok {
					index = funcList[cursor].Index
					cursor++
				}
				newDecls = append(newDecls, dstF.Decls[index])
			}
			dstF.Decls = newDecls
		}

		// save cleaned file
		fName := pass.Fset.Position(f.Pos()).Filename
		writer, err := os.OpenFile(fName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}

		err = decorator.Fprint(writer, dstF)
		if err != nil {
			panic(err)
		}
	}

	return nil, nil
}

func getRecvStructName(funcDecl *dst.FuncDecl) string {
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) == 1 && funcDecl.Name.IsExported() {
		if expr, ok := funcDecl.Recv.List[0].Type.(*dst.StarExpr); ok {
			if ident, ok := expr.X.(*dst.Ident); ok {
				return ident.Name
			}
		}
	}

	return ""
}

func isMock(s string) bool {
	return len(s) > 4 && s[0:4] == "Mock"
}
