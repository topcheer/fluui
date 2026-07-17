//go:build darwin || freebsd || netbsd || openbsd

package xterm

import "syscall"

var ioctlReadTermios = uint(syscall.TIOCGETA)
var ioctlWriteTermios = uint(syscall.TIOCSETA)
