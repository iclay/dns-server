package svc

import (
	"dns/match"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/dns/dnsmessage"
)

type DNSServer interface {
	Listen()
	Query(Packet)
}

var (
	log  *logrus.Logger
	wlog *logrus.Logger
	blog *logrus.Logger
)

func SetLogger(logMap map[string]*logrus.Logger) {
	for k, v := range logMap {
		switch k {
		case "log":
			log = v
		case "wlog":
			wlog = v

		case "blog":
			blog = v
		default:
		}
	}
}

type DNSService struct {
	conn       *net.UDPConn
	book       store
	memo       addrBag
	forwarders []net.UDPAddr
	opt        *Options
}

type Packet struct {
	addr    net.UDPAddr
	message dnsmessage.Message
}

const (
	udpPort   int = 53
	packetLen int = 512
)

var (
	errTypeNotSupport = errors.New("type not support")
	errIPInvalid      = errors.New("invalid IP address")
)

func (s *DNSService) Listen() {
	var err error
	s.conn, err = net.ListenUDP("udp", &net.UDPAddr{Port: udpPort})
	if err != nil {
		log.Fatal(err)
	}
	defer s.conn.Close()

	for {
		buf := make([]byte, packetLen)
		_, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			log.Error(err)
			continue
		}
		var m dnsmessage.Message
		err = m.Unpack(buf)
		if err != nil {
			log.Error(err)
			continue
		}
		if len(m.Questions) == 0 {
			continue
		}
		go s.Query(Packet{*addr, m})
	}
}
func (s *DNSService) filterDomin(domain string) bool {
	// return !match.DomainMatch(domain, s.opt.blacklistMap) || match.DomainMatch(domain, s.opt.whitelistMap)
	return match.DomainMatch(domain, s.opt.whitelistMap)
}

func (s *DNSService) checkAnswers_bak(searchType string, question dnsmessage.Question, answers []dnsmessage.Resource) {
	for _, answer := range answers {
		switch t := answer.Body.(type) {
		case *dnsmessage.AResource:
			body := answer.Body.(*dnsmessage.AResource)
			log.Debugf("ARSerouce response, question=%v answer=%v", question.Name.String(), body.GoString())
		case *dnsmessage.PTRResource:
			body := answer.Body.(*dnsmessage.PTRResource)
			log.Debugf("PTRResource response, question=%+v answer=%v", question.Name.String(), body.PTR.GoString())
			if s.filterDomin(body.PTR.String()) {
				fmt.Fprint(os.Stdout, body.PTR.String()+"\n")
				if err := s.opt.ptrHookAction([]string{body.PTR.String()}); err != nil {
					log.Error(err)
				}
			}
		default:
			log.Debugf("not support this type, dnsmessageType=%+v", t)
		}
	}

}

func printByteSlice(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	buf := make([]byte, 0, 5*len(b))
	buf = printUint8Bytes(buf, uint8(b[0]))
	for _, n := range b[1:] {
		buf = append(buf, '.')
		buf = printUint8Bytes(buf, uint8(n))
	}
	return string(buf)
}
func printUint8Bytes(buf []byte, i uint8) []byte {
	b := byte(i)
	if i >= 100 {
		buf = append(buf, b/100+'0')
	}
	if i >= 10 {
		buf = append(buf, b/10%10+'0')
	}
	return append(buf, b%10+'0')
}

