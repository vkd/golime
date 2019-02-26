package tools

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"

	"github.com/pkg/errors"
)

type AddCommentsResult struct {
	Pos  int    `json:"pos"`
	Text string `json:"text"`
}

func AddComments(filename string) (out []AddCommentsResult, err error) {
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
			// if !d.Name.IsExported() {
			// 	continue
			// }
			// log.Printf("func: %v", d.Name.String())
			// if d.Doc == nil || len(d.Doc.List) == 0 {
			// 	log.Printf("add comment: %#v, %d", d.Doc, d.Pos())

			// 	out = append(out, AddCommentsResult{
			// 		Pos:  int(d.Pos()) - 1,
			// 		Text: "// " + d.Name.String() + " - ...\n",
			// 	})
			// }
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			if len(d.Specs) == 1 {
				if s, ok := d.Specs[0].(*ast.TypeSpec); ok {
					if docIsEmpty(d.Doc) {
						out = appendDoc(out, s.Name, s.Doc, d.Pos())
					}
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
