package gencode

import (
	"go/build"
	"io/ioutil"

	"golang.org/x/tools/imports"
)

func SolveGoimports(fn string) error {
	src, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	opt := &imports.Options{Comments: true, TabIndent: true, TabWidth: 4}
	res, err := imports.Process(fn, src, opt)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fn, res, 0666)
}

func ImportPath(path string) (string, error) {
	absPath, err := AbsPath(path)
	if err != nil {
		return "", err
	}

	pkg, err := build.ImportDir(absPath, build.IgnoreVendor)
	if err != nil {
		return "", err
	}

	return pkg.ImportPath, nil
}
