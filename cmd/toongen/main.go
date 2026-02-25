package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// StructInfo holds info about a struct to generate
type StructInfo struct {
	Package    string
	Name       string
	Fields     []FieldInfo
	HeaderHash uint64
}

// FieldInfo holds info about a struct field
type FieldInfo struct {
	Name     string
	Type     string
	JSONName string
	ToonName string
}

const toonTmpl = `
// toonSchemaHash{{.Name}} is the pre-computed header hash for fast validation
const toonSchemaHash{{.Name}} = {{.HeaderHash}}

// toonHeader{{.Name}} is the expected header string
var toonHeader{{.Name}} = "{{.Name | lower}}{" + "{{range $i, $f := .Fields}}{{if $i}},{{end}}{{$f.ToonName}}{{end}}}:" + ""

// ToonTokenCount returns estimated token count for LLM context
func (s {{.Name}}) ToonTokenCount() int {
	count := {{len .Fields}} + 4 // separators + header overhead
	{{- range $f := .Fields}}
	{{- if eq $f.Type "int"}}
	count += estimateIntTokens(s.{{$f.Name}})
	{{- else if eq $f.Type "string"}}
	count += len(s.{{$f.Name}}) / 4
	{{- else if eq $f.Type "bool"}}
	count += 1
	{{- end}}
	{{- end}}
	return count
}

func (s {{.Name}}) MarshalTOON() ([]byte, error) {
	buf := make([]byte, 0, 256)
	buf = append(buf, toonHeader{{.Name}}...)
	{{- range $i, $f := .Fields}}
	{{- if $i}}
	buf = append(buf, ',')
	{{- end}}
	{{template "marshalField" $f}}
	{{- end}}
	return buf, nil
}

func (s *{{.Name}}) UnmarshalTOON(data []byte) error {
	var commaIdx int
	colonIdx := -1
	for i, b := range data {
		if b == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx == -1 {
		return strconv.ErrSyntax
	}
	if colonIdx+1 != len(toonHeader{{.Name}}) {
		return strconv.ErrSyntax
	}
	h := fnv.New64a()
	h.Write(data[:colonIdx])
	if h.Sum64() != toonSchemaHash{{.Name}} {
		return strconv.ErrSyntax
	}
	data = data[colonIdx+1:]
	{{- range $i, $f := .Fields}}
	{{- if eq $f.Type "int"}}
	{{template "parseInt" $f}}
	{{- else if eq $f.Type "string"}}
	{{template "parseString" $f}}
	{{- else if eq $f.Type "bool"}}
	{{template "parseBool" $f}}
	{{- else if eq $f.Type "float64"}}
	{{template "parseFloat" $f}}
	{{- end}}
	{{- end}}
	return nil
}
`

const marshalFieldTmpl = `{{define "marshalField"}}
{{- if eq .Type "int"}}
buf = strconv.AppendInt(buf, int64(s.{{.Name}}), 10)
{{- else if eq .Type "string"}}
buf = append(buf, s.{{.Name}}...)
{{- else if eq .Type "bool"}}
if s.{{.Name}} {
	buf = append(buf, '+')
} else {
	buf = append(buf, '-')
}
{{- else if eq .Type "float64"}}
buf = strconv.AppendFloat(buf, s.{{.Name}}, 'f', -1, 64)
{{- end}}
{{end}}`

const parseIntTmpl = `{{define "parseInt"}}
commaIdx = -1
for i, b := range data {
	if b == ',' {
		commaIdx = i
		break
	}
}
if commaIdx == -1 {
	commaIdx = len(data)
}
{{.Name}}Val, _ := strconv.ParseInt(string(data[:commaIdx]), 10, 64)
s.{{.Name}} = int({{.Name}}Val)
if commaIdx < len(data) {
	data = data[commaIdx+1:]
}
{{end}}`

const parseStringTmpl = `{{define "parseString"}}
commaIdx = -1
for i, b := range data {
	if b == ',' {
		commaIdx = i
		break
	}
}
if commaIdx == -1 {
	commaIdx = len(data)
}
s.{{.Name}} = string(data[:commaIdx])
if commaIdx < len(data) {
	data = data[commaIdx+1:]
}
{{end}}`

const parseFloatTmpl = `{{define "parseFloat"}}
commaIdx = -1
for i, b := range data {
	if b == ',' {
		commaIdx = i
		break
	}
}
if commaIdx == -1 {
	commaIdx = len(data)
}
{{.Name}}Val, _ := strconv.ParseFloat(string(data[:commaIdx]), 64)
s.{{.Name}} = {{.Name}}Val
if commaIdx < len(data) {
	data = data[commaIdx+1:]
}
{{end}}`

const parseBoolTmpl = `{{define "parseBool"}}
if len(data) > 0 && (data[0] == '+' || data[0] == '1' || data[0] == 't' || data[0] == 'T') {
	s.{{.Name}} = true
} else {
	s.{{.Name}} = false
}
commaIdx = -1
for i, b := range data {
	if b == ',' {
		commaIdx = i
		break
	}
}
if commaIdx == -1 {
	commaIdx = len(data)
}
if commaIdx < len(data) {
	data = data[commaIdx+1:]
}
{{end}}`

