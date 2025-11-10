package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type RoleOrXDGSurfaceObjectID interface {
	protocols.ObjectID[protocols.XdgSurface] |
		protocols.ObjectID[protocols.XdgPopup] |
		protocols.ObjectID[protocols.WlSubsurface] |
		protocols.ObjectID[protocols.XdgToplevel] |
		protocols.ObjectID[protocols.XwaylandSurfaceV1]
}

func GetSurfaceFromRole[T RoleOrXDGSurfaceObjectID](cs protocols.ClientState, id T) *WlSurface {
	return cs.GetSurfaceFromRole(protocols.AnyObjectID(id)).(*WlSurface)
}

func GetSurfaceIDFromRole[T RoleOrXDGSurfaceObjectID](cs protocols.ClientState, id T) *protocols.ObjectID[protocols.WlSurface] {
	surfaceID := cs.GetSurfaceIDFromRole(protocols.AnyObjectID(id))
	if surfaceID == nil {
		return nil
	}
	sid := protocols.ObjectID[protocols.WlSurface](*surfaceID)
	return &sid
}

func UnregisterRoleToSurface[T RoleOrXDGSurfaceObjectID](cs protocols.ClientState, id T) {
	cs.UnregisterRoleToSurface(protocols.AnyObjectID(id))
}

func RegisterRoleToSurface[T RoleOrXDGSurfaceObjectID](cs protocols.ClientState, roleID T, surfaceID protocols.ObjectID[protocols.WlSurface]) {
	cs.RegisterRoleToSurface(protocols.AnyObjectID(roleID), surfaceID)
}

func AddObject[T any](cs protocols.ClientState, id protocols.ObjectID[T], v *T) {
	cs.AddObject(protocols.AnyObjectID(id), v)
}

func RemoveObject[T any](cs protocols.ClientState, id protocols.ObjectID[T]) {
	cs.RemoveObject(protocols.AnyObjectID(id))
}

func SendError[T any, U ~uint32 | ~uint8](cs protocols.ClientState, id protocols.ObjectID[T], code U, message string) {
	cs.SendError(protocols.AnyObjectID(id), uint32(code), message)
}

type Size struct {
	Width  uint32
	Height uint32
}

func ToBytes[T ~uint8 | ~uint32](a []T) []byte {
	b := make([]byte, len(a))
	for i, v := range a {
		b[i] = byte(v)
	}
	return b
}

func AreSame[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return (*a) == (*b)
}

func GetWlPoolObject_FromBuffer(cs protocols.ClientState, id protocols.ObjectID[protocols.WlBuffer]) *WlShmPool {
	v := cs.GetObject(protocols.AnyObjectID(id))
	if v == nil {
		return nil
	}
	o := v.(protocols.WaylandObject[protocols.WlBuffer_delegate])
	d := o.GetDelegate()
	return d.(*WlShmPool)
}
