// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
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

package main

import (
	"fmt"
	"go/importer"
	"go/types"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const licence = `// Copyright (c) 2017-2020 Uber Technologies Inc.
// 
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

`

var typesHeader = template.Must(template.New("struct type").Parse(licence + `package types

`))

var mapperHeader = template.Must(template.New("struct type").Parse(licence + `package thrift

import (
	"github.com/uber/cadence/common/types"

{{range .}}	"{{.ThriftPackage}}"
{{end}}
)
`))

var structTemplate = template.Must(template.New("struct type").Parse(`
type {{.Type.Name}} struct {
{{range .Fields}}	{{.Name}} {{if .Type.IsMap}}map[string]{{end}}{{if .Type.IsArray}}[]{{end}}{{if .Type.IsPointer}}*{{end}}{{.Type.Name}}
{{end}}}
`))

var enumTemplate = template.Must(template.New("enum type").Parse(`
type {{.Type.Name}} int32

const ({{range $i, $v := .Values}}
	{{.}}{{if eq $i 0}} {{$.Type.Name}} = iota{{end}}{{end}}
)
`))

var structMapperTemplate = template.Must(template.New("struct mapper").Parse(`
// From{{.Type.Name}} converts internal {{.Type.Name}} type to thrift
func From{{.Type.Name}}(t *types.{{.Type.Name}}) *{{.Type.ThriftPackage}}.{{.Type.Name}} {
	if t == nil {
		return nil
	}
	return &{{.Type.ThriftPackage}}.{{.Type.Name}}{
{{range .Fields}}		{{.Name}}: {{if .Type.IsPrimitive}}t.{{.Name}}{{else}}From{{.Type.Name}}{{if .Type.IsArray}}Array{{end}}{{if .Type.IsMap}}Map{{end}}(t.{{.Name}}){{end}},
{{end}}	}
}

// To{{.Type.Name}} converts thrift {{.Type.Name}} type to internal
func To{{.Type.Name}}(t *{{.Type.ThriftPackage}}.{{.Type.Name}}) *types.{{.Type.Name}} {
	if t == nil {
		return nil
	}
	return &types.{{.Type.Name}}{
{{range .Fields}}		{{.Name}}: {{if .Type.IsPrimitive}}t.{{.Name}}{{else}}To{{.Type.Name}}{{if .Type.IsArray}}Array{{end}}{{if .Type.IsMap}}Map{{end}}(t.{{.Name}}){{end}},
{{end}}	}
}
`))

var arrayMapperTemplate = template.Must(template.New("array mapper").Parse(`
// From{{.Type.Name}}Array converts internal {{.Type.Name}} type array to thrift
func From{{.Type.Name}}Array(t []*types.{{.Type.Name}}) []*{{.Type.ThriftPackage}}.{{.Type.Name}} {
	if t == nil {
		return nil
	}
	v := make([]*{{.Type.ThriftPackage}}.{{.Type.Name}}, len(t))
	for i := range t {
		v[i] = {{if .Type.IsPrimitive}}t[i]{{.Name}}{{else}}From{{.Type.Name}}(t[i]){{end}}
	}
	return v
}

// To{{.Type.Name}}Array converts thrift {{.Type.Name}} type array to internal
func To{{.Type.Name}}Array(t []*{{.Type.ThriftPackage}}.{{.Type.Name}}) []*types.{{.Type.Name}} {
	if t == nil {
		return nil
	}
	v := make([]*types.{{.Type.Name}}, len(t))
	for i := range t {
		v[i] = {{if .Type.IsPrimitive}}t[i]{{.Name}}{{else}}To{{.Type.Name}}(t[i]){{end}}
	}
	return v
}
`))

var mapMapperTemplate = template.Must(template.New("map mapper").Parse(`
// From{{.Type.Name}}Map converts internal {{.Type.Name}} type map to thrift
func From{{.Type.Name}}Map(t map[string]{{if .Type.IsPointer}}*{{end}}types.{{.Type.Name}}) map[string]{{if .Type.IsPointer}}*{{end}}{{.Type.ThriftPackage}}.{{.Type.Name}} {
	if t == nil {
		return nil
	}
	v := make(map[string]{{if .Type.IsPointer}}*{{end}}{{.Type.ThriftPackage}}.{{.Type.Name}}, len(t))
	for key := range t {
		v[key] = {{if .Type.IsPrimitive}}t[key]{{else}}From{{.Type.Name}}(t[key]){{end}}
	}
	return v
}

// To{{.Type.Name}}Map converts thrift {{.Type.Name}} type map to internal
func To{{.Type.Name}}Map(t map[string]{{if .Type.IsPointer}}*{{end}}{{.Type.ThriftPackage}}.{{.Type.Name}}) map[string]{{if .Type.IsPointer}}*{{end}}types.{{.Type.Name}} {
	if t == nil {
		return nil
	}
	v := make(map[string]{{if .Type.IsPointer}}*{{end}}types.{{.Type.Name}}, len(t))
	for key := range t {
		v[key] = {{if .Type.IsPrimitive}}t[key]{{else}}To{{.Type.Name}}(t[key]){{end}}
	}
	return v
}
`))

var enumMapperTemplate = template.Must(template.New("enum mapper").Parse(`
// From{{.Type.Name}} converts internal {{.Type.Name}} type to thrift
func From{{.Type.Name}}(t {{if .Type.IsPointer}}*{{end}}types.{{.Type.Name}}) {{if .Type.IsPointer}}*{{end}}{{.Type.ThriftPackage}}.{{.Type.Name}} {
	{{if .Type.IsPointer}}if t == nil {
		return nil
	}{{end}}
	switch {{if .Type.IsPointer}}*{{end}}t { {{range .Values}}
		case types.{{.}}: {{if $.Type.IsPointer}}v := {{$.Type.ThriftPackage}}.{{.}}; return &v{{else}}return {{$.Type.ThriftPackage}}.{{.}}{{end}}{{end}}
	}
	panic("unexpected enum value")
}

// To{{.Type.Name}} converts thrift {{.Type.Name}} type to internal
func To{{.Type.Name}}(t {{if .Type.IsPointer}}*{{end}}{{.Type.ThriftPackage}}.{{.Type.Name}}) {{if .Type.IsPointer}}*{{end}}types.{{.Type.Name}} {
	{{if .Type.IsPointer}}if t == nil {
		return nil
	}{{end}}
	switch {{if .Type.IsPointer}}*{{end}}t { {{range .Values}}
		case {{$.Type.ThriftPackage}}.{{.}}: {{if $.Type.IsPointer}}v := types.{{.}}; return &v{{else}}return types.{{.}}{{end}}{{end}}
	}
	panic("unexpected enum value")
}
`))

var requiredArrayMappers = map[string]struct{}{}
var requiredMapMappers = map[string]struct{}{}

type (
	// Renderer can render internal type and its mappings
	Renderer interface {
		RenderType(w io.Writer)
		RenderMapper(w io.Writer)
	}

	// Type describes a type
	Type struct {
		FullThriftPackage string
		ThriftPackage     string
		Name              string
		IsPrimitive       bool
		IsArray           bool
		IsMap             bool
		IsPointer         bool
	}

	// Field describe a field within a struct
	Field struct {
		Name string
		Type Type
	}
	// Struct describe struct type
	Struct struct {
		Type   Type
		Fields []Field
	}
	// Array describes Array type
	Array struct {
		Type
	}
	// Map describes Map type
	Map struct {
		Type
	}
	// Enum describes Enum type
	Enum struct {
		Type   Type
		Values []string
	}
)

func (s *Struct) RenderType(w io.Writer) {
	err := structTemplate.Execute(w, s)
	if err != nil {
		panic(err)
	}
}
func (s *Struct) RenderMapper(w io.Writer) {
	err := structMapperTemplate.Execute(w, s)
	if err != nil {
		panic(err)
	}
}
func (a *Array) RenderMapper(w io.Writer) {
	err := arrayMapperTemplate.Execute(w, a)
	if err != nil {
		panic(err)
	}
}
func (m *Map) RenderMapper(w io.Writer) {
	err := mapMapperTemplate.Execute(w, m)
	if err != nil {
		panic(err)
	}
}
func (e *Enum) RenderType(w io.Writer) {
	err := enumTemplate.Execute(w, e)
	if err != nil {
		panic(err)
	}
}
func (e *Enum) RenderMapper(w io.Writer) {
	err := enumMapperTemplate.Execute(w, e)
	if err != nil {
		panic(err)
	}
}

// fullName example: []*github.com/uber/cadence/.gen/go/shared.VersionHistoryItem
func newType(fullName string) Type {
	isMap := false
	if strings.HasPrefix(fullName, "map[string]") {
		fullName = strings.TrimPrefix(fullName, "map[string]")
		isMap = true
	}
	isArray := false
	if strings.HasPrefix(fullName, "[]") {
		fullName = strings.TrimPrefix(fullName, "[]")
		isArray = true
	}
	isPointer := false
	if strings.HasPrefix(fullName, "*") {
		fullName = strings.TrimPrefix(fullName, "*")
		isPointer = true
	}
	pos := strings.LastIndexByte(fullName, '.')
	if pos > 0 {
		pkg := fullName[:pos]
		name := fullName[pos+1:]
		pos = strings.LastIndexByte(pkg, '/')
		short := pkg[pos+1:]
		return Type{pkg, short, name, false, isArray, isMap, isPointer}
	}
	return Type{"", "", fullName, true, isArray, isMap, isPointer}
}

func newStruct(obj types.Object) *Struct {
	u := obj.Type().Underlying().(*types.Struct)
	fields := make([]Field, u.NumFields())
	for i := 0; i < u.NumFields(); i++ {
		f := u.Field(i)
		field := Field{
			Name: f.Name(),
			Type: newType(f.Type().String()),
		}
		if field.Type.IsArray && !field.Type.IsPrimitive {
			requiredArrayMappers[f.Type().String()] = struct{}{}
		}
		if field.Type.IsMap && !field.Type.IsPrimitive {
			requiredMapMappers[f.Type().String()] = struct{}{}
		}
		fields[i] = field
	}
	return &Struct{
		Type:   newType(obj.Type().String()),
		Fields: fields,
	}
}

func newArray(fullType string) *Array {
	return &Array{
		Type: newType(fullType),
	}
}

func newMap(fullType string) *Map {
	return &Map{
		Type: newType(fullType),
	}
}

var enumPointerExceptions = map[string]struct{}{
	"IndexedValueType": {},
}

func newEnum(obj types.Object) *Enum {
	enumType := newType(obj.Type().String())
	if _, ok := enumPointerExceptions[enumType.Name]; !ok {
		enumType.IsPointer = true
	}
	enumValues := make([]string, 0, 128)
	pkg := obj.Pkg().Scope()
	for _, name := range pkg.Names() {
		eValue := pkg.Lookup(name)
		if isEnumValue(eValue, obj.Type()) {
			enumValues = append(enumValues, eValue.Name())
		}
	}
	return &Enum{
		Type:   enumType,
		Values: enumValues,
	}
}

func isEnumValue(v types.Object, t types.Type) bool {
	c, isConst := v.(*types.Const)
	return isConst && c.Type() == t
}

func rewriteFile(path string) *os.File {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	return f
}

func createRenderer(obj types.Object) Renderer {
	if _, ok := obj.(*types.TypeName); ok {
		switch obj.Type().Underlying().(type) {
		case *types.Struct:
			return newStruct(obj)
		case *types.Basic:
			return newEnum(obj)
		default:
			fmt.Printf("encountered unexpected type: %v\n", obj)
		}
	}
	return nil
}

type Package struct {
	ThriftPackage string
	TypesFile     string
	MapperFile    string
}

func main() {
	packages := []Package{
		{
			ThriftPackage: "github.com/uber/cadence/.gen/go/shared",
			TypesFile:     "common/types/shared.go",
			MapperFile:    "common/types/mapper/thrift/shared.go",
		},
	}

	for _, p := range packages {
		typesFile := rewriteFile(p.TypesFile)
		mapperFile := rewriteFile(p.MapperFile)

		typesHeader.Execute(typesFile, packages)
		mapperHeader.Execute(mapperFile, packages)

		pkg, err := importer.Default().Import(p.ThriftPackage)
		if err != nil {
			panic(err)
		}

		for _, name := range pkg.Scope().Names() {
			obj := pkg.Scope().Lookup(name)
			if !obj.Exported() {
				continue
			}
			if r := createRenderer(obj); r != nil {
				r.RenderType(typesFile)
				r.RenderMapper(mapperFile)
			}
		}
		for m := range requiredArrayMappers {
			newArray(m).RenderMapper(mapperFile)
		}
		for m := range requiredMapMappers {
			newMap(m).RenderMapper(mapperFile)
		}

		typesFile.Close()
		mapperFile.Close()
	}
}
