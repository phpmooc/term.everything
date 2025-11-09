package wayland

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"syscall"
	"time"

	"github.com/mmulet/term.everything/wayland/protocols"
)

type WaylandClient struct {
	drawableSurfaces map[protocols.ObjectID[protocols.WlSurface]]bool
	topLevelSurfaces map[protocols.ObjectID[protocols.XdgToplevel]]bool

	// transport
	conn *net.UnixConn

	// protocol version bits
	compositorVersion uint32

	// id of the wl_display object for this client (usually 1)
	displayID protocols.ObjectID[protocols.WlDisplay]

	// recv buffer
	recvBuf []byte

	// outgoing event queue (channel-based)
	outgoing chan protocols.OutgoingEvent

	// decoded message framing
	decoder messageDecoder

	// FD queue claimed by requests that accept an fd (single-threaded access)
	unclaimedFDs []protocols.FileDescriptor

	// object store
	objects map[protocols.AnyObjectID]any

	// role -> surface mapping
	rolesToSurfaces map[protocols.AnyObjectID]protocols.ObjectID[protocols.WlSurface]

	// frame callbacks queued by wl_surface.frame
	frameDrawRequests []protocols.ObjectID[protocols.WlCallback]

	// global binds: one holder keyed by GlobalID; the value is a pointer to the
	// correctly-typed map[*ObjectID[T]]Version, created on demand.
	globalBinds map[protocols.GlobalID]any
}

func (c *WaylandClient) AddFrameDrawRequest(cb protocols.ObjectID[protocols.WlCallback]) {
	c.frameDrawRequests = append(c.frameDrawRequests, cb)
}

func (c *WaylandClient) GetSurfaceIDFromRole(roleObjectID protocols.AnyObjectID) *protocols.ObjectID[protocols.WlSurface] {
	if sid, ok := c.rolesToSurfaces[roleObjectID]; ok {
		return &sid
	}
	return nil
}

func (c *WaylandClient) GetSurfaceFromRole(roleObjectID protocols.AnyObjectID) any {
	sidAny := c.GetSurfaceIDFromRole(roleObjectID)
	if sidAny == nil {
		return nil
	}
	surface := GetWlSurfaceObject(c, *sidAny)
	return surface
}

func (c *WaylandClient) UnregisterRoleToSurface(roleID protocols.AnyObjectID) {
	delete(c.rolesToSurfaces, roleID)
}

func (c *WaylandClient) RegisterRoleToSurface(roleID protocols.AnyObjectID, surfaceID protocols.ObjectID[protocols.WlSurface]) {
	c.rolesToSurfaces[roleID] = surfaceID
}

/**
 * Seed if maybe_desceneding_id is a descendant of surface_id
 * @param s
 * @param surface_id
 * @param maybe_descendant_id
 */
func (c *WaylandClient) FindDescendantSurface(surfaceID protocols.ObjectID[protocols.WlSurface], maybeDescendantID protocols.ObjectID[protocols.WlSurface]) bool {

	surface := GetWlSurfaceObject(c, surfaceID)
	if surface == nil {
		return false
	}

	for _, childID := range surface.ChildrenInDrawOrder {
		if childID == nil {
			continue
		}
		if *childID == maybeDescendantID {
			return true
		}
	}

	for _, childID := range surface.ChildrenInDrawOrder {
		if childID == nil {
			continue
		}
		if c.FindDescendantSurface(*childID, maybeDescendantID) {
			return true
		}
	}

	return false
}

//TODO continue looking through wayland client.t

func NewWaylandClient(conn *net.UnixConn) *WaylandClient {
	return &WaylandClient{
		conn:              conn,
		compositorVersion: 1,
		displayID:         protocols.ObjectID[protocols.WlDisplay](1),

		recvBuf: make([]byte, 64*1024),

		outgoing: make(chan protocols.OutgoingEvent, 256),

		unclaimedFDs:    make([]protocols.FileDescriptor, 0, 8),
		objects:         make(map[protocols.AnyObjectID]any),
		rolesToSurfaces: make(map[protocols.AnyObjectID]protocols.ObjectID[protocols.WlSurface]),

		drawableSurfaces: make(map[protocols.ObjectID[protocols.WlSurface]]bool),
		topLevelSurfaces: make(map[protocols.ObjectID[protocols.XdgToplevel]]bool),

		globalBinds: make(map[protocols.GlobalID]any),
	}
}

