package main

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
)

type BuildProtocolOut struct {
	ProtocolFile string
	HelperFile   string
}

func buildProtocol(fs embed.FS, file string, protocolsPackageNameForHelper string, interfacesToGenHelpersFor []string) (BuildProtocolOut, error) {
	data, err := fs.ReadFile(filepath.Join("resources", file))
	if err != nil {
		return BuildProtocolOut{}, err
	}

	proto, err := UnmarshalProtocolXML(data)
	if err != nil {
		return BuildProtocolOut{}, fmt.Errorf("unmarshal %s: %w", file, err)
	}

	var out strings.Builder

	var helperOut strings.Builder

	helperTemplate := template.Must(template.New("helperGetObject").Parse(`
func Get{{.Name}}Object(cs {{.Pkg}}ClientState, id {{.Pkg}}ObjectID[{{.Pkg}}{{.Name}}]) *{{.Name}} {
    v := cs.GetObject({{.Pkg}}AnyObjectID(id))
    if v == nil {
        return nil
    }
    o := v.({{.Pkg}}WaylandObject[{{.Pkg}}{{.Name}}_delegate])
    d := o.GetDelegate()
    return d.(*{{.Name}})
}
`))

	for _, intf := range proto.Interfaces {
		fmt.Fprintf(&out, "type %s_delegate interface {\n", intf.Name)
		out.WriteString(genInterfaceInterface(intf))
		out.WriteString("}\n\n")

		fmt.Fprintf(&out, "type %s struct {\n", intf.Name)
		fmt.Fprintf(&out, "    Delegate %s_delegate\n", intf.Name)
		out.WriteString("}\n\n")

		get_delegate_template := `func (p *%s) GetDelegate() %s_delegate {
			return p.Delegate
		}
		`
		fmt.Fprintf(&out, get_delegate_template, intf.Name, intf.Name)

		get_bindable_template := `func (p *%s) GetBindable() OnBindable {
			return p.Delegate
		}
		`
		fmt.Fprintf(&out, get_bindable_template, intf.Name)

		// get_object_template := `
		// func Get%sObject(cs ClientState, id ObjectID[%s]) WaylandObject[%s_delegate] {
		// 	v := cs.GetObject(AnyObjectID(id))
		// 	if v == nil {
		// 		return nil
		// 	}
		// 	return v.(WaylandObject[%s_delegate])
		// }
		// `
		// fmt.Fprintf(&out, get_object_template, intf.Name, intf.Name, intf.Name, intf.Name)
		// out.WriteString("\n")

		out.WriteString(genEvents(intf))
		out.WriteString("\n")

		requestHandler := genRequestHandler(intf)

		fmt.Fprintf(&out, "func (p *%s) OnRequest(s FileDescriptorClaimClientState, message Message) {\n", intf.Name)

		if requestHandler != "" {
			out.WriteString("    _data_in_offset__ := 0\n")
			out.WriteString("    _ = _data_in_offset__\n")
			out.WriteString("    d := p.Delegate\n")
		}
		out.WriteString("    switch message.Opcode {\n")
		out.WriteString(requestHandler)
		out.WriteString("    default:\n")
		fmt.Fprintf(&out, "        fmt.Println(\"Unknown opcode on %s\", message.Opcode)\n", intf.Name)
		out.WriteString("    }\n")
		out.WriteString("}\n\n")

		out.WriteString(genEnums(intf))
		out.WriteString("\n")

		if len(interfacesToGenHelpersFor) == 0 || slices.Contains(interfacesToGenHelpersFor, intf.Name) {
			var buf bytes.Buffer
			_ = helperTemplate.Execute(&buf, struct {
				Name string
				Pkg  string
			}{
				Name: intf.Name,
				Pkg:  filepath.Base(protocolsPackageNameForHelper) + ".",
			})
			helperOut.WriteString(buf.String())
			helperOut.WriteString("\n")

		}

	}

	return BuildProtocolOut{
		ProtocolFile: out.String(),
		HelperFile:   helperOut.String(),
	}, nil
}
