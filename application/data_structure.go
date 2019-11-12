package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack"
)

// Notice is a struck for unmarshalling msgpack
type Notice struct {
	Domain string `msgpack:"domain"`
	IP     uint32 `msgpack:"ip"`
}

func newNotice(received []byte) (data *Notice, err error) {
	err = msgpack.Unmarshal(received, &data)
	//	if err != nil {
	//		log.Fatalf("error: %v data: %v", err, received)
	//	}
	return data, err
}

func (data *Notice) String() string           { return fmt.Sprintf("%s : %d", data.Domain, data.IP) }
func (data *Notice) marshal() ([]byte, error) { return msgpack.Marshal(data) }

type row struct {
	ip  uint32
	lat int64
}

func newRow(ip uint32) *row   { return &row{ip, time.Now().Unix()} }
func (r *row) String() string { return intToIP(r.ip) + " " + time.Unix(r.lat, 0).String() }

func intToIP(ip uint32) string {
	result := make(net.IP, 4)
	result[0] = byte(ip >> 24)
	result[1] = byte(ip >> 16)
	result[2] = byte(ip >> 8)
	result[3] = byte(ip)
	return result.String()
}

type timeMap struct {
	data map[string]*row
	ttl  int64
	sync.Mutex
}

func newTimeMap(ttl, cleanPeriod int64) (tm *timeMap) {
	tm = new(timeMap)
	tm.ttl = ttl
	tm.data = make(map[string]*row)
	go func() {
		for now := range time.Tick(time.Duration(cleanPeriod) * time.Second) {
			deadline := now.Unix() - tm.ttl
			tm.Lock()
			for dom, item := range tm.data {
				if item.lat < deadline {
					delete(tm.data, dom)
				}
			}
			tm.Unlock()
		}
	}()
	return tm
}

func (tm *timeMap) add(new *Notice) {
	tm.Lock()
	defer tm.Unlock()
	tm.data[new.Domain] = newRow(new.IP)
}

func (tm *timeMap) len() int { return len(tm.data) }

func (tm *timeMap) String() (tms string) {
	tm.Lock()
	defer tm.Unlock()
	for dom, item := range tm.data {
		tms += dom + " " + item.String() + "\n"
	}
	tms += fmt.Sprintf("TOTAL: %d items", tm.len())
	return tms
}