func (c *WaylandClient) MainLoop() {
	for {
		// Drain all pending outgoing events first (non-blocking drain).
		for {
			select {
			case ev := <-c.outgoing:
				if !c.sendOne(ev) {
					return
				}
			default:
				goto drained
			}
		}
	drained:

		// Receive once with short deadline; parse and dispatch.
		n, fds, shouldContinue, err := GetMessageAndFileDescriptors(c.conn, c.recvBuf)
		if err != nil {
			// treat unexpected read errors as fatal
			log.Printf("Recv error: %v", err)
			return
		}
		if !shouldContinue {
			// orderly shutdown or fatal error
			return
		}

		// Add any received FDs to the queue (single-threaded context).
		if len(fds) > 0 && WaylandDebugTimeOnly() {
			log.Printf("client: received %d file descriptors", len(fds))
		}
		for _, fd := range fds {
			c.unclaimedFDs = append(c.unclaimedFDs, protocols.FileDescriptor(fd))
		}

		// If timeout / nothing read, avoid tight loop.
		if n == 0 {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		// Decode into frames and dispatch.
		msgs := c.decoder.Consume(c.recvBuf[:n])
		for i := range msgs {
			m := msgs[i]
			obj := c.GetObject(m.ObjectID)
			if obj == nil {
				if WaylandDebugTimeOnly() {
					log.Printf("client: request for unknown object %d", uint32(m.ObjectID))
				}
				continue
			}
			switch oo := obj.(type) {
			case protocols.WaylandObject[protocols.WlDisplay_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlRegistry_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlCompositor_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlShm_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlShmPool_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlBuffer_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlSurface_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlRegion_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlSeat_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlPointer_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlKeyboard_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlTouch_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlDataDeviceManager_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlDataDevice_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlDataSource_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlDataOffer_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlShell_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlShellSurface_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlSubcompositor_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.WlSubsurface_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.XdgWmBase_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.XdgSurface_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.XdgToplevel_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.XdgPositioner_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.XdgPopup_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.ZxdgToplevelDecorationV1_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.ZxdgDecorationManagerV1_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.ZwpXwaylandKeyboardGrabManagerV1_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.XwaylandShellV1_delegate]:
				oo.OnRequest(c, m)
			case protocols.WaylandObject[protocols.XwaylandSurfaceV1_delegate]:
				oo.OnRequest(c, m)
			default:
				log.Printf("client: object %d has unknown type; cannot dispatch", uint32(m.ObjectID))
			}
		}
	}
}

// Sender: enqueue event via channel
func (c *WaylandClient) Send(ev protocols.OutgoingEvent) {
	// Allow backpressure to naturally block the sender goroutine.
	c.outgoing <- ev
}

// encode and send one event (with optional FD)
func (c *WaylandClient) sendOne(ev protocols.OutgoingEvent) bool {
	if WaylandDebugTimeOnly() {
		log.Printf("client -> eid=%d opcode=%d len=%d fd=%v",
			uint32(ev.ObjectID), ev.Opcode, len(ev.Data), ev.FileDescriptor)
	}
	size := 8 + len(ev.Data)
	buf := make([]byte, size)

	// Wayland header: object_id (u32), size (u16), opcode (u16)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(ev.ObjectID))
	binary.LittleEndian.PutUint16(buf[4:6], uint16(size))
	binary.LittleEndian.PutUint16(buf[6:8], uint16(ev.Opcode))
	copy(buf[8:], ev.Data)

	var fds []int
	if ev.FileDescriptor != nil {
		fds = []int{int(*ev.FileDescriptor)}
	}

	n, ok, err := SendMessageAndFileDescriptors(c.conn, buf, fds)
	if err != nil {
		// if EPIPE or similar => disconnect
		if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
			return false
		}
		log.Printf("Send error: %v (n=%d ok=%v)", err, n, ok)
		return false
	}
	return true
}

// recvOnce is folded into MainLoop via GetMessageAndFileDescriptors

// messageDecoder turns a byte stream into DecodeState frames
type messageDecoder struct {
	buf []byte
}

