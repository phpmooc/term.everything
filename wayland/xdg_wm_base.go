package wayland

import (
	"fmt"

	"github.com/mmulet/term.everything/wayland/protocols"
)

type XdgWmBase struct {
	Version uint32
}

func (x *XdgWmBase) XdgWmBase_destroy(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgWmBase],
) bool {
	return true
}

func (x *XdgWmBase) XdgWmBase_create_positioner(
	s protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgWmBase],
	id protocols.ObjectID[protocols.XdgPositioner],
) {
	AddObject(s, id, MakeXdgPositioner())
}

func (x *XdgWmBase) XdgWmBase_get_xdg_surface(
	s protocols.ClientState,
	objectID protocols.ObjectID[protocols.XdgWmBase],
	xdgSurfaceID protocols.ObjectID[protocols.XdgSurface],
	surfaceID protocols.ObjectID[protocols.WlSurface],
) {
	surface := GetWlSurfaceObject(s, surfaceID)
	if surface == nil {
		SendError(s, objectID, protocols.XdgWmBaseError_enum_role, "surface not found")
		return
	}

	if surface.Role != nil {
		/**
		 * It is illegal to create an xdg_surface for a wl_surface which already has
		 * a assigned role
		 */
		fmt.Printf("XdgWmBase_get_xdg_surface: surface@%d already has a role\n", surfaceID)
		SendError(s,
			objectID,
			protocols.XdgWmBaseError_enum_role,
			"surface already has a role",
		)
		return
	}

	surface.XdgSurfaceState = &xdgSurfaceID

	// surface.xdg_surface_state = {
	//   on_configure: new Map(),
	//   latest_configure_serial: 0,
	//   xdg_surface_id: xdg_surface_id,
	//   window_geometry: { x: 0, y: 0, width: 0, height: 0 },
	// };

	RegisterRoleToSurface(s, xdgSurfaceID, surfaceID)

	AddObject(s, xdgSurfaceID, MakeXdgSurface(x.Version, xdgSurfaceID))
}

func (x *XdgWmBase) XdgWmBase_pong(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgWmBase],
	_ uint32,
) {
	// no-op
}

func (x *XdgWmBase) OnBind(
	_ protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	_ protocols.AnyObjectID,
	version uint32,
) {
	x.Version = version
}

func MakeXdgWmBase() *protocols.XdgWmBase {
	return &protocols.XdgWmBase{
		Delegate: &XdgWmBase{
			Version: 1,
		},
	}
}
