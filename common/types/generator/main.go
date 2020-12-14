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
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"
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

func internalName(name string) string {
	return noUnderscores(capitalizeID(name))
}

func noUnderscores(name string) string {
	return strings.ReplaceAll(name, "_", "")
}

func capitalizeID(name string) string {
	if index := strings.Index(name, "Id"); index > 0 {
		nextWordIndex := index + len("Id")
		if nextWordIndex >= len(name) || unicode.IsUpper([]rune(name)[nextWordIndex]) {
			return strings.Replace(name, "Id", "ID", 1)
		}
	}
	return name
}

var capitalizedEnums = map[string]struct{}{
	"DomainStatus":                               {},
	"TimeoutType":                                {},
	"ParentClosePolicy":                          {},
	"DecisionTaskFailedCause":                    {},
	"DecisionTaskTimedOutCause":                  {},
	"CancelExternalWorkflowExecutionFailedCause": {},
	"SignalExternalWorkflowExecutionFailedCause": {},
	"ChildWorkflowExecutionFailedCause":          {},
	"WorkflowExecutionCloseStatus":               {},
	"QueryTaskCompletedType":                     {},
	"QueryResultType":                            {},
	"PendingActivityState":                       {},
	"PendingDecisionState":                       {},
	"HistoryEventFilterType":                     {},
	"TaskListKind":                               {},
	"ArchivalStatus":                             {},
	"IndexedValueType":                           {},
	"QueryRejectCondition":                       {},
	"QueryConsistencyLevel":                      {},
	"TaskSource":                                 {},
}

func enumString(value, enum string) string {
	trimmed := strings.TrimPrefix(value, enum)
	if _, ok := capitalizedEnums[enum]; ok {
		capitalized := ""
		for i, c := range trimmed {
			if i > 0 && unicode.IsUpper(c) {
				capitalized = capitalized + "_"
			}
			capitalized = capitalized + string(unicode.ToUpper(c))
		}
		return capitalized
	}
	return trimmed
}

var funcMap = template.FuncMap{
	"internal":   internalName,
	"enumString": enumString,
	"upper":      strings.ToUpper,
}

var typesHeader = template.Must(template.New("struct type").Funcs(funcMap).Parse(licence + `package types

`))

var mapperHeader = template.Must(template.New("struct type").Funcs(funcMap).Parse(licence + `package thrift

import (
	"github.com/uber/cadence/common/types"

{{range .}}	"{{.ThriftPackage}}"
{{end}}
)
`))

var structTemplate = template.Must(template.New("struct type").Funcs(funcMap).Parse(`
// {{.Prefix}}{{internal .Name}} is an internal type (TBD...)
type {{.Prefix}}{{internal .Name}} struct {
{{range .Fields}}	{{internal .Name}} {{if .Type.IsMap}}map[{{.Type.MapKeyType}}]{{end}}{{if .Type.IsArray}}[]{{end}}{{if .Type.IsPointer}}*{{end}}{{.Type.Prefix}}{{internal .Type.Name}} ` + "`{{.Tag}}`" + `
{{end}}}
{{range .Fields}}
// Get{{internal .Name}} is an internal getter (TBD...)
func (v *{{$.Prefix}}{{internal $.Name}}) Get{{internal .Name}}() (o {{if .Type.IsMap}}map[{{.Type.MapKeyType}}]{{end}}{{if .Type.IsArray}}[]{{end}}{{if .Type.IsPointer | and (not .Type.IsPrimitive) | and (not .Type.IsEnum)}}*{{end}}{{.Type.Prefix}}{{internal .Type.Name}}) {
	if v != nil{{if .Type.IsMap | or .Type.IsArray | or .Type.IsPointer}} && v.{{internal .Name}} != nil{{end}} {
		return {{if .Type.IsPointer | and (or .Type.IsPrimitive .Type.IsEnum)}}*{{end}}v.{{internal .Name}}
	}
	{{if and (eq $.Name "RegisterDomainRequest") (eq .Name "EmitMetric") }}o = true
{{end}}return
}
{{end}}
`))

