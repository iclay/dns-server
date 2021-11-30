package sock

import (
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

type listener struct {
	fd          int
	inaddr      net.Addr
	proto, addr string //协议类型，暂时先做udp支持
}

func (l *listener) normative() (err error) {
	switch l.proto {
	case "udp", "udp4", "udp6":
		l.fd, l.inaddr, err = UDPSocket(l.proto, l.addr)
	case "tcp", "tcp4", "tcp6":
		err = fmt.Errorf("not support tcp current")
	default:
		_, err = fmt.Fprintf(os.Stdout, "not supprot this prototype, prototype=%v", l.proto)
	}
	return err
}

func UDPSocket(proto, addr string) (fd int, inaddr net.Addr, err error) {

	udpAddr, err := net.ResolveUDPAddr(proto, addr)
	if err != nil {
		return
	}
	proto, err = determineUDPProto(proto, udpAddr)
	if err != nil {
		return
	}
	var (
		family   int
		IPV6Only bool
		sockAddr unix.Sockaddr
		listenfd int
	)
	switch proto {
	case "udp4":
		if udpAddr.IP == nil {
			err = fmt.Errorf("udp_prototype=%v, udpAddr.IP=nil", proto)
			return
		}
		sa4 := &unix.SockaddrInet4{}
		if len(udpAddr.IP) == 16 {
			copy(sa4.Addr[:], udpAddr.IP[12:16]) //如果是通过v4到v6的转换，则取第12-16位
		} else {
			copy(sa4.Addr[:], udpAddr.IP)
		}
		sa4.Port = udpAddr.Port
		family = unix.AF_INET
		sockAddr = sa4
	case "udp6":
		IPV6Only = true
		fallthrough
	case "udp":
		if udpAddr.IP == nil {
			err = fmt.Errorf("udp_prototype=%v, udpAddr.IP=nil", proto)
			return
		}
		sa6 := &unix.SockaddrInet6{}
		copy(sa6.Addr[:], udpAddr.IP)
		sa6.Port = udpAddr.Port
		family = unix.AF_INET6
		if udpAddr.Zone != "" {
			var iface *net.Interface
			iface, err = net.InterfaceByName(udpAddr.Zone)
			if err != nil {
				return
			}

			sa6.ZoneId = uint32(iface.Index)
		}
		sockAddr = sa6
	default:
		err = fmt.Errorf("udp_prototype, not support for proto=%v", proto)
		return
	}

	if listenfd, err = unix.Socket(family, unix.SOCK_DGRAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_UDP); err != nil {
		return
	}
	defer func() {
		if err != nil {
			unix.Close(listenfd)
		}
	}()
	if family == unix.AF_INET6 && IPV6Only { //如果proto为"udp6"，责禁止v4到v6的转换
		if err = unix.SetsockoptInt(listenfd, unix.IPPROTO_IPV6, unix.IPV6_V6ONLY, 1); err != nil {
			return
		}
	}
	if err = os.NewSyscallError("setsockopt", unix.SetsockoptInt(listenfd, unix.SOL_SOCKET, unix.SO_BROADCAST, 1)); err != nil {
		return
	}
	err = os.NewSyscallError("bind", unix.Bind(listenfd, sockAddr))
	return listenfd, udpAddr, err
}

func determineUDPProto(proto_in string, addr *net.UDPAddr) (proto_out string, err error) {
	if addr.IP.To4() != nil {
		return "udp4", nil
	}
	if addr.IP.To16() != nil {
		return "udp6", nil
	}
	switch proto_in {
	case "udp", "udp4", "udp6":
		return proto_in, nil
	}
	return "", fmt.Errorf("not support this protprype", proto_in)
}

func InitListener(address string) (l *listener, err error) {
	proto := "tcp"
	var addr string
	proto = strings.ToLower(address)
	if strings.Contains(address, "//") {
		pair := strings.Split(address, "//")
		proto = pair[0]
		addr = pair[1]
	}
	l = &listener{
		proto: proto,
		addr:  addr,
	}
	if err = l.normative(); err != nil {
		return
	}
	return l, nil
}
