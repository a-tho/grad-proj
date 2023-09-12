package generator

import (
	"context"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"
)

func funcDefinitionParams(ctx context.Context, fields []types.Variable) *jen.Statement {
	c := &jen.Statement{}
	c.ListFunc(func(g *jen.Group) {
		for _, field := range fields {
			g.Id(toLowerFirst(field.Name)).Add(fieldType(ctx, field.Type, true))
		}
	})
	return c
}

func fieldType(ctx context.Context, field types.Type, allowEllipsis bool) *jen.Statement {
	c := &jen.Statement{}

	imported := false

	for field != nil {
		switch f := field.(type) {
		case types.TImport:
			if f.Import != nil {
				if srcFile, ok := ctx.Value("code").(file); ok {
					if strings.HasSuffix(f.Import.Package, f.Import.Base.Name) {
						srcFile.ImportName(f.Import.Package, f.Import.Base.Name)
					} else {
						srcFile.ImportAlias(f.Import.Package, f.Import.Base.Name)
					}
					c.Qual(f.Import.Package, "")
				} else {
					c.Qual(f.Import.Package, "")
				}
				imported = true
			}
			field = f.Next
		case types.TName:
			if !imported && !types.IsBuiltin(f) {
			} else {
				c.Id(f.TypeName)
			}
			field = nil
		case types.TArray:
			if f.IsSlice {
				c.Index()
			} else if f.ArrayLen > 0 {
				c.Index(jen.Lit(f.ArrayLen))
			}
			field = f.Next
		case types.TMap:
			return c.Map(fieldType(ctx, f.Key, false)).Add(fieldType(ctx, f.Value, false))
		case types.TPointer:
			c.Op("*")
			field = f.Next
		case types.TInterface:
			mhds := interfaceType(ctx, f.Interface)
			return c.Interface(mhds...)
		case types.TEllipsis:
			if allowEllipsis {
				c.Op("...")
			} else {
				c.Index()
			}
			field = f.Next
		default:
			return c
		}
	}
	return c
}

func interfaceType(ctx context.Context, p *types.Interface) (code []jen.Code) {
	for _, x := range p.Methods {
		code = append(code, functionDefinition(ctx, x))
	}
	return
}

func functionDefinition(ctx context.Context, signature *types.Function) *jen.Statement {
	return jen.Id(signature.Name).
		Params(funcDefinitionParams(ctx, signature.Args)).
		Params(funcDefinitionParams(ctx, signature.Results))
}

func structField(ctx context.Context, field types.StructField) *jen.Statement {
	s := jen.Id(toUpperFirst(field.Name))

	s.Add(fieldType(ctx, field.Variable.Type, false))

	tags := map[string]string{"json": field.Name}

	for tag, values := range field.Tags {
		tags[tag] = strings.Join(values, ",")
	}
	s.Tag(tags)

	if types.IsEllipsis(field.Variable.Type) {
		s.Comment("This field was defined with ellipsis (...).")
	}
	return s
}
