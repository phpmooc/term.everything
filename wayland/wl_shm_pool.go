package wayland

import (
	"fmt"

	"github.com/mmulet/term.everything/wayland/protocols"
)

type MapState int

const (
	MapStateDestroyed MapState = iota
	MapStateMmapped
	MapStateDestroyWhenBuffersEmpty
)

type BufferInfo struct {
	Offset int32
	Width  int32
	Height int32
	Stride int32
	Format protocols.WlShmFormat_enum
}

type WlShmPool struct {
	MapState          MapState
	Buffers           map[protocols.ObjectID[protocols.WlBuffer]]BufferInfo
	MemMaps           map[protocols.ObjectID[protocols.WlShmPool]]MemMapInfo
	WlShmPoolObjectID protocols.ObjectID[protocols.WlShmPool]
}

func (p *WlShmPool) WlShmPool_create_buffer(
	s protocols.ClientState,
	_objectID protocols.ObjectID[protocols.WlShmPool],
	id protocols.ObjectID[protocols.WlBuffer],
	offset int32,
	width int32,
	height int32,
	stride int32,
	format protocols.WlShmFormat_enum,
) {
	buf := &protocols.WlBuffer{
		Delegate: p,
	}
	AddObject(s, id, buf)
	p.Buffers[id] = BufferInfo{
		Offset: offset,
		Width:  width,
		Height: height,
		Stride: stride,
		Format: format,
	}
}

func (p *WlShmPool) OnDestroyShmPool(s protocols.ClientState, objectID protocols.ObjectID[protocols.WlShmPool]) {
	if memap, ok := p.MemMaps[objectID]; ok {
		memap.Unmap()
	}
	p.MapState = MapStateDestroyed
	RemoveObject(s, objectID)
}

/**
 * This can be called by either on the buffer delegate or the pool delegate
 * @param s
 * @param _object_id Check This!! to see if it is the buffer id or the pool id
 * @returns false because wl_shm_pool hndles remove objet by itself
 */
func (p *WlShmPool) WlShmPool_destroy(
	s protocols.ClientState,
	objectID protocols.ObjectID[protocols.WlShmPool],
) bool {
	buffersEmpty := len(p.Buffers) <= 0
	switch p.MapState {
	case MapStateDestroyed, MapStateDestroyWhenBuffersEmpty:
		return false
	case MapStateMmapped:
		if buffersEmpty {
			p.OnDestroyShmPool(s, objectID)
			return false
		}
		p.MapState = MapStateDestroyWhenBuffersEmpty
		return false
	default:
		panic("unexpected MapState")
	}
}

func (p *WlShmPool) WlShmPool_resize(
	s protocols.ClientState,
	_objectID protocols.ObjectID[protocols.WlShmPool],
	size int32,
) {
	switch p.MapState {
	case MapStateDestroyed:
		return
	case MapStateMmapped, MapStateDestroyWhenBuffersEmpty:
		if memap, ok := p.MemMaps[_objectID]; ok {
			err := memap.Remap(uint64(size))
			if err != nil {
				fmt.Printf("Failed to remap mmap for pool %d: %v\n", _objectID, err)
				p.MapState = MapStateDestroyed
				return
			}
			return
		}
		return
	default:
		panic("unexpected MapState")
	}
}

/**
 * This can be called by either on the buffer delegate or the pool delegate
 * check the name and compare to object id
 * @param _s
 * @param _name
 * @param new_id
 * @param version
 */
func (p *WlShmPool) WlShmPool_on_bind(
	_s protocols.ClientState,
	_name protocols.ObjectID[protocols.WlShmPool],
	_interface_ string,
	newID protocols.ObjectID[protocols.WlShmPool],
	version uint32,
) {
	fmt.Printf("wl_shm_pool on_bind called with new_id: %d, version#: %d\n", newID, version)
}

func MakeWlShmPool(
	client protocols.ClientState, // Assuming ClientState is the client
	wlShmPoolObjectID protocols.ObjectID[protocols.WlShmPool],
	fd protocols.FileDescriptor,
	size int32,
) *protocols.WlShmPool {

	pool := &WlShmPool{
		MapState:          MapStateDestroyed,
		Buffers:           make(map[protocols.ObjectID[protocols.WlBuffer]]BufferInfo),
		MemMaps:           make(map[protocols.ObjectID[protocols.WlShmPool]]MemMapInfo),
		WlShmPoolObjectID: wlShmPoolObjectID,
	}

	memMap, err := NewMemMapInfo(int(fd), uint64(size))
	if err != nil {
		fmt.Printf("Failed to create memmap for pool %d: %v\n", wlShmPoolObjectID, err)
		return &protocols.WlShmPool{Delegate: pool}
	}
	pool.MapState = MapStateMmapped
	pool.MemMaps[wlShmPoolObjectID] = memMap
	return &protocols.WlShmPool{Delegate: pool}
}

func (p *WlShmPool) OnBind(
	s protocols.ClientState,
	_ protocols.AnyObjectID,
	_ string,
	newId_any protocols.AnyObjectID,
	version uint32,
) {
	// No-op
}

func (p *WlShmPool) WlBuffer_destroy(
	s protocols.ClientState,
	bufferObjectID protocols.ObjectID[protocols.WlBuffer],
) bool {
	if _, exists := p.Buffers[bufferObjectID]; !exists {
		fmt.Printf("destroying a buffer that does not exist!, wl_shm_pool_id: %d, buffer_id: %d\n", p.WlShmPoolObjectID, bufferObjectID)
		return true
	}
	delete(p.Buffers, bufferObjectID)
	switch p.MapState {
	case MapStateDestroyed, MapStateMmapped:
		return true
	case MapStateDestroyWhenBuffersEmpty:
		if len(p.Buffers) > 0 {
			return true
		}
		p.OnDestroyShmPool(s, p.WlShmPoolObjectID)
		return true
	default:
		panic("unexpected MapState")
	}
}
