package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type XdgPopup struct {
	Version uint32
	Parent  *protocols.ObjectID[protocols.XdgSurface]
	State   XdgPositionerState
	/**
	 * Only one instead of a queue because if multiple
	 * position requests are sent we can ignore
	 * all but the last one
	 */
	pendingPosition       *XdgPositionerState
	pendingPositionSerial *uint32
}

func (x *XdgPopup) XdgPopup_destroy(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.XdgPopup],
) bool {
	surface := GetSurfaceFromRole(s, object_id)
	UnregisterRoleToSurface(s, object_id)
	if surface == nil {
		return true
	}
	if surface.Role == nil {
		return true
	}

	_, isPopup := surface.Role.(*SurfaceRoleXdgPopup)
	if !isPopup {
		return true
	}
	surface.ClearRoleData()
	return true
}

func (x *XdgPopup) XdgPopup_grab(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgPopup],
	_ protocols.ObjectID[protocols.WlSeat],
	_ uint32,
) {
	/** @TODO: Implement xdg_popup_grab */
}

func (x *XdgPopup) XdgPopup_reposition(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.XdgPopup],
	positionerID protocols.ObjectID[protocols.XdgPositioner],
	token uint32,
) {
	positioner := GetXdgPositionerObject(s, positionerID)
	if positioner == nil {
		return
	}
	new_state := positioner.state
	st := new_state // copy
	x.pendingPosition = &st
	x.pendingPositionSerial = &token

	protocols.XdgPopup_repositioned(s, x.Version, object_id, token)

	/**
	 * @TODO figure out what
	 * these values are
	 */
	protocols.XdgPopup_configure(s, object_id, 0, 0, int32(VirtualMonitorSize.Width), int32(VirtualMonitorSize.Height))

	surface := GetSurfaceFromRole(s, object_id)
	if surface == nil {
		return
	}

	if surface.XdgSurfaceState == nil {
		return
	}

	xdg_surface_state := GetXdgSurfaceObject(s, *surface.XdgSurfaceState)
	if xdg_surface_state == nil {
		return
	}

	go func() {
		xdg_surface_state.configure(s)
		// reposition the popup somehow
	}()

}

func (x *XdgPopup) OnBind(
	cs protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
}

func MakeXdgPopup(
	version uint32,
	parent *protocols.ObjectID[protocols.XdgSurface],
	state XdgPositionerState,
) *protocols.XdgPopup {
	return &protocols.XdgPopup{
		Delegate: &XdgPopup{
			Version: version,
			Parent:  parent,
			State:   state,
		},
	}
}