func (d *messageDecoder) Consume(in []byte) []protocols.DecodeState {
	// append
	d.buf = append(d.buf, in...)

	var out []protocols.DecodeState
	for {
		if len(d.buf) < 8 {
			break
		}
		objectID := binary.LittleEndian.Uint32(d.buf[0:4])
		size := binary.LittleEndian.Uint16(d.buf[4:6])
		opcode := binary.LittleEndian.Uint16(d.buf[6:8])

		if int(size) < 8 {
			// invalid; drop
			d.buf = nil
			break
		}
		if len(d.buf) < int(size) {
			// incomplete
			break
		}
		payload := make([]byte, int(size)-8)
		copy(payload, d.buf[8:size])

		out = append(out, protocols.DecodeState{
			ObjectID: protocols.AnyObjectID(objectID),
			Opcode:   opcode,
			Size:     size,
			Data:     payload,
		})
		d.buf = d.buf[size:]
	}
	return out
}

// -------- ClientState implementation (single-threaded via MainLoop) --------

func (c *WaylandClient) RemoveObject(id protocols.AnyObjectID) {
	delete(c.objects, id)
	// Optional: wl_display.delete_id could be emitted here if desired.
}

func (c *WaylandClient) RemoveGlobalBind(globalID protocols.GlobalID, id protocols.AnyObjectID) {
	switch globalID {
	case protocols.GlobalID_WlDisplay:
		if m := c.ensureGlobalMap_WlDisplay(); m != nil {
			delete(*m, protocols.ObjectID[protocols.WlDisplay](id))
		}
	case protocols.GlobalID_WlCompositor:
		if m := c.ensureGlobalMap_WlCompositor(); m != nil {
			delete(*m, protocols.ObjectID[protocols.WlCompositor](id))
		}
	case protocols.GlobalID_WlSubcompositor:
		if m := c.ensureGlobalMap_WlSubcompositor(); m != nil {
			delete(*m, protocols.ObjectID[protocols.WlSubcompositor](id))
		}
	case protocols.GlobalID_WlOutput:
		if m := c.ensureGlobalMap_WlOutput(); m != nil {
			delete(*m, protocols.ObjectID[protocols.WlOutput](id))
		}
	case protocols.GlobalID_WlSeat:
		if m := c.ensureGlobalMap_WlSeat(); m != nil {
			delete(*m, protocols.ObjectID[protocols.WlSeat](id))
		}
	case protocols.GlobalID_WlShm:
		if m := c.ensureGlobalMap_WlShm(); m != nil {
			delete(*m, protocols.ObjectID[protocols.WlShm](id))
		}
	case protocols.GlobalID_XdgWmBase:
		if m := c.ensureGlobalMap_XdgWmBase(); m != nil {
			delete(*m, protocols.ObjectID[protocols.XdgWmBase](id))
		}
	case protocols.GlobalID_WlDataDeviceManager:
		if m := c.ensureGlobalMap_WlDataDeviceManager(); m != nil {
			delete(*m, protocols.ObjectID[protocols.WlDataDeviceManager](id))
		}
	case protocols.GlobalID_WlKeyboard:
		if m := c.ensureGlobalMap_WlKeyboard(); m != nil {
			delete(*m, protocols.ObjectID[protocols.WlKeyboard](id))
		}
	case protocols.GlobalID_WlPointer:
		if m := c.ensureGlobalMap_WlPointer(); m != nil {
			delete(*m, protocols.ObjectID[protocols.WlPointer](id))
		}
	case protocols.GlobalID_ZwpXwaylandKeyboardGrabManagerV1:
		if m := c.ensureGlobalMap_ZwpXwaylandKeyboardGrabManagerV1(); m != nil {
			delete(*m, protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1](id))
		}
	case protocols.GlobalID_XwaylandShellV1:
		if m := c.ensureGlobalMap_XwaylandShellV1(); m != nil {
			delete(*m, protocols.ObjectID[protocols.XwaylandShellV1](id))
		}
	case protocols.GlobalID_WlDataDevice:
		if m := c.ensureGlobalMap_WlDataDevice(); m != nil {
			delete(*m, protocols.ObjectID[protocols.WlDataDevice](id))
		}
	case protocols.GlobalID_WlTouch:
		if m := c.ensureGlobalMap_WlTouch(); m != nil {
			delete(*m, protocols.ObjectID[protocols.WlTouch](id))
		}
	case protocols.GlobalID_ZxdgDecorationManagerV1:
		if m := c.ensureGlobalMap_ZxdgDecorationManagerV1(); m != nil {
			delete(*m, protocols.ObjectID[protocols.ZxdgDecorationManagerV1](id))
		}
	default:
	}
}

