package svc

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

var ip_mask = "255.255.255.0"

func subNetMaskToLen(netmask string, domain string) (int, error) {
	ipSplitArr := strings.Split(netmask, ".")
	if len(ipSplitArr) != 4 {
		return 0, fmt.Errorf("netmask:%v is not valid, pattern should like: 255.255.255.0, domain:%v", netmask, domain)
	}
	ipv4MaskArr := make([]byte, 4)
	for i, value := range ipSplitArr {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("ipMaskToInt call strconv.Atoi error:%v string value is:%v, netMask:%v, domain:%v", err, value, netmask, domain)
		}
		if intValue > 255 {
			return 0, fmt.Errorf("netmask cannot greater than 255, current value is:%v, netMask:%v,domain:%v", value, netmask, domain)
		}
		ipv4MaskArr[i] = byte(intValue)
	}

	ones, _ := net.IPv4Mask(ipv4MaskArr[0], ipv4MaskArr[1], ipv4MaskArr[2], ipv4MaskArr[3]).Size()
	return ones, nil
}

func parseIP(ipString string) (domain string, err error) {
	c := strings.Split(ipString, ".")
	if len(c) == net.IPv6len {
		var ipv6Byte [16]byte
		for i, _ := range c {
			if tmp, err := strconv.Atoi(c[i]); err == nil {
				ipv6Byte[i] = byte(tmp)
			} else {
				return "", err
			}
		}
		var ip_v6 net.IP = net.IP(ipv6Byte[:])
		if ip_v6.To16() != nil {
			if len, err := subNetMaskToLen(ip_mask, ip_v6.String()); err != nil {
				return "", err
			} else {
				return fmt.Sprintf("%v/%v", ip_v6.String(), len), nil
			}
			return ip_v6.String(), nil
		}
	} else if len(c) == net.IPv4len {
		var ipv4Byte [4]byte
		for i, _ := range c {
			if tmp, err := strconv.Atoi(c[i]); err == nil {
				ipv4Byte[i] = byte(tmp)
			} else {
				return "", err
			}
		}
		var ip_v4 net.IP = net.IP(ipv4Byte[:])
		if ip_v4.To4() != nil {
			if len, err := subNetMaskToLen(ip_mask, ip_v4.String()); err != nil {
				return "", err
			} else {
				return fmt.Sprintf("%v/%v", ip_v4.String(), len), nil
			}
		}
	}
	err = fmt.Errorf("not support parseIP, domain=%v", ipString)
	return
}