var enumTemplate = template.Must(template.New("enum type").Funcs(funcMap).Parse(`
// {{internal .Name}} is an internal type (TBD...)
type {{internal .Name}} int32

// Ptr is a helper function for getting pointer value
func (e {{internal .Name}}) Ptr() *{{internal .Name}} {
	return &e
}

// String returns a readable string representation of {{internal .Name}}.
func (e {{internal .Name}}) String() string {
	w := int32(e)
	switch w { {{range $i, $v := .EnumValues}}
	case {{$i}}: return "{{enumString . $.Name}}"{{end}}
	}
	return fmt.Sprintf("{{internal .Name}}(%d)", w)
}

// UnmarshalText parses enum value from string representation
func (e *{{internal .Name}}) UnmarshalText(value []byte) error {
	switch s := strings.ToUpper(string(value)); s { {{range $i, $v := .EnumValues}}
	case "{{upper (enumString . $.Name)}}":
		*e = {{internal .}}
		return nil{{end}}
	default:
		val, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return fmt.Errorf("unknown enum value %q for %q: %v", s, "{{internal .Name}}", err)
		}
		*e = {{internal .Name}}(val)
		return nil
	}
}

// MarshalText encodes {{internal .Name}} to text.
func (e {{internal .Name}}) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

const ({{range $i, $v := .EnumValues}}
	// {{internal .}} is an option for {{internal $.Name}}
	{{internal .}}{{if eq $i 0}} {{internal $.Name}} = iota{{end}}{{end}}
)
`))

var structMapperTemplate = template.Must(template.New("struct mapper").Funcs(funcMap).Parse(`
// From{{.Prefix}}{{internal .Name}} converts internal {{.Name}} type to thrift
func From{{.Prefix}}{{internal .Name}}(t *types.{{.Prefix}}{{internal .Name}}) *{{.ThriftPackage}}.{{.Name}} {
	if t == nil {
		return nil
	}
	return &{{.ThriftPackage}}.{{.Name}}{
{{range .Fields}}		{{.Name}}: {{if .Type.IsPrimitive}}t.{{internal .Name}}{{else}}From{{.Type.Prefix}}{{internal .Type.Name}}{{if .Type.IsArray}}Array{{end}}{{if .Type.IsMap}}Map{{end}}(t.{{internal .Name}}){{end}},
{{end}}	}
}

// To{{.Prefix}}{{internal .Name}} converts thrift {{.Name}} type to internal
func To{{.Prefix}}{{internal .Name}}(t *{{.ThriftPackage}}.{{.Name}}) *types.{{.Prefix}}{{internal .Name}} {
	if t == nil {
		return nil
	}
	return &types.{{.Prefix}}{{internal .Name}}{
{{range .Fields}}		{{internal .Name}}: {{if .Type.IsPrimitive}}t.{{.Name}}{{else}}To{{.Type.Prefix}}{{internal .Type.Name}}{{if .Type.IsArray}}Array{{end}}{{if .Type.IsMap}}Map{{end}}(t.{{.Name}}){{end}},
{{end}}	}
}
`))

var arrayMapperTemplate = template.Must(template.New("array mapper").Funcs(funcMap).Parse(`
// From{{internal .Name}}Array converts internal {{internal .Name}} type array to thrift
func From{{internal .Name}}Array(t []*types.{{internal .Name}}) []*{{.ThriftPackage}}.{{.Name}} {
	if t == nil {
		return nil
	}
	v := make([]*{{.ThriftPackage}}.{{.Name}}, len(t))
	for i := range t {
		v[i] = {{if .IsPrimitive}}t[i]{{.Name}}{{else}}From{{internal .Name}}(t[i]){{end}}
	}
	return v
}

// To{{internal .Name}}Array converts thrift {{.Name}} type array to internal
func To{{internal .Name}}Array(t []*{{.ThriftPackage}}.{{.Name}}) []*types.{{internal .Name}} {
	if t == nil {
		return nil
	}
	v := make([]*types.{{internal .Name}}, len(t))
	for i := range t {
		v[i] = {{if .IsPrimitive}}t[i]{{.Name}}{{else}}To{{internal .Name}}(t[i]){{end}}
	}
	return v
}
`))

