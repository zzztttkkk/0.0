package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"time"
)

func main() {
	begin := time.Now()

	files, err := filepath.Glob("../../apis/**/*.go")
	if err != nil {
		fmt.Println(err)
		return
	}

	var autoexports []string

	fs := token.NewFileSet()

	for _, fp := range files {
		node, err := parser.ParseFile(fs, fp, nil, parser.ParseComments)
		if err != nil {
			fmt.Println(err)
		}

		for _, decl := range node.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if ts.Name.Name != "AutoExport" {
					continue
				}

				if _, ok := ts.Type.(*ast.StructType); ok {
					autoexports = append(autoexports, fmt.Sprintf("%s.%s", node.Name.Name, ts.Name.Name))
				}
			}
		}
	}

	fmt.Println(time.Since(begin))
	fmt.Println(autoexports)
}