func (s *DNSService) checkAnswers(searchType string, question dnsmessage.Question, answers []dnsmessage.Resource) {

	var ans []string
	if len(answers) > 0 {
		answer := answers[0]
		switch t := answer.Body.(type) {
		case *dnsmessage.AResource:
			body := answer.Body.(*dnsmessage.AResource)
			log.Debugf("ARSource response, question=%v answer=%v", question.Name.String(), printByteSlice(body.A[:]))
			if s.filterDomin(printByteSlice(body.A[:])) {
				for _, v := range answers {
					body = v.Body.(*dnsmessage.AResource)
					ans = append(ans, printByteSlice(body.A[:]))
				}
				if len(ans) > 0 {
					wlog.Debugf("ARSource PutWgvpnResource, question=%v answer=%v", question.Name.String(), ans)
					err := s.opt.aHookAction(ans)
					if err != nil {
						log.Errorf("ARSourcehookaction PutWgvpnResource error, question=%v,ans=%v,err=%v", question.Name.String(), ans, err)
					}
				}
			} else {
				blog.Debugf("ARSerouce response, question=%v answer=%v", question.Name.String(), body.GoString())
			}
		case *dnsmessage.PTRResource:
			body := answer.Body.(*dnsmessage.PTRResource)
			log.Debugf("PTRResource response, question=%+v answer=%v", question.Name.String(), body.PTR.GoString())
			if s.filterDomin(body.PTR.String()) {
				fmt.Fprint(os.Stdout, body.PTR.String()+"\n")
				for _, v := range answers {
					body = v.Body.(*dnsmessage.PTRResource)
					ans = append(ans, body.PTR.String())
				}
				if len(ans) > 0 {
					wlog.Debugf("PTRSource PutWgvpnResource, question=%v answer=%v", question.Name.String(), ans)
					err := s.opt.aHookAction(ans)
					if err != nil {
						log.Errorf("PTRSourcehookaction PutWgvpnResource error, question=%v,ans=%v,err=%v", question.Name.String(), ans, err)
					}
				}
			} else {
				blog.Debugf("PTRSource response, question=%v answer=%v", question.Name.String(), body.PTR.String())
			}
		case *dnsmessage.AAAAResource:
			body := answer.Body.(*dnsmessage.AAAAResource)
			log.Debugf("AAAAResource response, question=%+v answer=%v", question.Name.String(), printByteSlice(body.AAAA[:]))
			if s.filterDomin(printByteSlice(body.AAAA[:])) {
				fmt.Fprint(os.Stdout, printByteSlice(body.AAAA[:])+"\n")
				for _, v := range answers {
					body = v.Body.(*dnsmessage.AAAAResource)
					ans = append(ans, printByteSlice(body.AAAA[:]))
				}
				if len(ans) > 0 {
					wlog.Debugf("AAAASource PutWgvpnResource, question=%v answer=%v", question.Name.String(), ans)
					err := s.opt.aHookAction(ans)
					if err != nil {
						log.Errorf("AAAASourcehookaction PutWgvpnResource error, question=%v,ans=%v,err=%v", question.Name.String(), ans, err)
					}
				}
			} else {
				blog.Debugf("AAAAResource response, question=%+v answer=%v", question.Name.String(), printByteSlice(body.AAAA[:]))
			}

		default:
			log.Debugf("not support this type, dnsmessageType=%+v", t)
		}

	}

}