var mapMapperTemplate = template.Must(template.New("map mapper").Funcs(funcMap).Parse(`
// From{{internal .Name}}Map converts internal {{internal .Name}} type map to thrift
func From{{internal .Name}}Map(t map[{{.MapKeyType}}]{{if .IsPointer}}*{{end}}types.{{internal .Name}}) map[{{.MapKeyType}}]{{if .IsPointer}}*{{end}}{{.ThriftPackage}}.{{.Name}} {
	if t == nil {
		return nil
	}
	v := make(map[{{.MapKeyType}}]{{if .IsPointer}}*{{end}}{{.ThriftPackage}}.{{.Name}}, len(t))
	for key := range t {
		v[key] = {{if .IsPrimitive}}t[key]{{else}}From{{internal .Name}}(t[key]){{end}}
	}
	return v
}

// To{{internal .Name}}Map converts thrift {{.Name}} type map to internal
func To{{internal .Name}}Map(t map[{{.MapKeyType}}]{{if .IsPointer}}*{{end}}{{.ThriftPackage}}.{{.Name}}) map[{{.MapKeyType}}]{{if .IsPointer}}*{{end}}types.{{internal .Name}} {
	if t == nil {
		return nil
	}
	v := make(map[{{.MapKeyType}}]{{if .IsPointer}}*{{end}}types.{{internal .Name}}, len(t))
	for key := range t {
		v[key] = {{if .IsPrimitive}}t[key]{{else}}To{{internal .Name}}(t[key]){{end}}
	}
	return v
}
`))

var enumMapperTemplate = template.Must(template.New("enum mapper").Funcs(funcMap).Parse(`
// From{{internal .Name}} converts internal {{internal .Name}} type to thrift
func From{{internal .Name}}(t {{if .IsPointer}}*{{end}}types.{{internal .Name}}) {{if .IsPointer}}*{{end}}{{.ThriftPackage}}.{{.Name}} {
	{{if .IsPointer}}if t == nil {
		return nil
	}{{end}}
	switch {{if .IsPointer}}*{{end}}t { {{range .EnumValues}}
		case types.{{internal .}}: {{if $.IsPointer}}v := {{$.ThriftPackage}}.{{.}}; return &v{{else}}return {{$.ThriftPackage}}.{{.}}{{end}}{{end}}
	}
	panic("unexpected enum value")
}

// To{{internal .Name}} converts thrift {{.Name}} type to internal
func To{{internal .Name}}(t {{if .IsPointer}}*{{end}}{{.ThriftPackage}}.{{.Name}}) {{if .IsPointer}}*{{end}}types.{{internal .Name}} {
	{{if .IsPointer}}if t == nil {
		return nil
	}{{end}}
	switch {{if .IsPointer}}*{{end}}t { {{range .EnumValues}}
		case {{$.ThriftPackage}}.{{.}}: {{if $.IsPointer}}v := types.{{internal .}}; return &v{{else}}return types.{{internal .}}{{end}}{{end}}
	}
	panic("unexpected enum value")
}
`))

var sharedMapperAdditions = template.Must(template.New("shared mapper additions").Parse(`
// FromHistoryArray converts internal History array to thrift
func FromHistoryArray(t []*types.History) []*shared.History {
	if t == nil {
		return nil
	}
	v := make([]*shared.History, len(t))
	for i := range t {
		v[i] = FromHistory(t[i])
	}
	return v
}

// ToHistoryArray converts thrift History array to internal
func ToHistoryArray(t []*shared.History) []*types.History {
	if t == nil {
		return nil
	}
	v := make([]*types.History, len(t))
	for i := range t {
		v[i] = ToHistory(t[i])
	}
	return v
}
`))

var historyMapperAdditions = template.Must(template.New("history mapper additions").Parse(`
// FromProcessingQueueStateArrayMap converts internal ProcessingQueueState array map to thrift
func FromProcessingQueueStateArrayMap(t map[string][]*types.ProcessingQueueState) map[string][]*history.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make(map[string][]*history.ProcessingQueueState, len(t))
	for key := range t {
		v[key] = FromProcessingQueueStateArray(t[key])
	}
	return v
}

// ToProcessingQueueStateArrayMap converts thrift ProcessingQueueState array map to internal
func ToProcessingQueueStateArrayMap(t map[string][]*history.ProcessingQueueState) map[string][]*types.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make(map[string][]*types.ProcessingQueueState, len(t))
	for key := range t {
		v[key] = ToProcessingQueueStateArray(t[key])
	}
	return v
}
`))