func (c *WaylandClient) AddObject(id protocols.AnyObjectID, v any) {
	if v == nil {
		log.Printf("AddObject: object is nil for id %d", uint32(id))
		return
	}
	c.objects[id] = v
}

func (c *WaylandClient) SetCompositorVersion(v uint32) { c.compositorVersion = v }
func (c *WaylandClient) GetCompositorVersion() uint32  { return c.compositorVersion }

func (c *WaylandClient) GetObject(id protocols.AnyObjectID) any {
	return c.objects[id]
}

func (c *WaylandClient) SendError(objectID protocols.AnyObjectID, code uint32, message string) {
	protocols.WlDisplay_error(c, c.displayID, objectID, code, message)
}

func (c *WaylandClient) DrawableSurfaces() map[protocols.ObjectID[protocols.WlSurface]]bool {
	return c.drawableSurfaces
}
func (c *WaylandClient) TopLevelSurfaces() map[protocols.ObjectID[protocols.XdgToplevel]]bool {
	return c.topLevelSurfaces
}

// Typed getter used by generated helpers (expects a pointer to a typed map)
func (c *WaylandClient) GetGlobalBinds(globalID protocols.GlobalID) any {
	switch globalID {
	case protocols.GlobalID_WlDisplay:
		return c.ensureGlobalMap_WlDisplay()
	case protocols.GlobalID_WlCompositor:
		return c.ensureGlobalMap_WlCompositor()
	case protocols.GlobalID_WlSubcompositor:
		return c.ensureGlobalMap_WlSubcompositor()
	case protocols.GlobalID_WlOutput:
		return c.ensureGlobalMap_WlOutput()
	case protocols.GlobalID_WlSeat:
		return c.ensureGlobalMap_WlSeat()
	case protocols.GlobalID_WlShm:
		return c.ensureGlobalMap_WlShm()
	case protocols.GlobalID_XdgWmBase:
		return c.ensureGlobalMap_XdgWmBase()
	case protocols.GlobalID_WlDataDeviceManager:
		return c.ensureGlobalMap_WlDataDeviceManager()
	case protocols.GlobalID_WlKeyboard:
		return c.ensureGlobalMap_WlKeyboard()
	case protocols.GlobalID_WlPointer:
		return c.ensureGlobalMap_WlPointer()
	case protocols.GlobalID_ZwpXwaylandKeyboardGrabManagerV1:
		return c.ensureGlobalMap_ZwpXwaylandKeyboardGrabManagerV1()
	case protocols.GlobalID_XwaylandShellV1:
		return c.ensureGlobalMap_XwaylandShellV1()
	case protocols.GlobalID_WlDataDevice:
		return c.ensureGlobalMap_WlDataDevice()
	case protocols.GlobalID_WlTouch:
		return c.ensureGlobalMap_WlTouch()
	case protocols.GlobalID_ZxdgDecorationManagerV1:
		return c.ensureGlobalMap_ZxdgDecorationManagerV1()
	default:
		return nil
	}
}

