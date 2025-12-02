//go:build with_utls

package uap

import (
	"net"
	"reflect"
	"unsafe"

	"github.com/sagernet/sing/common"
	utls "github.com/sagernet/utls"
)

func init() {
	tlsRegistry = append(tlsRegistry, func(conn net.Conn) (loaded bool, netConn net.Conn, reflectType reflect.Type, reflectPointer uintptr) {
		tlsConn, loaded := common.Cast[*utls.UConn](conn)
		if !loaded {
			return
		}
		return true, tlsConn.NetConn(), reflect.TypeOf(tlsConn.Conn).Elem(), uintptr(unsafe.Pointer(tlsConn.Conn))
	})
}
