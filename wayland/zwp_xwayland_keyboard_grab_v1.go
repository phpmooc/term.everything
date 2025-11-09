package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type ZwpXwaylandKeyboardGrabV1 struct {
}

func (g *ZwpXwaylandKeyboardGrabV1) ZwpXwaylandKeyboardGrabV1_destroy(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabV1],
) bool {
	/**
	 * @TODO ungrab the keyboard
	 */
	return true
}

func (g *ZwpXwaylandKeyboardGrabV1) OnBind(
	_ protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	_ protocols.AnyObjectID,
	_ uint32,
) {
}

func MakeZwpXwaylandKeyboardGrabV1() *protocols.ZwpXwaylandKeyboardGrabV1 {
	return &protocols.ZwpXwaylandKeyboardGrabV1{
		Delegate: &ZwpXwaylandKeyboardGrabV1{},
	}
}
