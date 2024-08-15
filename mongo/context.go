package mongo

import (
	"container/heap"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/qiafan666/gotato/commons/glog"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"net"
	"sync"
	"time"
)

// session
type session struct {
	*mgo.Session
	ref   int
	index int
}

// session heap
type sessionHeap []*session

// Len 返回堆的长度
func (h sessionHeap) Len() int {
	return len(h)
}

func (h sessionHeap) Less(i, j int) bool {
	return h[i].ref < h[j].ref
}

func (h sessionHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *sessionHeap) Push(s interface{}) {
	s.(*session).index = len(*h)
	*h = append(*h, s.(*session))
}

func (h *sessionHeap) Pop() interface{} {
	l := len(*h)
	s := (*h)[l-1]
	s.index = -1
	*h = (*h)[:l-1]
	return s
}

type DialContext struct {
	sync.Mutex
	sessions sessionHeap
}

// goroutine safe
func dial(url string, sessionNum int) (*DialContext, error) {
	c, err := dialWithTimeout(url, sessionNum, 10*time.Second, 0)
	return c, err
}

func dialTLS(url string, sessionNum int, caFilePath string) (*DialContext, error) {
	c, err := dialWithInfo(url, sessionNum, caFilePath)
	return c, err
}

// goroutine safe
func dialWithTimeout(url string, sessionNum int, dialTimeout time.Duration, timeout time.Duration) (*DialContext, error) {
	if sessionNum <= 0 {
		sessionNum = 100
		glog.Slog.WarnF(nil, "sessionNum should be greater than 0, use default value: %v", sessionNum)
	}

	s, err := mgo.DialWithTimeout(url, dialTimeout)
	if err != nil {
		return nil, err
	}
	s.SetSyncTimeout(timeout)
	s.SetSocketTimeout(timeout)

	c := new(DialContext)

	// sessions
	c.sessions = make(sessionHeap, sessionNum)
	c.sessions[0] = &session{s, 0, 0}
	for i := 1; i < sessionNum; i++ {
		c.sessions[i] = &session{s.New(), 0, i}
	}
	heap.Init(&c.sessions)

	return c, nil
}

// goroutine safe
func dialWithInfo(url string, sessionNum int, caFilePath string) (*DialContext, error) {
	if sessionNum <= 0 {
		sessionNum = 100
	}

	dialInfo, err := mgo.ParseURL(url)
	if err != nil {
		fmt.Println(err.Error())
	}
	tlsConfig, err := getCustomTLSConfig(caFilePath)
	if err != nil {
		fmt.Println(err.Error())
	}
	// update dialserver with tls Dial
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		if err != nil {
			fmt.Println(err.Error())
		}
		return conn, err
	}
	if err != nil {
		fmt.Println(err.Error())
	}
	s, err := mgo.DialWithInfo(dialInfo)

	c := new(DialContext)

	// sessions
	c.sessions = make(sessionHeap, sessionNum)
	c.sessions[0] = &session{s, 0, 0}
	for i := 1; i < sessionNum; i++ {
		c.sessions[i] = &session{s.New(), 0, i}
	}
	heap.Init(&c.sessions)

	return c, nil
}

// Close goroutine safe
func (c *DialContext) Close() {
	c.Lock()
	for _, s := range c.sessions {
		s.Close()
		if s.ref != 0 {
			glog.Slog.WarnF(nil, "session ref not zero: %v", s.ref)
		}
	}
	c.Unlock()
}

// Ref goroutine safe
func (c *DialContext) Ref() *session {
	c.Lock()
	s := c.sessions[0]
	if s.ref == 0 {
		s.Refresh()
	}
	s.ref++
	heap.Fix(&c.sessions, 0)
	c.Unlock()

	return s
}

// UnRef goroutine safe
func (c *DialContext) UnRef(s *session) {
	c.Lock()
	s.ref--
	heap.Fix(&c.sessions, s.index)
	c.Unlock()
}

func getCustomTLSConfig(caFile string) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	certs, err := ioutil.ReadFile(caFile)

	if err != nil {
		return tlsConfig, err
	}

	tlsConfig.RootCAs = x509.NewCertPool()
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs)

	if !ok {
		return tlsConfig, errors.New("failed parsing pem file")
	}

	return tlsConfig, nil
}
