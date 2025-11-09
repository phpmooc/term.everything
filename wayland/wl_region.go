package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

func auto_release(s protocols.ClientState,
	id protocols.ObjectID[protocols.WlRegion]) bool {
	return true
}

type WlRegion struct {
}

func (r *WlRegion) WlRegion_destroy(
	s protocols.ClientState,
	id protocols.ObjectID[protocols.WlRegion],
) bool {
	return true
}

func (r *WlRegion) WlRegion_add(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlRegion],
	x int32,
	y int32,
	width int32,
	height int32,
) {
	// TODO: Implement proper region union semantics.
}

func (r *WlRegion) WlRegion_subtract(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlRegion],
	_x int32,
	_y int32,
	_width int32,
	_height int32,
) {
	// TODO: Implement proper region subtraction semantics.
}

func (r *WlRegion) OnBind(
	s protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
	// No-op for wl_region
}

func MakeWlRegion() *protocols.WlRegion {
	return &protocols.WlRegion{
		Delegate: &WlRegion{},
	}
}
