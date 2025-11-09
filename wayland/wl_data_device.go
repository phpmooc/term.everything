package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type wl_data_device struct {
	Seat protocols.ObjectID[protocols.WlSeat]
}

func (w *wl_data_device) WlDataDevice_start_drag(
	_s protocols.ClientState,
	_object_id protocols.ObjectID[protocols.WlDataDevice],
	_source *protocols.ObjectID[protocols.WlDataSource],
	_origin protocols.ObjectID[protocols.WlSurface],
	_icon *protocols.ObjectID[protocols.WlSurface],
	_serial uint32,
) {
	/** @TODO: Implement wl_data_device_start_drag */
}

func (w *wl_data_device) WlDataDevice_set_selection(
	_s protocols.ClientState,
	_object_id protocols.ObjectID[protocols.WlDataDevice],
	_source *protocols.ObjectID[protocols.WlDataSource],
	_serial uint32,
) {
	/** @TODO: Implement wl_data_device_set_selection */
}

func (w *wl_data_device) WlDataDevice_release(
	_s protocols.ClientState,
	_object_id protocols.ObjectID[protocols.WlDataDevice],
) bool {
	return true
}

func (w *wl_data_device) WlDataDevice_on_bind(
	_s protocols.ClientState,
	_name protocols.ObjectID[protocols.WlDataDevice],
	_interface_ string,
	_new_id protocols.ObjectID[protocols.WlDataDevice],
	_version_number uint32,
) {
	/** @TODO: Implement wl_data_device_on_bind */
}

func MakeWlDataDevice(seat protocols.ObjectID[protocols.WlSeat]) *wl_data_device {
	return &wl_data_device{Seat: seat}
}
