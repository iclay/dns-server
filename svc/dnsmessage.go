package svc

import (
	"fmt"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

func qString(q dnsmessage.Question) string {
	b := make([]byte, q.Name.Length+2)
	for i := 0; i < int(q.Name.Length); i++ {
		b[i] = q.Name.Data[i]
	}
	b[q.Name.Length] = uint8(q.Type >> 8)
	b[q.Name.Length+1] = uint8(q.Type)

	return string(b)
}

func ntString(rName dnsmessage.Name, rType dnsmessage.Type) string {
	b := make([]byte, rName.Length+2)
	for i := 0; i < int(rName.Length); i++ {
		b[i] = rName.Data[i]
	}
	b[rName.Length] = uint8(rType >> 8)
	b[rName.Length+1] = uint8(rType)

	return string(b)
}

func rString(r dnsmessage.Resource) string {
	var sb strings.Builder
	sb.Write(r.Header.Name.Data[:])
	sb.WriteString(r.Header.Type.String())
	sb.WriteString(r.Body.GoString())

	return sb.String()
}

func pString(p Packet) string {
	return fmt.Sprint(p.message.ID)
}
