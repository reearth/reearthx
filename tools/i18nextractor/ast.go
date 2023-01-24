package i18nextractor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const pkgName = `"github.com/reearth/reearthx/i18n"`

func extractMessages(buf []byte) ([]*i18n.Message, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", buf, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	extractor := newExtractor(file)
	ast.Walk(extractor, file)
	return extractor.Messages, nil
}

func newExtractor(file *ast.File) *extractor {
	return &extractor{PackageName: findPackageName(file)}
}

type extractor struct {
	PackageName string
	Messages    []*i18n.Message
}

func (e *extractor) Visit(node ast.Node) ast.Visitor {
	e.extractMessages(node)
	return e
}

func (e *extractor) extractMessages(node ast.Node) {
	if msg := extractMessagesCallerExpr(node, e.PackageName); msg != nil {
		e.Messages = append(e.Messages, msg)
		return
	}

	cl, ok := node.(*ast.CompositeLit)
	if !ok {
		return
	}

	switch t := cl.Type.(type) {
	case *ast.SelectorExpr:
		if !e.isMessageType(t) {
			return
		}
		if msg := extractMessageCompositeLit(cl); msg != nil {
			e.Messages = append(e.Messages, msg)
		}
	case *ast.ArrayType:
		if !e.isMessageType(t.Elt) {
			return
		}
		for _, el := range cl.Elts {
			ecl, ok := el.(*ast.CompositeLit)
			if !ok {
				continue
			}
			if msg := extractMessageCompositeLit(ecl); msg != nil {
				e.Messages = append(e.Messages, msg)
			}
		}
	case *ast.MapType:
		if !e.isMessageType(t.Value) {
			return
		}
		for _, el := range cl.Elts {
			kve, ok := el.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			vcl, ok := kve.Value.(*ast.CompositeLit)
			if !ok {
				continue
			}
			if msg := extractMessageCompositeLit(vcl); msg != nil {
				e.Messages = append(e.Messages, msg)
			}
		}
	}
}

func extractMessagesCallerExpr(node ast.Node, packageName string) (m *i18n.Message) {
	cl, ok := node.(*ast.CallExpr)
	if !ok || len(cl.Args) == 0 {
		return
	}
	se, ok := cl.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	if se.Sel.Name != "T" {
		return
	}
	x, ok := se.X.(*ast.Ident)
	if !ok || x.Name != packageName {
		return
	}
	id, ok := extractStringLiteral(cl.Args[0])
	if !ok || id == "" {
		return
	}
	return &i18n.Message{ID: id}
}

func (e *extractor) isMessageType(expr ast.Expr) bool {
	se := unwrapSelectorExpr(expr)
	if se == nil {
		return false
	}
	if se.Sel.Name != "Message" && se.Sel.Name != "LocalizeConfig" {
		return false
	}
	x, ok := se.X.(*ast.Ident)
	if !ok {
		return false
	}
	return x.Name == e.PackageName
}

func unwrapSelectorExpr(e ast.Expr) *ast.SelectorExpr {
	switch et := e.(type) {
	case *ast.SelectorExpr:
		return et
	case *ast.StarExpr:
		se, _ := et.X.(*ast.SelectorExpr)
		return se
	default:
		return nil
	}
}

func extractMessageCompositeLit(cl *ast.CompositeLit) (m *i18n.Message) {
	data := make(map[string]string)
	for _, elt := range cl.Elts {
		kve, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kve.Key.(*ast.Ident)
		if !ok {
			continue
		}
		v, ok := extractStringLiteral(kve.Value)
		if !ok {
			continue
		}
		data[key.Name] = v
	}
	if len(data) == 0 {
		return
	}
	if messageID := data["MessageID"]; messageID != "" {
		data["ID"] = messageID
	}
	return i18n.MustNewMessage(data)
}

func extractStringLiteral(expr ast.Expr) (string, bool) {
	switch v := expr.(type) {
	case *ast.BasicLit:
		if v.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(v.Value)
		if err != nil {
			return "", false
		}
		return s, true
	case *ast.BinaryExpr:
		if v.Op != token.ADD {
			return "", false
		}
		x, ok := extractStringLiteral(v.X)
		if !ok {
			return "", false
		}
		y, ok := extractStringLiteral(v.Y)
		if !ok {
			return "", false
		}
		return x + y, true
	case *ast.Ident:
		if v.Obj == nil {
			return "", false
		}
		switch z := v.Obj.Decl.(type) {
		case *ast.ValueSpec:
			if len(z.Values) == 0 {
				return "", false
			}
			s, ok := extractStringLiteral(z.Values[0])
			if !ok {
				return "", false
			}
			return s, true
		}
		return "", false
	default:
		return "", false
	}
}

func findPackageName(file *ast.File) string {
	for _, i := range file.Imports {
		if i.Path.Kind == token.STRING && i.Path.Value == pkgName {
			if i.Name == nil {
				return "i18n"
			}
			return i.Name.Name
		}
	}
	return ""
}