type (
	// Type describes a type
	Type struct {
		FullThriftPackage string
		ThriftPackage     string
		Name              string
		IsPrimitive       bool
		IsArray           bool
		IsMap             bool
		IsPointer         bool
		IsStruct          bool
		IsEnum            bool
		MapKeyType        string
		Fields            []Field
		EnumValues        []string
		Prefix            string
	}

	// Field describe a field within a struct
	Field struct {
		Name string
		Type Type
		Tag  string
	}
)

var enumPointerExceptions = map[string]struct{}{
	"IndexedValueType": {},
}

func newMapType(m *types.Map) Type {
	t := newType(m.Elem())
	t.IsMap = true
	t.MapKeyType = m.Key().String()
	return t
}
func newPointerType(p *types.Pointer) Type {
	t := newType(p.Elem())
	t.IsPointer = true
	return t
}

func newSliceType(s *types.Slice) Type {
	t := newType(s.Elem())
	t.IsArray = true
	return t
}

func newNamedType(n *types.Named) Type {
	// This object has circular types definitions and causes stack overflow, we dont need it
	if n.Obj().Name() == "ThriftModule" {
		return Type{}
	}

	obj := n.Obj()
	pkg := obj.Pkg()
	t := newType(n.Underlying())
	t.Name = obj.Name()
	t.ThriftPackage = pkg.Name()
	t.FullThriftPackage = pkg.Path()
	if t.IsPrimitive {
		type enumConst struct {
			label string
			value int
		}
		enumConsts := []enumConst{}
		for _, name := range pkg.Scope().Names() {
			enumValue := pkg.Scope().Lookup(name)
			if isEnumValue(enumValue, n) {
				c := enumValue.(*types.Const)
				val, _ := strconv.Atoi(c.Val().String())
				enumConsts = append(enumConsts, enumConst{enumValue.Name(), val})
			}
		}
		if len(enumConsts) > 0 {
			t.IsPrimitive = false
			t.IsEnum = true
			if _, ok := enumPointerExceptions[t.Name]; !ok {
				t.IsPointer = true
			}
			sort.Slice(enumConsts, func(i, j int) bool {
				return enumConsts[i].value < enumConsts[j].value
			})
			for _, c := range enumConsts {
				t.EnumValues = append(t.EnumValues, c.label)
			}
		}
	}
	//TODO: fix this hack
	if t.Name == "IndexedValueType" {
		t.IsEnum = true
		t.IsPrimitive = false
	}
	if t.Name == "ContinueAsNewInitiator" {
		t.IsPrimitive = false
	}
	return t
}

func newStructType(s *types.Struct) Type {
	fields := make([]Field, s.NumFields())
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		fields[i] = Field{
			Name: f.Name(),
			Type: newType(f.Type()),
			Tag:  s.Tag(i),
		}
	}
	return Type{
		IsStruct: true,
		Fields:   fields,
	}
}

func newBasicType(b *types.Basic) Type {
	return Type{
		Name:        b.Name(),
		IsPrimitive: true,
	}
}

