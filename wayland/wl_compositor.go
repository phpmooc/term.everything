package wayland

import (
	"github.com/mmulet/term.everything/wayland/protocols"
)

type WlCompositor struct{}

func (c *WlCompositor) WlCompositor_create_surface(
	s protocols.ClientState,
	_ protocols.ObjectID[protocols.WlCompositor],
	id protocols.ObjectID[protocols.WlSurface],
) {
	surface := MakeWlSurface()

	AddObject(s, id, surface)

	// // s.bound_compositor_info?.surfaces.set(id, new Surface_Info(surface, 1));
	// // console.log("create surface", id);
	// /**
	//  * @TODO check this code to see if it is needed
	//  */
	// // s.get_global_binds(Global_Ids.wl_output)?.forEach((output_id) => {
	// //   console.log("Output enter", output_id, id);
	// //   wl_surface.enter(s, id, output_id);
	// // });
	//
	// // s.get_global_binds(Global_Ids.wl_keyboard)?.forEach((keyboard_id) => {
	// //   console.log("Keyboard enter", keyboard_id, id);
	// //   wl_keyboard.enter(s, keyboard_id, 0, id, []);
	// // });
	//
	// // s.get_global_binds(Global_Ids.wl_pointer)?.forEach((pointer_id) => {
	// //   console.log("Pointer enter", pointer_id, id);
	// //   wl_pointer.enter(
	// //     s,
	// //     pointer_id,
	// //     Math.round(Math.random() * 10_000),
	// //     id,
	// //     0,
	// //     0
	// //   );
	// //   wl_pointer.frame(s, pointer_id);
	// // });
}

func (c *WlCompositor) WlCompositor_create_region(
	s protocols.ClientState,
	_ protocols.ObjectID[protocols.WlCompositor],
	id protocols.ObjectID[protocols.WlRegion],
) {
	region := MakeWlRegion()
	AddObject(s, id, region)
}

func (c *WlCompositor) OnBind(
	s protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	_ protocols.AnyObjectID,
	version uint32,
) {
	s.SetCompositorVersion(version)
}

func MakeWlCompositor() *protocols.WlCompositor {
	return &protocols.WlCompositor{
		Delegate: &WlCompositor{},
	}
}
