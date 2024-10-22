package gzookeeper

import (
	"github.com/go-zookeeper/zk"
	"github.com/qiafan666/gotato/commons/gerr"
	"time"
)

type Config struct {
	ZkServers []string
	Scheme    string
	Username  string
	Password  string
	Timeout   time.Duration
}

func (s *ZkClient) RegisterConf2Registry(key string, conf []byte) error {
	path := s.getPath(key)

	exists, _, err := s.conn.Exists(path)
	if err != nil {
		return gerr.WrapMsg(err, "Exists failed", "path", path)
	}

	if exists {
		if err := s.conn.Delete(path, 0); err != nil {
			return gerr.WrapMsg(err, "Delete failed", "path", path)
		}
	}
	_, err = s.conn.Create(path, conf, 0, zk.WorldACL(zk.PermAll))
	if err != nil && err != zk.ErrNodeExists {
		return gerr.WrapMsg(err, "Create failed", "path", path)
	}
	return nil
}

func (s *ZkClient) GetConfFromRegistry(key string) ([]byte, error) {
	path := s.getPath(key)
	bytes, _, err := s.conn.Get(path)
	if err != nil {
		return nil, gerr.WrapMsg(err, "Get failed", "path", path)
	}
	return bytes, nil
}
