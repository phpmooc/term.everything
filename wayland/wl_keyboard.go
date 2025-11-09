package wayland

import (
	_ "embed"
	"os"

	"github.com/mmulet/term.everything/wayland/protocols"
)

//go:embed resources/server-1.xkb
var xkbKeymapData []byte

type WlKeyboard struct {
	Key_map_fd   protocols.FileDescriptor
	Key_map_size uint32
	// let's prevent the garbage collector from closing the file descriptor
	File *os.File
}

func (o *WlKeyboard) WlKeyboard_release(s protocols.ClientState, _ protocols.ObjectID[protocols.WlKeyboard]) bool {
	return true
}
func (o *WlKeyboard) OnBind(
	s protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	_ protocols.AnyObjectID,
	version uint32,
) {
}

func (o *WlKeyboard) AfterGetKeyboard(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlKeyboard],
) {
	protocols.WlKeyboard_keymap(
		s,
		object_id,
		protocols.WlKeyboardKeymapFormat_enum_xkb_v1,
		o.Key_map_fd,
		o.Key_map_size,
	)
}

func MakeWlKeyboard() *protocols.WlKeyboard {

	f, err := os.CreateTemp(os.TempDir(), "xkb-keymap-*.xkb")
	if err != nil {
		panic(err)
	}
	if _, werr := f.Write(xkbKeymapData); werr != nil {
		f.Close()
		panic(werr)
	}
	if _, serr := f.Seek(0, 0); serr != nil {
		f.Close()
		panic(serr)
	}
	return &protocols.WlKeyboard{
		Delegate: &WlKeyboard{
			Key_map_fd:   protocols.FileDescriptor(f.Fd()),
			Key_map_size: uint32(len(xkbKeymapData)),
			File:         f,
		},
	}

}