func (s *DNSService) checkQuestion(searchType string, question dnsmessage.Question, answers []dnsmessage.Resource) {
	que := question.Name.String()
	if strings.HasSuffix(que, ".") {
		que = que[:len(que)-1]
	}
	if !s.filterDomin(que) {
		log.Errorf("filterDomin error, question=%+v,que=%v, type=%v", question.Name.String(), que, searchType)
		return
	}
	log.Errorf("filterDomin question=%+v,que=%v, type=%v", question.Name.String(), que, searchType)
	var ans []string
	if len(answers) > 0 {
		answer := answers[0]
		switch t := answer.Body.(type) {
		case *dnsmessage.AResource:
			body := answer.Body.(*dnsmessage.AResource)
			log.Debugf("ARSource response, question=%v answer=%v", question.Name.String(), printByteSlice(body.A[:]))
			for _, v := range answers {
				body = v.Body.(*dnsmessage.AResource)
				ans = append(ans, printByteSlice(body.A[:]))
			}
			if len(ans) > 0 {
				wlog.Debugf("ARSource PutWgvpnResource, question=%v answer=%v", question.Name.String(), ans)
				err := s.opt.aHookAction(ans)
				if err != nil {
					log.Errorf("ARSourcehookaction PutWgvpnResource error, question=%v,ans=%v,err=%v", question.Name.String(), ans, err)
				}
			}

		case *dnsmessage.PTRResource:
			body := answer.Body.(*dnsmessage.PTRResource)
			log.Debugf("PTRResource response, question=%+v answer=%v", question.Name.String(), body.PTR.GoString())
			fmt.Fprint(os.Stdout, body.PTR.String()+"\n")
			for _, v := range answers {
				body = v.Body.(*dnsmessage.PTRResource)
				ans = append(ans, body.PTR.String())
			}
			if len(ans) > 0 {
				wlog.Debugf("PTRSource PutWgvpnResource, question=%v answer=%v", question.Name.String(), ans)
				err := s.opt.aHookAction(ans)
				if err != nil {
					log.Errorf("PTRSourcehookaction PutWgvpnResource error, question=%v,ans=%v,err=%v", question.Name.String(), ans, err)
				}
			}
		case *dnsmessage.AAAAResource:
			body := answer.Body.(*dnsmessage.AAAAResource)
			log.Debugf("AAAAResource response, question=%+v answer=%v", question.Name.String(), printByteSlice(body.AAAA[:]))
			fmt.Fprint(os.Stdout, printByteSlice(body.AAAA[:])+"\n")
			for _, v := range answers {
				body = v.Body.(*dnsmessage.AAAAResource)
				ans = append(ans, printByteSlice(body.AAAA[:]))
			}
			if len(ans) > 0 {
				wlog.Debugf("AAAASource PutWgvpnResource, question=%v answer=%v", question.Name.String(), ans)
				err := s.opt.aHookAction(ans)
				if err != nil {
					log.Errorf("AAAASourcehookaction PutWgvpnResource error, question=%v,ans=%v,err=%v", question.Name.String(), ans, err)
				}
			}
		default:
			log.Debugf("not support this type, dnsmessageType=%+v", t)
		}

	}
}
func (s *DNSService) Query(p Packet) {
	// 该response是从顶级域名返回结果发送给client
	if p.message.Header.Response {
		pKey := pString(p)
		if addrs, ok := s.memo.get(pKey); ok {
			s.checkQuestion("forward", p.message.Questions[0], p.message.Answers)
			for _, addr := range addrs {
				go sendPacket(s.conn, p.message, addr)
			}
			s.memo.remove(pKey)
			q := p.message.Questions[0]
			go s.saveBulk(qString(q), p.message.Answers)
		}
		return
	}

	q := p.message.Questions[0]
	val, ok := s.book.get(qString(q))
	if ok {
		p.message.Response = true
		p.message.Answers = append(p.message.Answers, val...) //如果本地有缓存，则直接发送至client
		s.checkQuestion("cache", q, p.message.Answers)
		go sendPacket(s.conn, p.message, p.addr)
	} else {
		for i := 0; i < len(s.forwarders); i++ { //如果本地没有，直接转发包至顶级域名递归查询
			s.memo.set(pString(p), p.addr)
			go sendPacket(s.conn, p.message, s.forwarders[i])
		}
	}
}

func sendPacket(conn *net.UDPConn, message dnsmessage.Message, addr net.UDPAddr) {
	packed, err := message.Pack()
	if err != nil {
		log.Println(err)
		return
	}

	_, err = conn.WriteToUDP(packed, &addr)
	if err != nil {
		log.Println(err)
	}
}

func NewDNService(rwDirPath string, forwarders []net.UDPAddr, opts ...Option) *DNSService {
	dns := &DNSService{
		book:       store{data: make(map[string]entry), rwDirPath: rwDirPath},
		memo:       addrBag{data: make(map[string][]net.UDPAddr)},
		forwarders: forwarders,
		opt:        loadOptions(opts...),
	}
	dns.book.load()

	go dns.Listen()
	return dns
}

