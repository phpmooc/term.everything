package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type XdgToplevel struct {
	Parent *protocols.ObjectID[protocols.XdgToplevel]

	Title *string
	AppID string

	Maximized  bool
	Fullscreen bool

	MinSize *Size
	MaxSize *Size

	pendingMinSize *Size
	pendingMaxSize *Size
}

func (t *XdgToplevel) HasPendingState() bool {
	return t.pendingMinSize != nil || t.pendingMaxSize != nil
}

func (t *XdgToplevel) XdgToplevel_destroy(
	s protocols.ClientState,
	objectID protocols.ObjectID[protocols.XdgToplevel],
) bool {
	surface := GetSurfaceFromRole(s, objectID)

	UnregisterRoleToSurface(s, objectID)
	s.TopLevelSurfaces()[objectID] = false
	if surface != nil {
		surface.ClearRoleData()
	}
	return true

}

func (t *XdgToplevel) XdgToplevel_set_parent(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgToplevel],
	parent *protocols.ObjectID[protocols.XdgToplevel],
) {
	t.Parent = parent
}

func (t *XdgToplevel) XdgToplevel_set_title(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgToplevel],
	title string,
) {
	t.Title = &title
}

func (t *XdgToplevel) XdgToplevel_set_app_id(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgToplevel],
	appID string,
) {
	t.AppID = appID
}

func (t *XdgToplevel) XdgToplevel_show_window_menu(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgToplevel],
	_ protocols.ObjectID[protocols.WlSeat],
	_ uint32, // serial
	_ int32, // x
	_ int32, // y
) {
	// TODO: Implement show_window_menu
}

func (t *XdgToplevel) XdgToplevel_move(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgToplevel],
	_ protocols.ObjectID[protocols.WlSeat],
	_ uint32, // serial
) {
	// TODO: Implement move
}

func (t *XdgToplevel) XdgToplevel_resize(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgToplevel],
	_ protocols.ObjectID[protocols.WlSeat],
	_ uint32, // serial
	_ protocols.XdgToplevelResizeEdge_enum, // edges
) {
	// TODO: Implement resize
}

func (t *XdgToplevel) XdgToplevel_set_max_size(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgToplevel],
	width int32,
	height int32,
) {
	if width <= 0 || height <= 0 {
		t.pendingMaxSize = nil
		return
	}
	t.pendingMaxSize = &Size{
		Width:  uint32(width),
		Height: uint32(height),
	}
}

func (t *XdgToplevel) XdgToplevel_set_min_size(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgToplevel],
	width int32,
	height int32,
) {
	if width <= 0 || height <= 0 {
		t.pendingMinSize = nil
		return
	}
	t.pendingMinSize = &Size{
		Width:  uint32(width),
		Height: uint32(height),
	}
}

func (t *XdgToplevel) XdgToplevel_set_maximized(
	s protocols.ClientState,
	objectID protocols.ObjectID[protocols.XdgToplevel],
) {
	go func() {
		if shouldChange := t.stateConfiguration(s, objectID, true, t.Fullscreen); shouldChange {
			t.Maximized = true
		}
	}()

}

func (t *XdgToplevel) XdgToplevel_unset_maximized(
	s protocols.ClientState,
	objectID protocols.ObjectID[protocols.XdgToplevel],
) {
	go func() {
		if shouldChange := t.stateConfiguration(s, objectID, false, t.Fullscreen); shouldChange {
			t.Maximized = false
		}
	}()
}

func (t *XdgToplevel) XdgToplevel_set_fullscreen(
	s protocols.ClientState,
	objectID protocols.ObjectID[protocols.XdgToplevel],
	_ *protocols.ObjectID[protocols.WlOutput],
) {
	go func() {
		if shouldChange := t.stateConfiguration(s, objectID, t.Maximized, true); shouldChange {
			t.Fullscreen = true
		}
	}()
}

func (t *XdgToplevel) XdgToplevel_unset_fullscreen(
	s protocols.ClientState,
	objectID protocols.ObjectID[protocols.XdgToplevel],
) {
	go func() {
		if shouldChange := t.stateConfiguration(s, objectID, t.Maximized, false); shouldChange {
			t.Fullscreen = false
		}
	}()
}

func (t *XdgToplevel) XdgToplevel_set_minimized(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgToplevel],
) {
	// TODO: Implement minimize behavior if desired
}

func (t *XdgToplevel) OnBind(
	cs protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
	// No-op
}

func (t *XdgToplevel) stateConfiguration(
	s protocols.ClientState,
	objectID protocols.ObjectID[protocols.XdgToplevel],
	maximized bool,
	fullscreen bool,
) bool {
	// const data = s.get_role_data_from_role(object_id, "xdg_toplevel");
	// if (!data) {
	//   return false;
	// }

	// Ensure the role resolves to a wl_surface
	surface := GetSurfaceFromRole(s, objectID)
	if surface == nil {
		return false
	}
	if surface.XdgSurfaceState == nil {
		return false
	}

	xdg_surface_State := GetXdgSurfaceObject(s, *surface.XdgSurfaceState)

	if xdg_surface_State == nil {
		return false
	}

	var states []protocols.XdgToplevelState_enum
	if maximized {
		states = append(states, protocols.XdgToplevelState_enum_maximized)
	}
	if fullscreen {
		states = append(states, protocols.XdgToplevelState_enum_fullscreen)
	}

	protocols.XdgToplevel_configure(
		s,
		objectID,
		int32(VirtualMonitorSize.Width),
		int32(VirtualMonitorSize.Height),
		ToBytes(states),
	)
	xdg_surface_State.configure(s)

	return true
}

func MakeXdgToplevel() *protocols.XdgToplevel {
	return &protocols.XdgToplevel{
		Delegate: &XdgToplevel{},
	}
}
