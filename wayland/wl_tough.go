package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type WlTouchDelegate struct {
}

func (w *WlTouchDelegate) WlTouch_release(s protocols.ClientState, object_id protocols.ObjectID[protocols.WlTouch]) bool {
	return true
}

func (w *WlTouchDelegate) OnBind(s protocols.ClientState, name protocols.AnyObjectID, interface_ string, new_id protocols.AnyObjectID, version_number uint32) {
	/** @TODO: Implement wl_touch_on_bind */
}

func MakeWlTouch() *protocols.WlTouch {
	return &protocols.WlTouch{
		Delegate: &WlTouchDelegate{},
	}
}
