package generator

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"

	"github.com/a-tho/grad-proj/internal/tags"
)

type method struct {
	*types.Function
	tags tags.Tags
	svc  *service

	inFields  []types.StructField
	outFields []types.StructField

	cookiesByArgs    map[string]string // cookies by input variables
	cookiesByResults map[string]string // cookies by output variables
	headersByArgs    map[string]string // headers by input variables
	queryArgsByArgs  map[string]string // query arguments by input variables
}

func newMethod(svc *service, f *types.Function) *method {
	m := &method{
		Function: f,
		tags:     tags.Parse(f.Docs),
		svc:      svc,
	}

	m.inFields = m.fields(m.insWithoutContext())
	m.outFields = m.fields(m.outsWithoutError())

	m.cookiesByArgs = make(map[string]string)
	m.headersByArgs = make(map[string]string)
	m.cookiesByResults = make(map[string]string)
	m.queryArgsByArgs = make(map[string]string)

	for arg, cookie := range m.valuesByTag(tagHTTPCookie) {
		if m.arg(arg) != nil {
			m.cookiesByArgs[arg] = cookie
		}
	}
	for arg, header := range m.valuesByTag(tagHTTPHeader) {
		if m.arg(arg) != nil {
			m.headersByArgs[arg] = header
		}
	}
	for result, cookie := range m.valuesByTag(tagHTTPCookie) {
		if m.result(result) != nil {
			m.cookiesByResults[result] = cookie
		}
	}
	for arg, queryArg := range m.valuesByTag(tagHTTPQuery) {
		if m.arg(arg) != nil {
			m.queryArgsByArgs[arg] = queryArg
		}
	}

	return m
}