func newType(t types.Type) Type {
	switch tt := t.(type) {
	case *types.Map:
		return newMapType(tt)
	case *types.Pointer:
		return newPointerType(tt)
	case *types.Slice:
		return newSliceType(tt)
	case *types.Struct:
		return newStructType(tt)
	case *types.Basic:
		return newBasicType(tt)
	case *types.Named:
		return newNamedType(tt)
	case *types.Signature:
		// Dont care
		return Type{}
	}
	fmt.Printf("unexpected type: %v", t)
	return Type{}
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

// Package describes a thrift package to convert to internal types
type Package struct {
	ThriftPackage   string
	TypesFile       string
	MapperFile      string
	DuplicatePrefix string

	// the following two are a hack designed to make adding cases which are not caught
	// by the core script easier
	MapperAdditions *template.Template
	TypesAdditions  *template.Template
}

func main() {

	var (
		arrayMappers = map[string]Type{}
		mapMappers   = map[string]Type{}
		allNames     = map[string]int{}
	)

	packages := []Package{
		{
			ThriftPackage:   "github.com/uber/cadence/.gen/go/shared",
			TypesFile:       "common/types/shared.go",
			MapperFile:      "common/types/mapper/thrift/shared.go",
			MapperAdditions: sharedMapperAdditions,
		},
		{
			ThriftPackage: "github.com/uber/cadence/.gen/go/replicator",
			TypesFile:     "common/types/replicator.go",
			MapperFile:    "common/types/mapper/thrift/replicator.go",
		},
		{
			ThriftPackage:   "github.com/uber/cadence/.gen/go/history",
			TypesFile:       "common/types/history.go",
			MapperFile:      "common/types/mapper/thrift/history.go",
			DuplicatePrefix: "History",
			MapperAdditions: historyMapperAdditions,
		},
		{
			ThriftPackage:   "github.com/uber/cadence/.gen/go/admin",
			TypesFile:       "common/types/admin.go",
			MapperFile:      "common/types/mapper/thrift/admin.go",
			DuplicatePrefix: "Admin",
		},
		{
			ThriftPackage:   "github.com/uber/cadence/.gen/go/matching",
			TypesFile:       "common/types/matching.go",
			MapperFile:      "common/types/mapper/thrift/matching.go",
			DuplicatePrefix: "Matching",
		},
		{
			ThriftPackage: "github.com/uber/cadence/.gen/go/health",
			TypesFile:     "common/types/health.go",
			MapperFile:    "common/types/mapper/thrift/health.go",
		},
	}

	for _, p := range packages {
		currentArrayMappers := map[string]Type{}
		currentMapMappers := map[string]Type{}

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
			t := newType(obj.Type())
			if _, isConst := obj.(*types.Const); isConst {
				continue
			}
			if count, ok := allNames[t.Name]; ok {
				allNames[t.Name] = count + 1
			} else {
				allNames[t.Name] = 1
			}
		}

		for _, name := range pkg.Scope().Names() {
			obj := pkg.Scope().Lookup(name)
			if !obj.Exported() {
				continue
			}
			if _, isConst := obj.(*types.Const); isConst {
				continue
			}

			t := newType(obj.Type())

			if t.Name == "" {
				continue
			}

			if count, ok := allNames[t.Name]; ok && count > 1 {
				t.Prefix = p.DuplicatePrefix
			}

			if t.IsStruct {
				for i, f := range t.Fields {
					if count := allNames[f.Type.Name]; count > 1 && f.Type.FullThriftPackage == p.ThriftPackage {
						t.Fields[i].Type.Prefix = p.DuplicatePrefix
					}

					if f.Type.IsArray && !f.Type.IsPrimitive {
						currentArrayMappers[f.Type.Name] = f.Type
					}
					if f.Type.IsMap && !f.Type.IsPrimitive {
						currentMapMappers[f.Type.Name] = f.Type
					}
				}

				if err := structTemplate.Execute(typesFile, t); err != nil {
					panic(err)
				}
				if err := structMapperTemplate.Execute(mapperFile, t); err != nil {
					panic(err)
				}
			}
			if t.IsEnum {
				if err := enumTemplate.Execute(typesFile, t); err != nil {
					panic(err)
				}
				if err := enumMapperTemplate.Execute(mapperFile, t); err != nil {
					panic(err)
				}
			}
		}

		for name, m := range currentArrayMappers {
			if _, ok := arrayMappers[name]; !ok {
				if err := arrayMapperTemplate.Execute(mapperFile, m); err != nil {
					panic(err)
				}
				arrayMappers[name] = m
			}
		}
		for name, m := range currentMapMappers {
			if _, ok := mapMappers[name]; !ok {
				if err := mapMapperTemplate.Execute(mapperFile, m); err != nil {
					panic(err)
				}
				mapMappers[name] = m
			}
		}

		if p.TypesAdditions != nil {
			p.TypesAdditions.Execute(typesFile, nil)
		}
		if p.MapperAdditions != nil {
			p.MapperAdditions.Execute(mapperFile, nil)
		}

		typesFile.Close()
		mapperFile.Close()
	}
}
