package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func glob(dir string, ext string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func main() {
	begin := time.Now()

	files, err := glob("../../apis", ".go")
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
					autoexports = append(autoexports, node.Name.Name)
				}
			}
		}
	}

	f, err := os.OpenFile("./autoload.go", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_ = f.Truncate(0)

	sb := strings.Builder{}
	sb.WriteString(`// Code generated .* DO NOT EDIT.
package main

import (
`)

	for i, name := range autoexports {
		sb.WriteString(fmt.Sprintf("\t_%d \"github.com/zzztttkkk/0.0/apis/%s\"\r\n", i, name))
	}
	sb.WriteString(")\r\n\r\nvar (\r\n")

	for i := range autoexports {
		sb.WriteString(fmt.Sprintf("\t_ _%d.AutoExport\r\n", i))
	}
	sb.WriteString(")\r\n")

	if _, err := f.WriteString(sb.String()); err != nil {
		panic(err)
	}

	fmt.Printf("AutoLoadDone %s", time.Since(begin))
}
