package tools

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"sort"
	"unicode/utf8"

	"github.com/pkg/errors"
)

type AddCommentsResult struct {
	Pos  int    `json:"pos"`
	Text string `json:"text"`
}

func AddComments(filename string, isRuneCount bool) (out []AddCommentsResult, err error) {
	// log.Printf("add comments on: %v", filename)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, errors.Wrap(err, "error on parse file")
	}

	for i := range file.Decls {
		switch d := file.Decls[i].(type) {
		case *ast.FuncDecl:
			out = appendDoc(out, d.Name, d.Doc, d.Pos())
		case *ast.GenDecl:
			if d.Tok == token.IMPORT {
				continue
			}
			if len(d.Specs) == 1 {
				if !docIsEmpty(d.Doc) {
					continue
				}
				switch s := d.Specs[0].(type) {
				case *ast.TypeSpec:
					out = appendDoc(out, s.Name, s.Doc, d.Pos())
				case *ast.ValueSpec:
					if len(s.Names) != 1 {
						continue
					}
					out = appendDoc(out, s.Names[0], s.Doc, d.Pos())
					// default:
					// 	log.Printf("Unknown spec: %T", s)
				}
				continue
			}
			for _, s := range d.Specs {
				switch s := s.(type) {
				case *ast.TypeSpec:
					out = appendDoc(out, s.Name, s.Doc, s.Pos())
					// default:
					// 	log.Printf("Unknown spec: %T", s)
				}
			}
			// default:
			// 	log.Printf("Unknown decl: %T", d)
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Pos > out[j].Pos })

	if isRuneCount {
		bs, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, errors.Wrap(err, "error on read file")
		}
		for i, p := range out {
			out[i].Pos = utf8.RuneCount(bs[:p.Pos])
		}
	}

	// log.Printf("out: %v", out)
	// log.Printf("----------------------OK----------------------")
	return
}

func appendDoc(out []AddCommentsResult, name *ast.Ident, doc *ast.CommentGroup, pos token.Pos) []AddCommentsResult {
	if !name.IsExported() {
		return out
	}
	if !docIsEmpty(doc) {
		return out
	}
	out = append(out, AddCommentsResult{
		Pos:  int(pos) - 1,
		Text: fmt.Sprintf("// %s ...\n", name),
	})
	return out
}

func docIsEmpty(doc *ast.CommentGroup) bool {
	return doc == nil || len(doc.List) == 0
}
