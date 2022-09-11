//go:build !linux
// +build !linux

package log

import "io"

func GetSyslog(protocol, conn, tag string) (io.Writer, error) {
	return nil, nil
}
