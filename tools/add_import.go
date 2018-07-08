package tools

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// AddImportResult - result of add_import command
type AddImportResult struct {
	// L/Rpos - position on original file
	// it means old import block declaration
	Lpos int64 `json:"l_pos"`
	Rpos int64 `json:"r_pos"`

	// Text for replace on original place with new import block
	Text string `json:"text"`
}

// AddImport - add import to go source file
func AddImport(filename string, importName string) (*AddImportResult, error) {
	fl, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		return nil, errors.Wrap(err, "error on open file")
	}
	defer fl.Close()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, fl, parser.ImportsOnly)
	if err != nil {
		return nil, errors.Wrap(err, "error on parse file")
	}

	oldImport := getImportDecl(file)

	var res = &AddImportResult{}

	// make border of old import block
	if oldImport == nil {
		// add next '\n' char
		res.Lpos = int64(file.Name.End() + 1)
		// add extra '\n' char on end
		res.Rpos = int64(file.Name.End() - 1)
	} else {
		res.Lpos = int64(oldImport.Pos() - 1)
		res.Rpos = int64(oldImport.End() - 1)
	}

	imp := addImportAst(importName, file)
	ast.SortImports(fset, file)

	var bs bytes.Buffer
	err = format.Node(&bs, fset, imp)
	if err != nil {
		return nil, errors.Wrap(err, "error on format import to string")
	}

	res.Text = bs.String()

	return res, nil
}

func addImportAst(importName string, file *ast.File) *ast.GenDecl {
	importSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Value: strconv.Quote(importName)},
	}

	for _, d := range file.Decls {
		switch d := d.(type) {
		case *ast.GenDecl:
			if d.Tok == token.IMPORT {
				// for case is `import "fmt"`
				// needs set not zero Lparen
				if len(d.Specs) == 1 && d.Lparen == 0 {
					d.Lparen = d.Specs[0].Pos()
				}

				// skip if import already exists
				for _, s := range d.Specs {
					if is, ok := s.(*ast.ImportSpec); ok {
						if is.Path.Value == importSpec.Path.Value {
							return d
						}
					}
				}

				if strings.ContainsRune(importName, '.') {
					// importSpec.Path.ValuePos = d.Specs[len(d.Specs)-1].Pos() + 1
					d.Specs = append(d.Specs, importSpec)

				} else {
					d.Specs = append([]ast.Spec{importSpec}, d.Specs...)
				}

				return d
			}
		}
	}
	d := &ast.GenDecl{
		Tok: token.IMPORT,
		// TokPos: file.Name.End(),
		Specs: []ast.Spec{importSpec},
	}
	file.Decls = append([]ast.Decl{d}, file.Decls...)
	return d
}
