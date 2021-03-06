package main

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	g := rpc{}
	protogen.Options{}.Run(g.Generate)
}

type rpc struct{}

// Generate generate service code
func (md *rpc) Generate(plugin *protogen.Plugin) error {
	for _, f := range plugin.Files {
		if len(f.Services) == 0 {
			continue
		}
		fileName := f.GeneratedFilenamePrefix + ".svr.go"
		t := plugin.NewGeneratedFile(fileName, f.GoImportPath)
		t.P("// Code generated by protoc-gen-corpc.")
		t.P()
		pkg := fmt.Sprintf("package %s", f.GoPackageName)
		t.P(pkg)
		t.P()
		for _, s := range f.Services {
			t.P(fmt.Sprintf(`%stype %sClient interface {`,
				getComments(s.Comments), s.Desc.Name()))
			for _, m := range s.Methods {
				funcCode := fmt.Sprintf(`	%s(args *%s) (*%s, error)`,
					m.Desc.Name(), m.Input.Desc.Name(), m.Output.Desc.Name())
				t.P(funcCode)
			}
			t.P("}")
			t.P()
			t.P(fmt.Sprintf(`type %sClient struct {
				cc *corpc.Client
			}`, unexport(s.GoName)))
			t.P()
			t.P("func New", s.Desc.Name(), "Client (cc *corpc.Client) ", s.Desc.Name(), "Client {")
			t.P("return &", unexport(s.GoName), "Client{cc}")
			t.P("}")
			t.P()
			for _, m := range s.Methods {
				funcCode := fmt.Sprintf(`func(c *%s) %s(args *%s) (*%s, error){
					out := new(%s)
					err := c.cc.Call("%s", args, out)
					if err != nil {
						return nil, err
					}
					return out,nil
				}
				`, unexport(s.GoName)+"Client", m.Desc.Name(), m.Input.Desc.Name(), m.Output.Desc.Name(),
					m.Output.Desc.Name(),
					s.Desc.Name()+"Service."+m.Desc.Name())
				t.P(funcCode)
			}

			t.P("type ", s.Desc.Name(), "Server interface {")
			for _, m := range s.Methods {
				funcCode := fmt.Sprintf(`	%s(*%s,*%s) error`,
					m.Desc.Name(), m.Input.Desc.Name(), m.Output.Desc.Name())
				t.P(funcCode)
			}
			t.P("}")
			t.P()
			t.P("func Register", s.Desc.Name(), "Server(s *corpc.Server,srv ", s.Desc.Name(), "Server) {")
			t.P(`s.RegisterName("`, s.Desc.Name(), `Service", srv)`)
			t.P("}")
			t.P()
		}
		for _, s := range f.Services {
			serviceCode := fmt.Sprintf(`%stype %s struct{}`,
				getComments(s.Comments), s.Desc.Name())
			t.P(serviceCode)
			t.P()
			for _, m := range s.Methods {
				funcCode := fmt.Sprintf(`%sfunc(s *%s) %s(args *%s,reply *%s) error {
					// define your service ...
					return nil
				}
				`, getComments(m.Comments), s.Desc.Name(),
					m.Desc.Name(), m.Input.Desc.Name(), m.Output.Desc.Name())
				t.P(funcCode)
			}
		}
	}
	return nil
}

// getComments get comment details
func getComments(comments protogen.CommentSet) string {
	c := make([]string, 0)
	c = append(c, strings.Split(string(comments.Leading), "\n")...)
	c = append(c, strings.Split(string(comments.Trailing), "\n")...)

	res := ""
	for _, comment := range c {
		if strings.TrimSpace(comment) == "" {
			continue
		}
		res += "//" + comment + "\n"
	}
	return res
}

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }
