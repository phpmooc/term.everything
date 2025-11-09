package wayland

import (
	"fmt"

	"github.com/mmulet/term.everything/wayland/protocols"
)

type XwaylandSurfaceV1 struct {
}

func (x *XwaylandSurfaceV1) XwaylandSurfaceV1_set_serial(
	_s protocols.ClientState,
	_object_id protocols.ObjectID[protocols.XwaylandSurfaceV1],
	_serial_lo uint32,
	_serial_hi uint32,
) {
	/** @TODO: Implement XwaylandSurfaceV1_set_serial */
}

func (x *XwaylandSurfaceV1) XwaylandSurfaceV1_destroy(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.XwaylandSurfaceV1],
) bool {
	surface := GetSurfaceFromRole(s, object_id)

	if surface == nil || surface.Role == nil {
		fmt.Printf("XwaylandSurfaceV1_destroy: surface not found")
		return true
	}
	if role, ok := surface.Role.(*SurfaceRoleXWaylandSurface); ok {
		role.Data = nil
	}
	return true
}

func (x *XwaylandSurfaceV1) XwaylandSurfaceV1_on_bind(
	_s protocols.ClientState,
	_name protocols.AnyObjectID,
	_interface_ string,
	_new_id protocols.AnyObjectID,
	_version_number uint32,
) {
}

func (x *XwaylandSurfaceV1) OnBind(
	cs protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
}

func MakeXwaylandSurfaceV1() *protocols.XwaylandSurfaceV1 {
	return &protocols.XwaylandSurfaceV1{
		Delegate: &XwaylandSurfaceV1{},
	}
}
