package wayland

import (
	"slices"

	"github.com/mmulet/term.everything/wayland/protocols"
)

type WlSubsurface struct {
	Parent   protocols.ObjectID[protocols.WlSurface]
	Sync     bool
	Position Point
}

/**
 * wl_subsurface methods
 */

func (ss *WlSubsurface) WlSubsurface_destroy(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlSubsurface],
) bool {
	/**
	 * Delete the association between the role object and the surface
	 */
	surfaceID := GetSurfaceIDFromRole(s, object_id)

	UnregisterRoleToSurface(s, object_id)

	if surfaceID == nil {
		return true
	}

	surface := GetWlSurfaceObject(s, *surfaceID)
	if surface == nil {
		return true
	}

	if !surface.HasRoleData() {
		return true
	}

	_, isSubSurface := surface.Role.(*SurfaceRoleSubSurface)
	if !isSubSurface {
		return true
	}

	parent_surface := GetWlSurfaceObject(s, ss.Parent)
	if parent_surface != nil {
		parent_surface.ChildrenInDrawOrder = slices.DeleteFunc(parent_surface.ChildrenInDrawOrder, func(id *protocols.ObjectID[protocols.WlSurface]) bool {
			return id == surfaceID
		})
	}
	surface.ClearRoleData()
	return true
}

func (ss *WlSubsurface) WlSubsurface_set_position(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlSubsurface],
	x int32,
	y int32,
) {
	surfaceID := GetSurfaceIDFromRole(s, object_id)
	if surfaceID == nil {
		SendError(s, object_id, protocols.WlSubsurfaceError_enum_bad_surface, "surface not found")
		return
	}

	parent := GetWlSurfaceObject(s, ss.Parent)
	if parent == nil {
		SendError(s, object_id, protocols.WlSubsurfaceError_enum_bad_surface, "parent not found")
		return
	}
	if (*parent).PendingUpdate.SetChildPosition == nil {
		(*parent).PendingUpdate.SetChildPosition = []ChildPosition{}
	}

	(*parent).PendingUpdate.SetChildPosition = append(
		(*parent).PendingUpdate.SetChildPosition,
		ChildPosition{
			Child: *surfaceID,
			X:     x,
			Y:     y,
		},
	)
}

func (ss *WlSubsurface) WlSubsurface_place_above(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlSubsurface],
	sibling_or_parent_id protocols.ObjectID[protocols.WlSurface],
) {
	ss.placeSubsurface(s, object_id, sibling_or_parent_id, ZOrderTypeAbove)
}

func (ss *WlSubsurface) placeSubsurface(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlSubsurface],
	sibling_or_parent_id protocols.ObjectID[protocols.WlSurface],
	aboveOrBelow ZOrder,
) {
	surfaceID := GetSurfaceIDFromRole(s, object_id)
	if surfaceID == nil {
		SendError(s, object_id, protocols.WlSubsurfaceError_enum_bad_surface, "surface not found")
		return
	}

	parent := GetWlSurfaceObject(s, ss.Parent)

	if parent == nil {
		SendError(s, object_id, protocols.WlSubsurfaceError_enum_bad_surface, "parent not found")
		return
	}

	var id *protocols.ObjectID[protocols.WlSurface]
	if sibling_or_parent_id == ss.Parent {
		id = nil
	} else {
		id = &sibling_or_parent_id
	}

	if (*parent).PendingUpdate.ZOrderSubsurfaces == nil {
		(*parent).PendingUpdate.ZOrderSubsurfaces = []ZOrderSubsurface{}
	}

	(*parent).PendingUpdate.ZOrderSubsurfaces = append(
		(*parent).PendingUpdate.ZOrderSubsurfaces,
		ZOrderSubsurface{
			Type:        aboveOrBelow,
			ChildToMove: *surfaceID,
			RelativeTo:  id,
		},
	)
}
func (ss *WlSubsurface) WlSubsurface_place_below(
	s protocols.ClientState,
	object_id protocols.ObjectID[protocols.WlSubsurface],
	sibling_or_parent_id protocols.ObjectID[protocols.WlSurface],
) {
	ss.placeSubsurface(s, object_id, sibling_or_parent_id, ZOrderTypeBelow)
}

func (ss *WlSubsurface) WlSubsurface_set_sync(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlSubsurface],
) {
	ss.Sync = true
}

func (ss *WlSubsurface) WlSubsurface_set_desync(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.WlSubsurface],
) {
	ss.Sync = false
}

func (ss *WlSubsurface) OnBind(
	cs protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
	// No-op
}

// Constructor (literal to TS `static make(parent)` style)
func MakeWlSubsurface(parent protocols.ObjectID[protocols.WlSurface]) *protocols.WlSubsurface {
	return &protocols.WlSubsurface{
		Delegate: &WlSubsurface{
			Parent:   parent,
			Sync:     true,
			Position: Point{X: 0, Y: 0},
		},
	}
}
