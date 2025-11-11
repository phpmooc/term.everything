package wayland

//go:generate sh -c "go run ./generate ./protocols . $(go list) WlSurface XdgPositioner XdgSurface WlPointer WlSubsurface XdgToplevel"