func (m *method) restTransport(file file) jen.Code {
	return jen.Func().Params(
		jen.Id("tr").Op("*").Id(toLowerFirst(m.svc.Name)+"REST"),
	).Id("serve"+toUpperFirst(m.Name)).Params(
		jen.Id("w").Id("http").Dot("ResponseWriter"),
		jen.Id("r").Op("*").Id("http").Dot("Request"),
	).BlockFunc(func(g *jen.Group) {
		g.Var().Defs(
			jen.Id("err").Id("error"),
			jen.Id("successCode").Op("=").Id(m.successCode()),
		).Line()

		g.Id("request").Op(":=").StructFunc(func(g *jen.Group) {
			for _, arg := range m.insWithoutContext() {
				g.Id(toUpperFirst(arg.Name)).
					Id(arg.Type.String()).
					Tag(map[string]string{"json": toLowerFirst(arg.Name)})
			}
		}).Values().Line()

		if httpMethod := m.HTTPMethod(); httpMethod == "Post" || httpMethod == "Patch" { // request body is allowed
			g.If(jen.Id("err").Op("=").Id("json").Dot("NewDecoder").Params(jen.Id("r.Body")).Dot("Decode").Params(jen.Op("&").Id("request")).
				Op(";").Id("err").Op("!=").Nil()).Block(
				jen.Id("err").Op("=").Id("fmt").Dot("Errorf").Params(jen.Lit("failed to decode request body: %w"), jen.Id("err")),
				jen.Id("tr").Dot("writeResponse").Params(jen.Id("w"), jen.Id("err.Error()"), jen.Id("http.StatusBadRequest")),
				jen.Id("return"),
			).Line()
		}

		for arg, cookie := range m.cookiesByArgs {
			g.If(jen.List(jen.Id("cookie"), jen.Id("err")).Op(":=").Id("r").Dot("Cookie").Params(jen.Lit(cookie)).Op(";").Id("err").Op("==").Nil()).Block(
				jen.Id("request").Dot(toUpperFirst(arg)).Op("=").Id("cookie").Dot("Value"),
			)
		}
		if len(m.cookiesByArgs) > 0 {
			g.Line()
		}

		for arg, header := range m.headersByArgs {
			g.If(jen.List(jen.Id("headerVals"), jen.Id("ok")).Op(":=").Id("r").Dot("Header").Index(jen.Lit(header)).Op(";").Id("ok")).Block(
				jen.Id("request").Dot(toUpperFirst(arg)).Op("=").Id("headerVals").Index(jen.Lit(0)),
			)
		}
		if len(m.headersByArgs) > 0 {
			g.Line()
		}
		if len(m.queryArgsByArgs) > 0 {
			g.Id("queryVals").Op(":=").Id("r").Dot("URL").Dot("Query").Params()
			for arg, queryArg := range m.queryArgsByArgs {
				variable := m.arg(arg)
				if variable == nil {
					continue
				}
				argID := jen.Id(toLowerFirst(arg))
				typ := variable.Type
				typName := typ.String()
				switch t := typ.(type) {
				case types.TPointer:
					argID = jen.Op("&").Add(argID)
					typName = t.NextType().String()
				}
				g.If(jen.Id(toLowerFirst(arg) + "Raw").Op(":=").Id("queryVals").Dot("Get").Params(jen.Lit(queryArg)).Op(";").Id(toLowerFirst(arg) + "Raw").Op("!=").Lit("")).BlockFunc(func(g *jen.Group) {
					typeName := types.TypeName(typ)
					if typeName == nil {
						panic("invalid type for " + typName)
					}
					id := jen.Id(toLowerFirst(arg))
					raw := jen.Id(toLowerFirst(arg) + "Raw")
					switch *typeName {
					case "string":
						g.Add(id).Op(":=").Add(raw)
					case "int":
						g.List(id, jen.Err()).Op(":=").Qual(pkgStrconv, "Atoi").Call(raw)
						g.If(jen.Err().Op("!=").Nil()).Block(
							jen.Err().Op("=").Qual(pkgFmt, "Errorf").Params(jen.Lit(fmt.Sprintf("failed to decode query arguments (%s): %%w", toLowerFirst(arg))), jen.Err()),
							jen.Id("tr").Dot("writeResponse").Params(jen.Id("w"), jen.Err().Dot("Error").Params(), jen.Qual(pkgNetHTTP, "StatusBadRequest")),
							jen.Return(),
						)
					default:
						panic("only string/int query arguments are implemented for now")
					}
					g.Id("request").Dot(toUpperFirst(arg)).Op("=").Add(argID)
				})
			}
			g.Line()
		}

		ctx := context.WithValue(context.Background(), "code", file) // nolint

		var (
			results  []jen.Code
			response jen.Code = jen.Nil()
		)
		if len(m.outFields) > 0 {
			g.Id("response").Op(":=").StructFunc(func(g *jen.Group) {
				for _, out := range m.outFields {
					g.Add(structField(ctx, out))
				}
			}).Values().Line()

			for _, out := range m.outsWithoutError() {
				results = append(results, jen.Id("response").Dot(toUpperFirst(out.Name)))
			}

			response = jen.Id("response")
		}
		results = append(results, jen.Err())
		params := []jen.Code{jen.Id("r").Dot("Context").Params()}
		for _, arg := range m.insWithoutContext() {
			params = append(params, jen.Id("request").Dot(toUpperFirst(arg.Name)))
		}
		g.If(jen.List(results...).Op("=").Id("tr").Dot(toLowerFirst(m.Name)).Params(
			params...,
		).Op(";").Err().Op("!=").Nil()).Block(
			jen.Id("tr").Dot("writeResponse").Params(jen.Id("w"), jen.Err().Dot("Error").Params(), jen.Qual(pkgNetHTTP, "StatusInternalServerError")),
			jen.Return(),
		).Line()

		g.Id("tr").Dot("writeResponse").Params(jen.Id("w"), response, jen.Id("successCode"))
	})
}

func (m *method) middlewares(ctx context.Context) jen.Code {
	return jen.Type().Id(toLowerFirst(m.svc.Name) + toUpperFirst(m.Name)).Func().Params(funcDefinitionParams(ctx, m.Args)).Params(funcDefinitionParams(ctx, m.Results))
}

