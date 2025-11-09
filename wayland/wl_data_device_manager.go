package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type WlDataDeviceManagerImpl struct{}

func (w *WlDataDeviceManagerImpl) WlDataDeviceManager_create_data_source(s protocols.ClientState, _object_id protocols.ObjectID[protocols.WlDataDeviceManager], id protocols.ObjectID[protocols.WlDataSource]) {
	s.AddObject(protocols.AnyObjectID(id), MakeWlDataSource())
}

func (w *WlDataDeviceManagerImpl) WlDataDeviceManager_get_data_device(s protocols.ClientState, _object_id protocols.ObjectID[protocols.WlDataDeviceManager], id protocols.ObjectID[protocols.WlDataDevice], seat protocols.ObjectID[protocols.WlSeat]) {
	s.AddObject(protocols.AnyObjectID(id), MakeWlDataDevice(seat))
	/** @TODO: Implement wl_data_device_manager_get_data_device */
}

func (w *WlDataDeviceManagerImpl) OnBind(
	s protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	_ protocols.AnyObjectID,
	version uint32,
) {
	/** @TODO: Implement wl_data_device_manager_on_bind */
}

func MakeWlDataDeviceManager() *protocols.WlDataDeviceManager {
	return &protocols.WlDataDeviceManager{
		Delegate: &WlDataDeviceManagerImpl{},
	}
}
