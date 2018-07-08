package tools

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"path"
	"strings"
)

func GetAllImportPaths() ([]string, error) {
	var imports []string
	for p := range parseDir(path.Join(build.Default.GOROOT, "src"), skipVendor) {
		if p.err != nil {
			return nil, fmt.Errorf("Error on parse dir (%s): %v", build.Default.GOROOT, p.err)
		}

		imports = append(imports, p.path)
	}

	for p := range parseDir(path.Join(build.Default.GOPATH, "src")) {
		if p.err != nil {
			return nil, fmt.Errorf("Error on parse dir (%s): %v", build.Default.GOPATH, p.err)
		}

		imports = append(imports, p.path)
	}

	return imports, nil
}

type importFunc func(i importPath)
type importFuncOverride func(fn importFunc) importFunc

func override(fn importFunc, fns ...importFuncOverride) importFunc {
	outFn := fn
	for _, f := range fns {
		outFn = f(outFn)
	}
	return outFn
}

func skipVendor(fn importFunc) importFunc {
	return func(i importPath) {
		if i.isVendor {
			return
		}
		fn(i)
	}
}

type importPath struct {
	path string
	err  error

	isVendor bool
}

func parseDir(pathDir string, oo ...importFuncOverride) <-chan importPath {
	out := make(chan importPath)

	cropIndex := len(pathDir) + 1
	sendImportFunc := override(func(i importPath) {
		i.path = i.path[cropIndex:]
		out <- i
	}, oo...)

	go func() {
		err := parsingDir(pathDir, sendImportFunc)
		if err != nil {
			out <- importPath{err: err}
		}
		close(out)
	}()
	return out
}

func parsingDir(pathDir string, fn importFunc) error {
	fs, err := ioutil.ReadDir(pathDir)
	if err != nil {
		return err
	}
	// current dir has '.go' files
	var name string
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		name = f.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			fn(importPath{path: pathDir})
			break
		}
	}

	// walk dirs
	for _, f := range fs {
		if !f.IsDir() {
			continue
		}
		switch f.Name() {
		case "testdata", "cmd":
			// skip
		case "vendor":
			err = parsingDir(path.Join(pathDir, f.Name()), vendor(fn))
		default:
			err = parsingDir(path.Join(pathDir, f.Name()), fn)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func vendor(fn importFunc) importFunc {
	return func(i importPath) {
		i.isVendor = true
		fn(i)
	}
}