func (m *method) wrapMiddlewares() jen.Code {
	return jen.Type().Id("Wrap" + toUpperFirst(m.svc.Name) + toUpperFirst(m.Name)).Func().Params(jen.Id("next").Id(toLowerFirst(m.svc.Name) + toUpperFirst(m.Name))).Id(toLowerFirst(m.svc.Name) + toUpperFirst(m.Name))
}

func (m *method) valuesByTag(tag string) map[string]string {
	values := make(map[string]string)

	if valuesStr := m.tags.Value(tag); valuesStr != "" {
		valuePairs := strings.Split(valuesStr, ",")

		for _, pairStr := range valuePairs {
			if pair := strings.Split(pairStr, tagDelim); len(pair) == 2 {
				key := strings.TrimSpace(pair[0])
				value := strings.TrimSpace(pair[1])
				values[key] = value
			}
		}
	}
	return values
}

func (m *method) fields(variables []types.Variable) []types.StructField {
	var fields []types.StructField

	for _, variable := range variables {
		field := types.StructField{
			Variable: variable,
			Tags:     make(map[string][]string), // not actually used
		}

		fields = append(fields, field)
	}

	return fields
}

func (m *method) arg(name string) *types.Variable {
	for _, arg := range m.Args {
		if arg.Name == name {
			return &arg
		}
	}
	return nil
}

func (m *method) result(name string) *types.Variable {
	for _, result := range m.Results {
		if result.Name == name {
			return &result
		}
	}
	return nil
}

func (m *method) insWithoutContext() []types.Variable {
	if hasContextAsFirst(m.Args) {
		return m.Args[1:]
	}
	return m.Args
}

func (m *method) outsWithoutError() []types.Variable {
	if hasErrorAsLast(m.Results) {
		return m.Results[:len(m.Results)-1]
	}
	return m.Results
}

func hasContextAsFirst(vars []types.Variable) bool {
	if len(vars) == 0 {
		return false
	}
	name := types.TypeName(vars[0].Type)
	return name != nil && *name == "Context" &&
		types.TypeImport(vars[0].Type) != nil && types.TypeImport(vars[0].Type).Package == pkgContext
}

func hasErrorAsLast(vars []types.Variable) bool {
	if len(vars) == 0 {
		return false
	}
	name := types.TypeName(vars[len(vars)-1].Type)
	return name != nil && *name == "error" &&
		types.TypeImport(vars[len(vars)-1].Type) == nil
}

func (m *method) isValid() bool {
	return m.svc.tags.Contains(tagRESTServer) && m.tags.Contains(tagHTTPMethod)
}

func (m *method) HTTPMethod() string {
	switch strings.ToUpper(m.tags.Value(tagHTTPMethod)) {
	case "GET":
		return "Get"
	case "PUT":
		return "Put"
	case "PATCH":
		return "Patch"
	case "DELETE":
		return "Delete"
	case "OPTIONS":
		return "Options"
	default:
		return "Post"
	}
}

func (m *method) Path() string {
	parts := []string{"/"}
	prefix := m.svc.tags[tagHTTPPrefix]
	suffix := path.Join("/", toLowerFirst(m.svc.Name), toLowerFirst(m.Name))
	if customSuffix, ok := m.tags[tagHTTPPath]; ok {
		suffix = customSuffix
	}
	fullPath := path.Join(append(parts, prefix, suffix)...)
	return fullPath
}

func (m *method) successCode() string {
	successCode := defaultSuccessCode
	codeStr, ok := m.tags[tagHTTPSuccess]
	if !ok {
		return successCode
	}
	code, err := strconv.Atoi(codeStr)
	if err != nil {
		return successCode
	}
	if code > 299 || code < 200 {
		return successCode
	}
	successCode = codeStr
	return successCode
}