func (c *WaylandClient) AddGlobalBind(globalID protocols.GlobalID, objectID protocols.AnyObjectID, version protocols.Version) {
	switch globalID {
	case protocols.GlobalID_WlDisplay:
		m := c.ensureGlobalMap_WlDisplay()
		(*m)[protocols.ObjectID[protocols.WlDisplay](objectID)] = version
	case protocols.GlobalID_WlCompositor:
		m := c.ensureGlobalMap_WlCompositor()
		(*m)[protocols.ObjectID[protocols.WlCompositor](objectID)] = version
	case protocols.GlobalID_WlSubcompositor:
		m := c.ensureGlobalMap_WlSubcompositor()
		(*m)[protocols.ObjectID[protocols.WlSubcompositor](objectID)] = version
	case protocols.GlobalID_WlOutput:
		m := c.ensureGlobalMap_WlOutput()
		(*m)[protocols.ObjectID[protocols.WlOutput](objectID)] = version
	case protocols.GlobalID_WlSeat:
		m := c.ensureGlobalMap_WlSeat()
		(*m)[protocols.ObjectID[protocols.WlSeat](objectID)] = version
	case protocols.GlobalID_WlShm:
		m := c.ensureGlobalMap_WlShm()
		(*m)[protocols.ObjectID[protocols.WlShm](objectID)] = version
	case protocols.GlobalID_XdgWmBase:
		m := c.ensureGlobalMap_XdgWmBase()
		(*m)[protocols.ObjectID[protocols.XdgWmBase](objectID)] = version
	case protocols.GlobalID_WlDataDeviceManager:
		m := c.ensureGlobalMap_WlDataDeviceManager()
		(*m)[protocols.ObjectID[protocols.WlDataDeviceManager](objectID)] = version
	case protocols.GlobalID_WlKeyboard:
		m := c.ensureGlobalMap_WlKeyboard()
		(*m)[protocols.ObjectID[protocols.WlKeyboard](objectID)] = version
	case protocols.GlobalID_WlPointer:
		m := c.ensureGlobalMap_WlPointer()
		(*m)[protocols.ObjectID[protocols.WlPointer](objectID)] = version
	case protocols.GlobalID_ZwpXwaylandKeyboardGrabManagerV1:
		m := c.ensureGlobalMap_ZwpXwaylandKeyboardGrabManagerV1()
		(*m)[protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1](objectID)] = version
	case protocols.GlobalID_XwaylandShellV1:
		m := c.ensureGlobalMap_XwaylandShellV1()
		(*m)[protocols.ObjectID[protocols.XwaylandShellV1](objectID)] = version
	case protocols.GlobalID_WlDataDevice:
		m := c.ensureGlobalMap_WlDataDevice()
		(*m)[protocols.ObjectID[protocols.WlDataDevice](objectID)] = version
	case protocols.GlobalID_WlTouch:
		m := c.ensureGlobalMap_WlTouch()
		(*m)[protocols.ObjectID[protocols.WlTouch](objectID)] = version
	case protocols.GlobalID_ZxdgDecorationManagerV1:
		m := c.ensureGlobalMap_ZxdgDecorationManagerV1()
		(*m)[protocols.ObjectID[protocols.ZxdgDecorationManagerV1](objectID)] = version
	default:
	}
}

// FileDescriptorClaimClientState
func (c *WaylandClient) ClaimFileDescriptor() *protocols.FileDescriptor {
	if len(c.unclaimedFDs) == 0 {
		return nil
	}
	fd := c.unclaimedFDs[0]
	copy(c.unclaimedFDs, c.unclaimedFDs[1:])
	c.unclaimedFDs = c.unclaimedFDs[:len(c.unclaimedFDs)-1]
	return &fd
}

// -------- Global binds typed-map creators (on demand) --------

