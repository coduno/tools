package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

func firstType(path string) (typeName string) {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		log.Fatal("parser:", err)
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				if typeSpec.Name != nil {
					return typeSpec.Name.Name
				}
			}
		}
	}
	return
}

func pluralize(singular string) (plural string) {
	if strings.HasSuffix(singular, "y") {
		return singular[:len(singular)-1] + "ies"
	}
	return singular + "s"
}

func funcMap(typeName string) template.FuncMap {
	return template.FuncMap{
		"t":     func() string { return "ƨ" },
		"type":  func() string { return typeName },
		"slice": func() string { return typeName + "s" },
		"kind":  func() string { return `"` + typeName + `"` },
	}
}

func main() {
	log.SetPrefix("generator: ")

	gofile := os.Getenv("GOFILE")
	funcs := funcMap(firstType(gofile))
	buf := new(bytes.Buffer)

	warning := fmt.Sprintf(`// This file was automatically generated from
//
//	%s
//
// by
//
//	%s
//
// DO NOT EDIT

`, gofile, strings.Join(os.Args, " "))

	impls, err := filepath.Glob("impl*.got")
	if err != nil {
		panic(err)
	}

	for _, p := range impls {
		t := template.New(path.Base(p))
		t.Funcs(funcs)
		if _, err := t.ParseFiles(p); err != nil {
			log.Fatal("template.ParseFiles: ", err.Error())
		}

		buf.Reset()
		buf.WriteString(warning)

		if err := t.Execute(buf, nil); err != nil {
			log.Fatal("template.Execute: ", err.Error())
		}

		target := strings.TrimSuffix(gofile, ".go") + "_" + strings.TrimSuffix(p, ".got") + ".go"

		if err := ioutil.WriteFile(target, buf.Bytes(), 0666); err != nil {
			log.Fatal("ioutil.WriteFile: ", err.Error())
		}

		exec.Command("gofmt", "-w", target).Run()
	}
}
