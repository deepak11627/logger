//go:build !windows
// +build !windows

package log

import (
	"io"
	"log/syslog"
)

func GetSyslog(protocol, conn, tag string) (io.Writer, error) {
	sysLog, err := syslog.Dial(
		protocol,
		conn,
		syslog.LOG_DEBUG|syslog.LOG_DAEMON,
		tag,
	)
	if err != nil {
		return nil, err
	}
	return sysLog, nil
}
