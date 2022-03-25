package nic

import (
	"os"
	"syscall"
	"unsafe"
)

// https://github.com/golang/go/blob/go1.4.1/src/net/interface_windows.go#L22-L39
func GetAdapterList() (*syscall.IpAdapterInfo, error) {
	b := make([]byte, 1000)
	l := uint32(len(b))
	a := (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
	// TODO(mikio): GetAdaptersInfo returns IP_ADAPTER_INFO that
	// contains IPv4 address list only. We should use another API
	// for fetching IPv6 stuff from the kernel.
	err := syscall.GetAdaptersInfo(a, &l)
	if err == syscall.ERROR_BUFFER_OVERFLOW {
		b = make([]byte, l)
		a = (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
		err = syscall.GetAdaptersInfo(a, &l)
	}
	if err != nil {
		return nil, os.NewSyscallError("GetAdaptersInfo", err)
	}
	return a, nil
}

// https://github.com/golang/go/blob/go1.4.1/src/net/interface_windows.go#L13-L20
func BytePtrToString(p *uint8) string {
	a := (*[10000]uint8)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(a[:i])
}

func GetAdapterNames() (ret []string, err error) {
	ai, err := GetAdapterList()
	if err != nil {
		return
	}

	for ; ai != nil; ai = ai.Next {
		ret = append(ret, BytePtrToString(&ai.AdapterName[0]))

	}
	return
}
