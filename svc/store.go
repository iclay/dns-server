package svc

import (
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

const (
	//域名解析后的映射关系保存至本地文件，采用gob编码，每次重启会重新加载文件值至内存
	storeName   string = "store"
	storeBkName string = "store_bk"
)

func init() {
	gob.Register(&dnsmessage.AResource{})
	gob.Register(&dnsmessage.NSResource{})
	gob.Register(&dnsmessage.CNAMEResource{})
	gob.Register(&dnsmessage.SOAResource{})
	gob.Register(&dnsmessage.PTRResource{})
	gob.Register(&dnsmessage.MXResource{})
	gob.Register(&dnsmessage.AAAAResource{})
	gob.Register(&dnsmessage.SRVResource{})
	gob.Register(&dnsmessage.TXTResource{})
	gob.Register(&dnsmessage.PTRResource{})
}

type store struct {
	sync.RWMutex
	data      map[string]entry
	rwDirPath string
}

type entry struct {
	Resources []dnsmessage.Resource
	TTL       uint32
	Created   int64
}

func (s *store) get(key string) ([]dnsmessage.Resource, bool) {
	s.RLock()
	e, ok := s.data[key]
	s.RUnlock()
	now := time.Now().Unix()
	if e.TTL > 1 && (e.Created+int64(e.TTL) < now) { //判断dns缓存是否超时，如果超时直接删除
		s.remove(key, nil)
		return nil, false
	}
	return e.Resources, ok
}

func (s *store) set(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool {
	changed := false
	s.Lock()
	if _, ok := s.data[key]; ok {
		if old != nil {
			for i, rec := range s.data[key].Resources {
				if rString(rec) == rString(*old) {
					s.data[key].Resources[i] = resource
					changed = true
					break
				}
			}
		} else {
			e := s.data[key]
			e.Resources = append(e.Resources, resource)
			s.data[key] = e
			changed = true
		}
	} else {
		e := entry{
			Resources: []dnsmessage.Resource{resource},
			TTL:       resource.Header.TTL,
			Created:   time.Now().Unix(),
		}
		s.data[key] = e
		changed = true
	}
	s.Unlock()

	return changed
}

func (s *store) override(key string, resources []dnsmessage.Resource) {
	s.Lock()
	e := entry{
		Resources: resources,
		Created:   time.Now().Unix(),
	}
	if len(resources) > 0 {
		e.TTL = resources[0].Header.TTL
	}
	s.data[key] = e
	s.Unlock()
}

func (s *store) remove(key string, r *dnsmessage.Resource) bool {
	ok := false
	s.Lock()
	if r == nil {
		_, ok = s.data[key]
		delete(s.data, key)
	} else {
		if _, ok = s.data[key]; ok {
			for i, rec := range s.data[key].Resources {
				if rString(rec) == rString(*r) {
					e := s.data[key]
					copy(e.Resources[i:], e.Resources[i+1:])
					var blank dnsmessage.Resource
					e.Resources[len(e.Resources)-1] = blank
					e.Resources = e.Resources[:len(e.Resources)-1]
					s.data[key] = e
					ok = true
					break
				}
			}
		}
	}
	s.Unlock()
	return ok
}

func (s *store) save() {
	bk, err := os.OpenFile(filepath.Join(s.rwDirPath, storeBkName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("err open store bak file %v", err)
		return
	}
	defer bk.Close()

	dst, err := os.OpenFile(filepath.Join(s.rwDirPath, storeName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("err open store file %v", err)
		return
	}
	defer dst.Close()
	_, err = io.Copy(bk, dst)
	if err != nil {
		log.Errorf("err copy store file %v", err)
		return
	}

	enc := gob.NewEncoder(dst)
	book := s.clone()
	if err = enc.Encode(book); err != nil {
		//log.Fatal(err)
		log.Error(err)
	}
}

func (s *store) load() {
	fReader, err := os.Open(filepath.Join(s.rwDirPath, storeName))
	if err != nil {
		log.Errorf("err load store file %v maybe first start,please ignore", err)
		return
	}
	defer fReader.Close()

	dec := gob.NewDecoder(fReader)

	s.Lock()
	defer s.Unlock()

	if err = dec.Decode(&s.data); err != nil {
		log.Fatalf("err decode store file %v", err)
	}
}

func (s *store) clone() map[string]entry {
	cp := make(map[string]entry)
	s.RLock()
	for k, v := range s.data {
		cp[k] = v
	}
	s.RUnlock()
	return cp
}
