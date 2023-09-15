package generator

import (
	"os/exec"

	"github.com/dave/jennifer/jen"
)

type file struct {
	*jen.File
}

func newFile(packageName string) file {
	return file{File: jen.NewFile(packageName)}
}

func (f *file) save(filename string) error {
	if err := f.Save(filename); err != nil {
		return err
	}

	return f.format(filename)
}

func (f *file) format(filename string) error {
	path, err := exec.LookPath("goimports")
	if err != nil {
		return nil
	}
	return exec.Command(path, "-local", "-w", filename).Run()
}
