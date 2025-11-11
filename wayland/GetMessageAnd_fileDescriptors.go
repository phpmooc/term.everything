package wayland

import (
	"errors"
	"io"
	"net"
	"syscall"
	"time"
)

const (
	GetMessage_timeout      = 1 * time.Millisecond
	GetMessage_maxFDsInCmsg = 10  // matches C++: CMSG_SPACE(sizeof(int) * 10)
	GetMessage_hardFDLimit  = 255 // matches C++ guard in the copy loop
	GetMessage_intSizeBytes = 4   // sizeof(int) on Linux
)

var oob = make([]byte, syscall.CmsgSpace(GetMessage_intSizeBytes*GetMessage_maxFDsInCmsg))

func GetMessageAndFileDescriptors(conn *net.UnixConn, buf []byte) (n int, fds []int, err error) {

	// 10ms "select"-style timeout
	if err := conn.SetReadDeadline(time.Now().Add(GetMessage_timeout)); err != nil {
		return 0, nil, err
	}
	defer conn.SetReadDeadline(time.Time{})

	n, oobn, _, _, rerr := conn.ReadMsgUnix(buf, oob)

	// Timeout -> continue with no data/fds.
	if ne, ok := rerr.(net.Error); ok && ne.Timeout() {
		return 0, nil, nil
	}
	if rerr != nil {
		if errors.Is(rerr, io.EOF) {
			return n, nil, nil
		}
		// Treat as terminal like the C++ (returns false).
		return n, nil, rerr
	}
	if n == 0 {
		// EOF on stream
		return 0, nil, nil
	}

	// Parse as many rights as fit; ignore truncation like the C++.
	if oobn > 0 {
		if cmsgs, perr := syscall.ParseSocketControlMessage(oob[:oobn]); perr == nil {
			for _, cmsg := range cmsgs {
				if rights, rerr := syscall.ParseUnixRights(&cmsg); rerr == nil && len(rights) > 0 {
					fds = append(fds, rights...)
					if len(fds) >= GetMessage_hardFDLimit {
						fds = fds[:GetMessage_hardFDLimit]
						break
					}
				}
			}
		}
	}

	return n, fds, nil
}
