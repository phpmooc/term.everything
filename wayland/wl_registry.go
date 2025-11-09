package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type WlRegistryDelegateImpl struct{}

func (w *WlRegistryDelegateImpl) WlRegistry_bind(s protocols.ClientState, object_id protocols.ObjectID[protocols.WlRegistry], name uint32, idInterface string, idVersion uint32, idID protocols.AnyObjectID) {
	object := s.GetObject(protocols.AnyObjectID(name))
	s.AddObject(idID, object)
	version := protocols.Version(idVersion)

	switch name {
	case uint32(protocols.GlobalID_WlShm):
	case uint32(protocols.GlobalID_WlSeat):
	case uint32(protocols.GlobalID_WlOutput):
	case uint32(protocols.GlobalID_WlKeyboard):
	case uint32(protocols.GlobalID_WlPointer):
	case uint32(protocols.GlobalID_WlTouch):
	case uint32(protocols.GlobalID_WlDataDevice):
	case uint32(protocols.GlobalID_ZwpXwaylandKeyboardGrabManagerV1):
		s.AddGlobalBind(protocols.GlobalID(name), idID, version)
		// const set = s.global_binds.get(name) ?? new Set();
		// set.add(id_id);
		// s.global_binds.set(name, set);
	}
	// TODO turn this back on
	//  if (wayland_debug_time_only()) {
	//     console.log(
	//       `client#${s.client_socket} ${id_interface}_on_bind(name:`,
	//       name,
	//       ",interface:",
	//       id_interface,
	//       ",id:",
	//       id_id,
	//       ",version:",
	//       id_version,
	//       `)`
	//     );
	//   }

	if object != nil && object != 0 {

		delegate := object.(protocols.WaylandObject[protocols.OnBindable]).Delegate()
		delegate.OnBind(
			s,
			protocols.AnyObjectID(name),
			idInterface,
			idID,
			idVersion,
		)
	}
}

func (w *WlRegistryDelegateImpl) OnBind(
	s protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
}

func MakeWlRegistry() *protocols.WlRegistry {
	return &protocols.WlRegistry{
		Delegate: &WlRegistryDelegateImpl{},
	}
}
