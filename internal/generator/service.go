package generator

import (
	"context"
	"fmt"
	"path"
	"path/filepath"

	"github.com/dave/jennifer/jen"
	"github.com/rs/zerolog"
	"github.com/vetcher/go-astra/types"

	pathExt "github.com/a-tho/grad-proj/internal/path"
	"github.com/a-tho/grad-proj/internal/tags"
)

type service struct {
	types.Interface

	log zerolog.Logger

	restServer bool
	serviceDir string
	tags       tags.Tags
	methods    []*method
	tr         *Transport
	pkgPath    string
	pkgName    string
}

func newService(
	iface types.Interface,
	tags map[string]string,
	tr *Transport,
	log zerolog.Logger,
	filePath string,
	pkgName string,
) (*service, error) {
	log = log.With().Str("module", "service").Logger()
	log.Warn().Any("tags", tags).Msg("")
	_, restServer := tags[tagRESTServer]
	// other servers/clients (redis, jrpc, gprc etc)

	service := &service{
		Interface:  iface,
		log:        log,
		restServer: restServer,
		serviceDir: filepath.Base(filePath),
		tags:       tags,
		tr:         tr,
		pkgName:    pkgName,
	}

	for _, method := range iface.Methods {
		service.methods = append(service.methods, newMethod(service, method))
	}

	pkgPath, err := pathExt.PkgPath(filepath.Dir(filePath))
	log.Warn().Str("pkgPath", pkgPath).Msg("inside newService")
	if err != nil {
		log.Err(err).Msg("pathExt.PkgPath in newService")
		return nil, err
	}
	service.pkgPath = pkgPath

	return service, nil
}

func (s *service) generate(outPath string) error {
	if err := s.generateMiddleware(outPath); err != nil {
		s.log.Err(err).Msg("generateMiddleware")
		return err
	}

	if s.restServer {
		if err := s.generateREST(outPath); err != nil {
			s.log.Err(err).Msg("generateREST")
			return err
		}
	}

	return nil
}

func (s *service) generateMiddleware(outPath string) error {
	file := newFile(filepath.Base(outPath))

	s.log.Warn().Str("pkgPath", s.pkgPath).Str("pkgName", s.pkgName).Msg("inside generateMiddleware")
	file.PackageComment(fmt.Sprintf(genFileHeader, serverCmd, s.serviceDir))
	file.ImportName(pkgContext, "context")
	file.ImportName(s.pkgPath, s.pkgName)

	ctx := context.WithValue(context.Background(), "code", file) // nolint

	for _, method := range s.methods {
		file.Add(method.middlewares(ctx))
	}
	file.Line()

	file.Type().Id("Wrap"+toUpperFirst(s.Name)).Func().Params(jen.Id("next").Qual(s.pkgPath, s.Name)).Qual(s.pkgPath, s.Name).Line()

	for _, method := range s.methods {
		file.Add(method.wrapMiddlewares())
	}
	file.Line()

	return file.Save(path.Join(outPath, toLowerFirst(s.Name)+".middleware.xua.go"))
}

func (s *service) generateREST(outPath string) error {
	file := newFile(filepath.Base(outPath))

	file.PackageComment(fmt.Sprintf(genFileHeader, serverCmd, s.serviceDir))
	file.ImportName(pkgEncodingJSON, "json")
	file.ImportName(pkgFmt, "fmt")
	file.ImportName(pkgNetHTTP, "http")
	file.ImportName(pkgStrconv, "strconv")
	file.Line()
	file.ImportName(pkgZerolog, "zerolog")
	file.Line()
	file.ImportName(s.pkgPath, s.pkgName)

	file.Add(s.restType()).Line()

	file.Add(s.initRESTFunc()).Line()

	file.Add(s.addRoutesFunc()).Line()

	for _, method := range s.methods {
		file.Add(method.restTransport(file)).Line()
	}

	file.Add(s.writeResponse()).Line()

	return file.Save(path.Join(outPath, toLowerFirst(s.Name)+".server.xua.go"))
}

func (s *service) restType() jen.Code {
	return jen.Type().Id(toLowerFirst(s.Name) + "REST").StructFunc(func(g *jen.Group) {
		g.Id("log").Qual(pkgZerolog, "Logger").Line()
		g.Id("svc").Qual(s.pkgPath, s.Name).Line()
		for _, method := range s.methods {
			g.Id(toLowerFirst(method.Name)).Id(toLowerFirst(s.Name) + method.Name)
		}
	})
}

func (s *service) initRESTFunc() jen.Code {
	return jen.Func().Params(jen.Id("s").Op("*").Id("Server")).Id("Init"+s.Name+"Server").Params(jen.Id("svc").Qual(s.pkgPath, s.Name)).Block(
		jen.Id("s").Dot(toLowerFirst(s.Name)+"REST").Op("=").Op("&").Id(toLowerFirst(s.Name)+"REST").Values(jen.DictFunc(func(dict jen.Dict) {
			dict[jen.Id("log")] = jen.Id("s").Dot("log")
			dict[jen.Id("svc")] = jen.Id("svc")

			for _, method := range s.methods {
				dict[jen.Id(toLowerFirst(method.Name))] = jen.Id("svc").Dot(method.Name)
			}
		})),
		jen.Line().Id("s").Dot("AddRoutes"+s.Name).Call(),
	)
}

func (s *service) addRoutesFunc() jen.Code {
	return jen.Func().Params(jen.Id("s").Op("*").Id("Server")).Id("AddRoutes" + s.Name).Params().BlockFunc(func(g *jen.Group) {
		for _, method := range s.methods {
			if !method.isValid() {
				continue
			}
			g.Id("s").Dot("Mux").Dot(method.HTTPMethod()).Call(jen.Lit(method.Path()), jen.Id("s").Dot(toLowerFirst(s.Name)+"REST").Dot("serve"+toUpperFirst(method.Name)))
		}
	})
}

func (s *service) writeResponse() jen.Code {
	return jen.Func().Params(jen.Id("tr").Op("*").Id(toLowerFirst(s.Name)+"REST")).Id("writeResponse").Params(jen.Id("w").Qual(pkgNetHTTP, "ResponseWriter"), jen.Id("respBody").Id("any"), jen.Id("code").Id("int")).BlockFunc(func(g *jen.Group) {

		g.Id("w").Dot("Header").Params().Dot("Add").Params(jen.Lit("Content-Type"), jen.Lit("application/json"))
		g.Id("w").Dot("WriteHeader").Params(jen.Id("code"))
		g.If(jen.Id("respBody").Op("!=").Nil()).Block(
			jen.If(jen.Err().Op(":=").Qual(pkgEncodingJSON, "NewEncoder").Params(jen.Id("w")).Dot("Encode").Params(jen.Id("respBody"))).Op(";").Err().Op("!=").Nil().Block(
				jen.Id("tr").Dot("log").Dot("Error").Params().Dot("Err").Params(jen.Err()).Dot("Str").Params(jen.Lit("body"), jen.Qual(pkgFmt, "Sprintf").Params(jen.Lit("%+v"), jen.Id("respBody"))).Dot("Msg").Params(jen.Lit("failed to write into response body")),
			),
		).Else().Block(
			jen.Id("tr").Dot("log").Dot("Info").Params().Dot("Msg").Params(jen.Lit("no body to send in response")),
		)

	})
}
