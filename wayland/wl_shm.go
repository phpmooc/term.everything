package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type WlShm struct{}

func (s *WlShm) WlShm_create_pool(
	cs protocols.ClientState,
	_objectID protocols.ObjectID[protocols.WlShm],
	id protocols.ObjectID[protocols.WlShmPool],
	fd *protocols.FileDescriptor,
	size int32,
) {
	AddObject(cs, id, MakeWlShmPool(cs, id, *fd, size))
}

/**
 * Here's what this does according to the docs:
 * Using this request a client can tell the server that it is not going to use the shm object anymore.
 *   Objects created via this interface remain unaffected.
 *
 * So I guess remove the object from the client, but leave all pools alone?
 * @param s
 * @param _object_id
 */
func (s *WlShm) WlShm_release(
	_cs protocols.ClientState,
	_objectID protocols.ObjectID[protocols.WlShm],
) bool {
	return true
}

func (s *WlShm) OnBind(
	cs protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {

	// WlShm_on_bind(
	// 	cs protocols.ClientState,
	// 	_name protocols.ObjectID[protocols.WlShm],
	// 	_interface_ string,
	// 	newID protocols.ObjectID[protocols.WlShm],
	// 	_version uint32,
	// ) {
	newID := protocols.ObjectID[protocols.WlShm](newId_any)

	protocols.WlShm_format(cs, newID, protocols.WlShmFormat_enum_argb8888)
}

// Helper to construct a protocol object with this delegate (like static make() in TS)
func MakeWlShm() *protocols.WlShm {
	return &protocols.WlShm{Delegate: &WlShm{}}
}
