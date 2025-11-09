package wayland

import "github.com/mmulet/term.everything/wayland/protocols"

type wl_display struct{}

func (wd *wl_display) WlDisplay_sync(s protocols.ClientState, _object_id protocols.ObjectID[protocols.WlDisplay], callback protocols.ObjectID[protocols.WlCallback]) {
	protocols.WlCallback_done(s, callback, 0)
}

func (wd *wl_display) WlDisplay_get_registry(s protocols.ClientState, _object_id protocols.ObjectID[protocols.WlDisplay], registry protocols.ObjectID[protocols.WlRegistry]) {
	registry_object := MakeWlRegistry()
	s.AddObject(protocols.AnyObjectID(registry), registry_object)
	for _, global := range protocols.AdvertisedGlobalObjectNames {
		protocols.WlRegistry_global(s, registry, uint32(global.Id), global.Name, global.Version)
	}
}

func (wd *wl_display) OnBind(
	s protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	_ protocols.AnyObjectID,
	version uint32,
) {
}

func Make() *protocols.WlDisplay {
	return &protocols.WlDisplay{
		Delegate: &wl_display{},
	}
}
