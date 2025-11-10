package main

import (
	"fmt"
	"strings"
)

func genInterfaceInterface(iface Interface) string {
	on_bind_declaration := "OnBind(s ClientState, name AnyObjectID, interface_ string, new_id AnyObjectID, version_number uint32)\n"
	if len(iface.Requests) == 0 {
		return on_bind_declaration
	}

	var b strings.Builder

	for _, req := range iface.Requests {
		params := []string{
			"s ClientState",
			fmt.Sprintf("object_id ObjectID[%s]", iface.Name),
		}
		for _, a := range req.Args {
			params = append(params, generateGoType(iface.Name, a, false))
		}

		methodName := fmt.Sprintf("%s_%s", iface.Name, req.Name)
		signature := fmt.Sprintf("%s(%s)", methodName, strings.Join(params, ", "))

		if req.Name == "destroy" || req.Name == "release" {
			signature += " bool"
		}

		b.WriteString(signature)
		b.WriteByte('\n')
	}

	// b.WriteString(fmt.Sprintf(
	// 	// "%s_on_bind(s ClientState, name ObjectID[%s], interface_ string, new_id ObjectID[%s], version_number uint32)\n",
	// 	"%s_on_bind(s ClientState, name ObjectID[%s], interface_ string, new_id ObjectID[%s], version_number uint32)\n",
	// 	iface.Name, iface.Name, iface.Name,
	// ))

	b.WriteString(on_bind_declaration)

	switch iface.Name {
	case "WlKeyboard":
		b.WriteString("AfterGetKeyboard(s ClientState, object_id ObjectID[WlKeyboard])\n")
	case "WlPointer":
		b.WriteString("AfterGetPointer(s ClientState, object_id ObjectID[WlPointer])\n")
	}

	return b.String()
}