func (c *WaylandClient) ensureGlobalMap_WlDisplay() *map[protocols.ObjectID[protocols.WlDisplay]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_WlDisplay]; ok {
		return v.(*map[protocols.ObjectID[protocols.WlDisplay]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.WlDisplay]]protocols.Version)
	c.globalBinds[protocols.GlobalID_WlDisplay] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_WlCompositor() *map[protocols.ObjectID[protocols.WlCompositor]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_WlCompositor]; ok {
		return v.(*map[protocols.ObjectID[protocols.WlCompositor]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.WlCompositor]]protocols.Version)
	c.globalBinds[protocols.GlobalID_WlCompositor] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_WlSubcompositor() *map[protocols.ObjectID[protocols.WlSubcompositor]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_WlSubcompositor]; ok {
		return v.(*map[protocols.ObjectID[protocols.WlSubcompositor]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.WlSubcompositor]]protocols.Version)
	c.globalBinds[protocols.GlobalID_WlSubcompositor] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_WlOutput() *map[protocols.ObjectID[protocols.WlOutput]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_WlOutput]; ok {
		return v.(*map[protocols.ObjectID[protocols.WlOutput]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.WlOutput]]protocols.Version)
	c.globalBinds[protocols.GlobalID_WlOutput] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_WlSeat() *map[protocols.ObjectID[protocols.WlSeat]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_WlSeat]; ok {
		return v.(*map[protocols.ObjectID[protocols.WlSeat]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.WlSeat]]protocols.Version)
	c.globalBinds[protocols.GlobalID_WlSeat] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_WlShm() *map[protocols.ObjectID[protocols.WlShm]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_WlShm]; ok {
		return v.(*map[protocols.ObjectID[protocols.WlShm]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.WlShm]]protocols.Version)
	c.globalBinds[protocols.GlobalID_WlShm] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_XdgWmBase() *map[protocols.ObjectID[protocols.XdgWmBase]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_XdgWmBase]; ok {
		return v.(*map[protocols.ObjectID[protocols.XdgWmBase]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.XdgWmBase]]protocols.Version)
	c.globalBinds[protocols.GlobalID_XdgWmBase] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_WlDataDeviceManager() *map[protocols.ObjectID[protocols.WlDataDeviceManager]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_WlDataDeviceManager]; ok {
		return v.(*map[protocols.ObjectID[protocols.WlDataDeviceManager]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.WlDataDeviceManager]]protocols.Version)
	c.globalBinds[protocols.GlobalID_WlDataDeviceManager] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_WlKeyboard() *map[protocols.ObjectID[protocols.WlKeyboard]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_WlKeyboard]; ok {
		return v.(*map[protocols.ObjectID[protocols.WlKeyboard]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.WlKeyboard]]protocols.Version)
	c.globalBinds[protocols.GlobalID_WlKeyboard] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_WlPointer() *map[protocols.ObjectID[protocols.WlPointer]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_WlPointer]; ok {
		return v.(*map[protocols.ObjectID[protocols.WlPointer]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.WlPointer]]protocols.Version)
	c.globalBinds[protocols.GlobalID_WlPointer] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_ZwpXwaylandKeyboardGrabManagerV1() *map[protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_ZwpXwaylandKeyboardGrabManagerV1]; ok {
		return v.(*map[protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.ZwpXwaylandKeyboardGrabManagerV1]]protocols.Version)
	c.globalBinds[protocols.GlobalID_ZwpXwaylandKeyboardGrabManagerV1] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_XwaylandShellV1() *map[protocols.ObjectID[protocols.XwaylandShellV1]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_XwaylandShellV1]; ok {
		return v.(*map[protocols.ObjectID[protocols.XwaylandShellV1]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.XwaylandShellV1]]protocols.Version)
	c.globalBinds[protocols.GlobalID_XwaylandShellV1] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_WlDataDevice() *map[protocols.ObjectID[protocols.WlDataDevice]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_WlDataDevice]; ok {
		return v.(*map[protocols.ObjectID[protocols.WlDataDevice]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.WlDataDevice]]protocols.Version)
	c.globalBinds[protocols.GlobalID_WlDataDevice] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_WlTouch() *map[protocols.ObjectID[protocols.WlTouch]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_WlTouch]; ok {
		return v.(*map[protocols.ObjectID[protocols.WlTouch]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.WlTouch]]protocols.Version)
	c.globalBinds[protocols.GlobalID_WlTouch] = &m
	return &m
}
func (c *WaylandClient) ensureGlobalMap_ZxdgDecorationManagerV1() *map[protocols.ObjectID[protocols.ZxdgDecorationManagerV1]]protocols.Version {
	if v, ok := c.globalBinds[protocols.GlobalID_ZxdgDecorationManagerV1]; ok {
		return v.(*map[protocols.ObjectID[protocols.ZxdgDecorationManagerV1]]protocols.Version)
	}
	m := make(map[protocols.ObjectID[protocols.ZxdgDecorationManagerV1]]protocols.Version)
	c.globalBinds[protocols.GlobalID_ZxdgDecorationManagerV1] = &m
	return &m
}

// -------- Extra helpers (literal conversion parity) --------

// Debug hook compatible with existing code’s checks
func WaylandDebugTimeOnly() bool {
	return false
}

// Small utility: string for debug
func (c *WaylandClient) String() string {
	return fmt.Sprintf("WaylandClient(%v)", c.conn)
}
