package eagletunnel

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	SOCKS_ERROR = iota
	SOCKS_CONNECT
	SOCKS_BIND
	SOCKS_UDP
)

type Socks5 struct {
}

func (conn *Socks5) handle(request Request, tunnel *Tunnel) bool {
	var result bool
	version := request.requestMsg[0]
	if version == '\u0005' {
		reply := "\u0005\u0000"
		count, _ := tunnel.writeLeft([]byte(reply))
		if count > 0 {
			var buffer = make([]byte, 1024)
			count, _ = tunnel.readLeft(buffer)
			if count >= 2 {
				cmdType := buffer[1]
				switch cmdType {
				case SOCKS_CONNECT:
					return conn.handleTCPReq(buffer[:count], tunnel)
				default:
				}
			}
		}
	}
	return result
}

func (conn *Socks5) handleTCPReq(req []byte, tunnel *Tunnel) bool {
	var result bool
	ip := conn.getIP(req)
	port := conn.getPort(req)
	if ip != "" && port > 0 {
		var reply string
		var e = NetArg{ip: ip, port: port, tunnel: tunnel, theType: ET_TCP}
		conn := EagleTunnel{}
		if conn.send(&e) {
			reply = "\u0005\u0000\u0000\u0001\u0000\u0000\u0000\u0000\u0000\u0000"
			_, err := tunnel.writeLeft([]byte(reply))
			result = err == nil
		} else {
			reply = "\u0005\u0001\u0000\u0001\u0000\u0000\u0000\u0000\u0000\u0000"
			tunnel.writeLeft([]byte(reply))
		}
	}
	return result
}

func (conn *Socks5) getIP(request []byte) string {
	var ip string
	var destype = request[3]
	switch destype {
	case 1:
		ip = fmt.Sprintf("%d.%d.%d.%d", request[4], request[5], request[6], request[7])
	case 3:
		len := request[4]
		domain := string(request[5 : 5+len])
		newConn := EagleTunnel{}
		e := NetArg{domain: domain, theType: ET_DNS}
		if newConn.send(&e) {
			ip = e.ip
		}
	}
	return ip
}

func (conn *Socks5) getPort(request []byte) int {
	destype := request[3]
	var port int16
	var buffer []byte
	var err error
	switch destype {
	case 1:
		buffer = request[8:10]
	case 3:
		len := request[4]
		buffer = request[5+len : 7+len]
	default:
		buffer = make([]byte, 0)
		err = errors.New("invalid destype")
	}
	if err == nil {
		port = int16(binary.BigEndian.Uint16(buffer))
	}
	return int(port)
}
