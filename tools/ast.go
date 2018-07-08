package tools

import (
	"go/ast"
	"go/token"
)

func getImportDecl(file *ast.File) *ast.GenDecl {
	for _, d := range file.Decls {
		switch d := d.(type) {
		case *ast.GenDecl:
			if d.Tok == token.IMPORT {
				return d
			}
		}
	}
	return nil
}
