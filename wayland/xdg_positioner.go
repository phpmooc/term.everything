package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type anchorRect struct {
	X      int32
	Y      int32
	Width  int32
	Height int32
}

type offset struct {
	X int32
	Y int32
}

type XdgPositionerState struct {
	Width                 int32
	Height                int32
	AnchorRect            anchorRect
	Anchor                protocols.XdgPositionerAnchor_enum
	Gravity               protocols.XdgPositionerGravity_enum
	ConstraintAdjustment  protocols.XdgPositionerConstraintAdjustment_enum
	Offset                offset
	Reactive              bool
	ParentSize            Size
	ParentConfigureSerial uint32
}

type XdgPositioner struct {
	state XdgPositionerState
}

func (x *XdgPositioner) XdgPositioner_destroy(
	s protocols.ClientState,
	id protocols.ObjectID[protocols.XdgPositioner],
) bool {
	return true
}

func (x *XdgPositioner) XdgPositioner_set_size(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgPositioner],
	width int32,
	height int32,
) {
	x.state.Width = width
	x.state.Height = height
}

func (x *XdgPositioner) XdgPositioner_set_anchor_rect(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgPositioner],
	ax int32,
	ay int32,
	aw int32,
	ah int32,
) {
	x.state.AnchorRect = anchorRect{X: ax, Y: ay, Width: aw, Height: ah}
}

func (x *XdgPositioner) XdgPositioner_set_anchor(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgPositioner],
	anchor protocols.XdgPositionerAnchor_enum,
) {
	x.state.Anchor = anchor
}

func (x *XdgPositioner) XdgPositioner_set_gravity(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgPositioner],
	gravity protocols.XdgPositionerGravity_enum,
) {
	x.state.Gravity = gravity
}

func (x *XdgPositioner) XdgPositioner_set_constraint_adjustment(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgPositioner],
	adj protocols.XdgPositionerConstraintAdjustment_enum,
) {
	/** @TODO: Implement xdg_positioner_set_constraint_adjustment */
}

func (x *XdgPositioner) XdgPositioner_set_offset(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgPositioner],
	ox int32,
	oy int32,
) {
	x.state.Offset = offset{X: ox, Y: oy}
}

func (x *XdgPositioner) XdgPositioner_set_reactive(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgPositioner],
) {
	x.state.Reactive = true
}

func (x *XdgPositioner) XdgPositioner_set_parent_size(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgPositioner],
	parentWidth int32,
	parentHeight int32,
) {
	var pw, ph uint32
	if parentWidth > 0 {
		pw = uint32(parentWidth)
	}
	if parentHeight > 0 {
		ph = uint32(parentHeight)
	}
	x.state.ParentSize = Size{Width: pw, Height: ph}
}

func (x *XdgPositioner) XdgPositioner_set_parent_configure(
	_ protocols.ClientState,
	_ protocols.ObjectID[protocols.XdgPositioner],
	serial uint32,
) {
	x.state.ParentConfigureSerial = serial
}

func (x *XdgPositioner) OnBind(
	cs protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
}

func MakeXdgPositioner() *protocols.XdgPositioner {
	return &protocols.XdgPositioner{
		Delegate: &XdgPositioner{},
	}
}
