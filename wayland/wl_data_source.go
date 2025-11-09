package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type WlDataSource struct {
	MimeTypes []string
	Actions   protocols.WlDataDeviceManagerDndAction_enum
}

func (w *WlDataSource) WlDataSource_offer(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlDataSource],
	mime_type string,
) {
	w.MimeTypes = append(w.MimeTypes, mime_type)
}

func (w *WlDataSource) WlDataSource_destroy(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlDataSource],
) bool {
	return true
}

func (w *WlDataSource) WlDataSource_set_actions(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlDataSource],
	dnd_actions protocols.WlDataDeviceManagerDndAction_enum,
) {
	w.Actions = dnd_actions
}

func (w *WlDataSource) OnBind(
	s protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	_ protocols.AnyObjectID,
	version uint32,
) {
	// TODO: Implement wl_data_source_on_bind
}

func MakeWlDataSource() *protocols.WlDataSource {
	ws := &WlDataSource{
		MimeTypes: []string{},
		Actions:   protocols.WlDataDeviceManagerDndAction_enum_none,
	}
	return &protocols.WlDataSource{
		Delegate: ws,
	}
}
