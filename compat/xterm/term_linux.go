//go:build linux

package xterm

import "syscall"

var ioctlReadTermios = uint(syscall.TCGETS)
var ioctlWriteTermios = uint(syscall.TCSETS)
