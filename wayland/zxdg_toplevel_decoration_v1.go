package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type ZxdgToplevelDecorationV1 struct {
	XdgToplevel protocols.ObjectID[protocols.XdgToplevel]
}

func (z *ZxdgToplevelDecorationV1) ZxdgToplevelDecorationV1_destroy(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.ZxdgToplevelDecorationV1],
) bool {
	return true
}

func (z *ZxdgToplevelDecorationV1) ZxdgToplevelDecorationV1_set_mode(
	s protocols.ClientState,
	_ protocols.ObjectID[protocols.ZxdgToplevelDecorationV1],
	_mode protocols.ZxdgToplevelDecorationV1Mode_enum,
) {
	surface := GetSurfaceFromRole(s, z.XdgToplevel)
	if surface == nil {
		return
	}
	if surface.XdgSurfaceState == nil {
		return
	}
	xdgSurf := GetXdgSurfaceObject(s, *surface.XdgSurfaceState)
	if xdgSurf == nil {
		return
	}
	xdgSurfaceID := xdgSurf.XdgSurfaceID
	xdgSurf.LatestSerial += 1
	protocols.XdgSurface_configure(s, xdgSurfaceID, xdgSurf.LatestSerial)
}

func (z *ZxdgToplevelDecorationV1) ZxdgToplevelDecorationV1_unset_mode(
	s protocols.ClientState,
	_ protocols.ObjectID[protocols.ZxdgToplevelDecorationV1],
) {
	surface := GetSurfaceFromRole(s, z.XdgToplevel)
	if surface == nil {
		return
	}
	if surface.XdgSurfaceState == nil {
		return
	}
	xdgSurf := GetXdgSurfaceObject(s, *surface.XdgSurfaceState)
	if xdgSurf == nil {
		return
	}
	xdgSurfaceID := xdgSurf.XdgSurfaceID
	xdgSurf.LatestSerial += 1
	protocols.XdgSurface_configure(s, xdgSurfaceID, xdgSurf.LatestSerial)
}

func (z *ZxdgToplevelDecorationV1) OnBind(
	_ protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	_ protocols.AnyObjectID,
	_ uint32,
) {
	// no-op
}

func MakeZxdgToplevelDecorationV1(xdgToplevel protocols.ObjectID[protocols.XdgToplevel]) *protocols.ZxdgToplevelDecorationV1 {
	return &protocols.ZxdgToplevelDecorationV1{
		Delegate: &ZxdgToplevelDecorationV1{
			XdgToplevel: xdgToplevel,
		},
	}
}