func main() {
	var (
		input  = flag.String("i", ".", "input directory")
		output = flag.String("o", "", "output file (default: <input>/_toon.go)")
	)
	flag.Parse()

	if *output == "" {
		*output = filepath.Join(*input, "_toon.go")
	}

	structs, pkgName, err := parseStructs(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing: %v\n", err)
		os.Exit(1)
	}

	if len(structs) == 0 {
		fmt.Println("No structs with //toon:generate found")
		return
	}

	if err := generate(*output, pkgName, structs); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s with %d structs\n", *output, len(structs))
}

func parseStructs(dir string) ([]StructInfo, string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, "", err
	}

	var structs []StructInfo
	var pkgName string

	for name, pkg := range pkgs {
		if strings.HasSuffix(name, "_test") {
			continue
		}
		pkgName = name

		for filename, file := range pkg.Files {
			if strings.HasSuffix(filename, "_toon.go") || strings.HasSuffix(filename, ".gen.go") {
				continue
			}

			ast.Inspect(file, func(n ast.Node) bool {
				genDecl, ok := n.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					return true
				}

				// Check for //toon:generate comment
				if !hasToonGenerate(genDecl.Doc) {
					return true
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					fields := extractFields(structType)
					info := StructInfo{
						Name:       typeSpec.Name.Name,
						Fields:     fields,
						HeaderHash: computeHeaderHash(typeSpec.Name.Name, fields),
					}
					structs = append(structs, info)
				}
				return true
			})
		}
	}

	return structs, pkgName, nil
}

func hasToonGenerate(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, comment := range doc.List {
		if strings.Contains(comment.Text, "toon:generate") {
			return true
		}
	}
	return false
}

func extractFields(st *ast.StructType) []FieldInfo {
	var fields []FieldInfo
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue
		}

		name := field.Names[0].Name
		if !ast.IsExported(name) {
			continue
		}

		fieldType := exprToString(field.Type)
		toonName := name
		if tag := field.Tag; tag != nil {
			tagValue := strings.Trim(tag.Value, "`")
			if val := extractTag(tagValue, "toon"); val != "" && val != "-" {
				toonName = val
			}
		}

		fields = append(fields, FieldInfo{
			Name:     name,
			Type:     fieldType,
			ToonName: strings.ToLower(toonName),
		})
	}
	return fields
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func extractTag(tag, key string) string {
	// Simple tag parser: `key:"value"`
	prefix := key + ":\""
	start := strings.Index(tag, prefix)
	if start == -1 {
		return ""
	}
	start += len(prefix)
	end := strings.Index(tag[start:], "\"")
	if end == -1 {
		return ""
	}
	return tag[start : start+end]
}

// computeHeaderHash computes FNV-1a hash of the expected header (without trailing colon)
func computeHeaderHash(name string, fields []FieldInfo) uint64 {
	h := fnv.New64a()
	h.Write([]byte(strings.ToLower(name)))
	h.Write([]byte("{"))
	for i, f := range fields {
		if i > 0 {
			h.Write([]byte(","))
		}
		h.Write([]byte(f.ToonName))
	}
	h.Write([]byte("}"))
	return h.Sum64()
}

func generate(filename, pkgName string, structs []StructInfo) error {
	funcs := template.FuncMap{
		"lower": strings.ToLower,
	}

	tmpl, err := template.New("toon").Funcs(funcs).Parse(toonTmpl)
	if err != nil {
		return err
	}

	// Parse sub-templates
	tmpl, err = tmpl.Parse(marshalFieldTmpl)
	if err != nil {
		return err
	}
	tmpl, err = tmpl.Parse(parseIntTmpl)
	if err != nil {
		return err
	}
	tmpl, err = tmpl.Parse(parseStringTmpl)
	if err != nil {
		return err
	}
	tmpl, err = tmpl.Parse(parseFloatTmpl)
	if err != nil {
		return err
	}
	tmpl, err = tmpl.Parse(parseBoolTmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write package header once
	fmt.Fprintf(file, "// Code generated by toongen. DO NOT EDIT.\n\n")
	fmt.Fprintf(file, "package %s\n\n", pkgName)
	fmt.Fprintf(file, "import (\n\t\"hash/fnv\"\n\t\"strconv\"\n)\n\n")
	fmt.Fprintf(file, "func estimateIntTokens(n int) int {\n")
	fmt.Fprintf(file, "\tif n == 0 { return 1 }\n")
	fmt.Fprintf(file, "\tcount := 0\n")
	fmt.Fprintf(file, "\tif n < 0 { count++; n = -n }\n")
	fmt.Fprintf(file, "\tfor n > 0 { count++; n /= 10 }\n")
	fmt.Fprintf(file, "\treturn count\n}\n")

	for _, s := range structs {
		s.Package = pkgName
		if err := tmpl.Execute(file, s); err != nil {
			return err
		}
	}

	return nil
}
