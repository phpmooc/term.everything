package wayland

import (
	"errors"
	"io"
	"net"
	"syscall"
	"time"
)

func GetMessageAndFileDescriptors(conn *net.UnixConn, buf []byte) (n int, fds []int, err error) {
	const (
		timeout      = 1 * time.Millisecond
		maxFDsInCmsg = 10  // matches C++: CMSG_SPACE(sizeof(int) * 10)
		hardFDLimit  = 255 // matches C++ guard in the copy loop
		intSizeBytes = 4   // sizeof(int) on Linux
	)

	oob := make([]byte, syscall.CmsgSpace(maxFDsInCmsg*intSizeBytes))

	// 10ms "select"-style timeout
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
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
					if len(fds) >= hardFDLimit {
						fds = fds[:hardFDLimit]
						break
					}
				}
			}
		}
	}

	return n, fds, nil
}
