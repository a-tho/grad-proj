package generator

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/rs/zerolog"
	"github.com/vetcher/go-astra"

	"github.com/a-tho/grad-proj/internal/tags"
)

type Transport struct {
	log        zerolog.Logger
	serviceDir string

	services map[string]*service
	tags     tags.Tags
}

func NewTransport(log zerolog.Logger, serviceDir string) (*Transport, error) {
	tr := &Transport{
		log:        log.With().Str("struct", "Transport").Logger(),
		serviceDir: serviceDir,
		services:   make(map[string]*service),
	}

	files, err := os.ReadDir(serviceDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}
		serviceDir, _ = filepath.Abs(serviceDir)
		filepath := path.Join(serviceDir, file.Name())
		serviceAST, err := astra.ParseFile(filepath)
		if err != nil {
			return nil, err
		}
		tr.tags = tags.Parse(serviceAST.Docs)
		for _, iface := range serviceAST.Interfaces {
			if tags := tags.Parse(iface.Docs); len(tags) != 0 {
				service, err := newService(iface, tags, tr, log, filepath, serviceAST.Name)
				if err != nil {
					return nil, err
				}
				tr.services[iface.Name] = service
			}
		}
	}
	return tr, nil
}

func (t *Transport) GenerateServer(outPath string) error {
	t.deleteGenFiles(outPath, genFilePrefix)

	if err := os.MkdirAll(outPath, 0777); err != nil {
		return err
	}

	if err := t.generateServer(outPath); err != nil {
		t.log.Err(err).Msg("generateServer")
	}

	for _, name := range t.serviceNames() {
		if err := t.services[name].generate(outPath); err != nil {
			t.log.Err(err).Msg("generateService")
		}
	}

	return nil
}

func (t *Transport) generateServer(outPath string) error {
	file := newFile(filepath.Base(outPath))

	file.PackageComment(fmt.Sprintf(genFileHeader, serverCmd, t.serviceDir))
	file.ImportName(pkgChi, "chi")
	file.ImportName(pkgZerolog, "zerolog")

	file.Line().Add(t.typeServer())
	file.Line().Add(t.funcNewServer())

	return file.save(path.Join(outPath, "server.xua.go"))
}

func (t *Transport) serviceNames() []string {
	var names []string
	for serviceName := range t.services {
		names = append(names, serviceName)
	}
	sort.Strings(names)
	return names
}

func (t *Transport) typeServer() jen.Code {
	return jen.Type().Id("Server").StructFunc(func(g *jen.Group) {
		g.Id("log").Qual(pkgZerolog, "Logger").Line()
		g.Id("").Op("*").Qual(pkgChi, "Mux").Line()
		for _, name := range t.serviceNames() {
			service := t.services[name]
			fieldName := toLowerFirst(name)
			if service.restServer {
				fieldName += suffixRESTField
			}
			g.Id(fieldName).Op("*").Id(fieldName)
		}
	})
}

func (t *Transport) funcNewServer() jen.Code {
	return jen.Func().Id("NewServer").Params(jen.Id("log").Qual(pkgZerolog, "Logger")).Op("*").Id("Server").
		BlockFunc(func(g *jen.Group) {
			g.Return().Op("&").Id("Server").Values(jen.DictFunc(func(dict jen.Dict) {
				dict[jen.Id("log")] = jen.Id("log")
				dict[jen.Id("Mux")] = jen.Qual(pkgChi, "NewRouter").Call()
			}))
		})
}
