// Copyright (C) 2009-2026 The Libphonenumber Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command allcheck verifies that every Go package has exactly one package doc
// comment.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type packageInfo struct {
	name     string
	docFiles int
}

func main() {
	root, err := moduleRoot()
	if err != nil {
		fail(err)
	}

	packages := make(map[string]*packageInfo)
	var problems []string
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(contents, []byte("Code generated ")) && bytes.Contains(contents, []byte("DO NOT EDIT.")) {
			return nil
		}

		set := token.NewFileSet()
		file, err := parser.ParseFile(set, path, contents, parser.ParseComments)
		if err != nil {
			return err
		}
		directory := filepath.Dir(path)
		info := packages[directory]
		if info == nil {
			info = &packageInfo{name: file.Name.Name}
			packages[directory] = info
		}
		if file.Doc != nil {
			info.docFiles++
		}
		problems = append(problems, exportedDocProblems(root, path, file)...)

		return nil
	})
	if err != nil {
		fail(err)
	}

	for directory, info := range packages {
		relative, err := filepath.Rel(root, directory)
		if err != nil {
			fail(err)
		}
		if info.docFiles != 1 {
			problems = append(problems, fmt.Sprintf("%s: package %s has %d package doc comments; want 1", relative, info.name, info.docFiles))
		}
	}
	sort.Strings(problems)
	for _, problem := range problems {
		if _, err := fmt.Fprintln(os.Stderr, problem); err != nil {
			os.Exit(2)
		}
	}
	if len(problems) != 0 {
		os.Exit(1)
	}
}

func fail(err error) {
	if _, writeErr := fmt.Fprintln(os.Stderr, err); writeErr != nil {
		os.Exit(2)
	}
	os.Exit(2)
}

func exportedDocProblems(root, path string, file *ast.File) []string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		relative = path
	}
	var problems []string
	check := func(name string, doc *ast.CommentGroup) {
		if !ast.IsExported(name) {
			return
		}
		if doc == nil || !strings.HasPrefix(strings.TrimSpace(doc.Text()), name) {
			problems = append(problems, fmt.Sprintf("%s: exported %s lacks a doc comment starting with its name", relative, name))
		}
	}
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			check(declaration.Name.Name, declaration.Doc)
		case *ast.GenDecl:
			for _, specification := range declaration.Specs {
				switch specification := specification.(type) {
				case *ast.TypeSpec:
					doc := specification.Doc
					if doc == nil && len(declaration.Specs) == 1 {
						doc = declaration.Doc
					}
					check(specification.Name.Name, doc)
				case *ast.ValueSpec:
					doc := specification.Doc
					if doc == nil && len(declaration.Specs) == 1 && len(specification.Names) == 1 {
						doc = declaration.Doc
					}
					for _, name := range specification.Names {
						check(name.Name, doc)
					}
				}
			}
		}
	}
	return problems
}

func moduleRoot() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("could not find go.mod from %s", directory)
		}
		directory = parent
	}
}
