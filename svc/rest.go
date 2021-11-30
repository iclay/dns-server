package svc

import (
	"encoding/json"
	"net/http"
)

type RestServer interface {
	Create() http.HandlerFunc
	Read() http.HandlerFunc
	Update() http.HandlerFunc
	Delete() http.HandlerFunc
}

type RestService struct {
	Dn *DNSService
}

// // Add a SRV record
// curl -X POST \
//   http://localhost/dns \
//   -H 'Content-Type: application/json' \
//   -d '{
// 	"Host": "_sip._tcp.example.com.",
// 	"TTL": 300,
// 	"Type": "SRV",
// 	"SRV": {
// 		"Priority": 0,
// 		"Weight": 5,
// 		"Port": 5060,
// 		"Target": "sipserver.example.com."
// 	}
// }'

// // Update an A record from 124.108.115.87 to 127.0.0.1
// curl -X PUT \
//   http://localhost/dns \
//   -H 'Content-Type: application/json' \
//   -d '{
// 	"Host": "example.com.",
// 	"TTL": 600,
// 	"Type": "A",
// 	"OldData": "124.108.115.87",
// 	"Data": "127.0.0.1"
// }'

// // Delete a record
// curl -X DELETE \
//   http://localhost/dns \
//   -H 'Content-Type: application/json' \
//   -d '{
// 	"Host": "example.com.",
// 	"Type": "A"
// }'

type request struct {
	Host    string
	TTL     uint32
	Type    string
	Data    string
	OldData string
	SOA     requestSOA
	OldSOA  requestSOA
	MX      requestMX
	OldMX   requestMX
	SRV     requestSRV
	OldSRV  requestSRV
}

type requestSOA struct {
	NS      string
	MBox    string
	Serial  uint32
	Refresh uint32
	Retry   uint32
	Expire  uint32
	MinTTL  uint32
}

type requestMX struct {
	Pref uint16
	MX   string
}

type requestSRV struct {
	Priority uint16
	Weight   uint16
	Port     uint16
	Target   string
}

type get struct {
	Host string
	TTL  uint32
	Type string
	Data string
}

func (s *RestService) Create(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resource, err := toResource(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.Dn.save(ntString(resource.Header.Name, resource.Header.Type), resource, nil)
	w.WriteHeader(http.StatusCreated)
}

func (s *RestService) Read(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(s.Dn.all())
}

func (s *RestService) Update(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	oldReq := request{Host: req.Host, Type: req.Type, Data: req.OldData}
	old, err := toResource(oldReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resource, err := toResource(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok := s.Dn.save(ntString(resource.Header.Name, resource.Header.Type), resource, &old)
	if ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "", http.StatusNotFound)
}

func (s *RestService) Delete(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok := false
	h, err := toResourceHeader(req.Host, req.Type)
	if err == nil {
		ok = s.Dn.remove(ntString(h.Name, h.Type), nil)
	}

	if ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "", http.StatusNotFound)
}