func (s *DNSService) save(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool {
	ok := s.book.set(key, resource, old)
	go s.book.save()

	return ok
}

func (s *DNSService) saveBulk(key string, resources []dnsmessage.Resource) {
	s.book.override(key, resources)
	go s.book.save()
}

func (s *DNSService) all() []get {
	book := s.book.clone()
	var recs []get
	for _, r := range book {
		for _, v := range r.Resources {
			body := v.Body.GoString()
			i := strings.Index(body, "{")
			recs = append(recs, get{
				Host: v.Header.Name.String(),
				TTL:  v.Header.TTL,
				Type: v.Header.Type.String()[4:],
				Data: body[i : len(body)-1],
			})
		}
	}
	return recs
}

func (s *DNSService) remove(key string, r *dnsmessage.Resource) bool {
	ok := s.book.remove(key, r)
	if ok {
		go s.book.save()
	}
	return ok
}

func toResource(req request) (dnsmessage.Resource, error) {
	rName, err := dnsmessage.NewName(req.Host)
	none := dnsmessage.Resource{}
	if err != nil {
		return none, err
	}

	var rType dnsmessage.Type
	var rBody dnsmessage.ResourceBody

	switch req.Type {
	case "A":
		rType = dnsmessage.TypeA
		ip := net.ParseIP(req.Data)
		if ip == nil {
			return none, errIPInvalid
		}
		rBody = &dnsmessage.AResource{A: [4]byte{ip[12], ip[13], ip[14], ip[15]}}
	case "NS":
		rType = dnsmessage.TypeNS
		ns, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.NSResource{NS: ns}
	case "CNAME":
		rType = dnsmessage.TypeCNAME
		cname, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.CNAMEResource{CNAME: cname}
	case "SOA":
		rType = dnsmessage.TypeSOA
		soa := req.SOA
		soaNS, err := dnsmessage.NewName(soa.NS)
		if err != nil {
			return none, err
		}
		soaMBox, err := dnsmessage.NewName(soa.MBox)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.SOAResource{NS: soaNS, MBox: soaMBox, Serial: soa.Serial, Refresh: soa.Refresh, Retry: soa.Retry, Expire: soa.Expire}
	case "PTR":
		rType = dnsmessage.TypePTR
		ptr, err := dnsmessage.NewName(req.Data)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.PTRResource{PTR: ptr}
	case "MX":
		rType = dnsmessage.TypeMX
		mxName, err := dnsmessage.NewName(req.MX.MX)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.MXResource{Pref: req.MX.Pref, MX: mxName}
	case "AAAA":
		rType = dnsmessage.TypeAAAA
		ip := net.ParseIP(req.Data)
		if ip == nil {
			return none, errIPInvalid
		}
		var ipV6 [16]byte
		copy(ipV6[:], ip)
		rBody = &dnsmessage.AAAAResource{AAAA: ipV6}
	case "SRV":
		rType = dnsmessage.TypeSRV
		srv := req.SRV
		srvTarget, err := dnsmessage.NewName(srv.Target)
		if err != nil {
			return none, err
		}
		rBody = &dnsmessage.SRVResource{Priority: srv.Priority, Weight: srv.Weight, Port: srv.Port, Target: srvTarget}
	case "TXT":
		fallthrough
	case "OPT":
		fallthrough
	default:
		return none, errTypeNotSupport
	}

	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  rName,
			Type:  rType,
			Class: dnsmessage.ClassINET,
			TTL:   req.TTL,
		},
		Body: rBody,
	}, nil
}

func toRType(sType string) dnsmessage.Type {
	switch sType {
	case "A":
		return dnsmessage.TypeA
	case "NS":
		return dnsmessage.TypeNS
	case "CNAME":
		return dnsmessage.TypeCNAME
	case "SOA":
		return dnsmessage.TypeSOA
	case "PTR":
		return dnsmessage.TypePTR
	case "MX":
		return dnsmessage.TypeMX
	case "AAAA":
		return dnsmessage.TypeAAAA
	case "SRV":
		return dnsmessage.TypeSRV
	case "TXT":
		return dnsmessage.TypeTXT
	case "OPT":
		return dnsmessage.TypeOPT
	default:
		return 0
	}
}

func toResourceHeader(name string, sType string) (h dnsmessage.ResourceHeader, err error) {
	h.Name, err = dnsmessage.NewName(name)
	if err != nil {
		return
	}
	h.Type = toRType(sType)
	return
}
