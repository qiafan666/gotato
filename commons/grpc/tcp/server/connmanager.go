package server

import (
	"context"
	"errors"
	"github.com/cloudwego/netpoll"
	"github.com/qiafan666/gotato/commons/gcache"
	"github.com/qiafan666/gotato/commons/gface"
)

type ConnManager struct {
	connRequests *gcache.ShardLockMap[string, *Request] // connId -> requests
	reqKeys      *gcache.ShardLockMap[string, string]   // reqKey -> connId
	logger       gface.Logger
}

func NewConnManager(logger gface.Logger) *ConnManager {
	m := &ConnManager{
		connRequests: gcache.NewShardLockMap[*Request](),
		reqKeys:      gcache.NewShardLockMap[string](),
		logger:       logger,
	}
	return m
}

// NewConn 建立连接时调用
func (cm *ConnManager) NewConn(
	ctx context.Context,
	connId string,
	conn netpoll.Connection,
) context.Context {
	if cm.connRequests.Has(connId) {
		return ctx
	}

	cm.logger.DebugF(nil, "ConnManage.NewConn, connId=%+v", connId)

	req := NewRequest(ctx, conn)
	cm.connRequests.Set(connId, req)

	ctx = context.WithValue(ctx, "connId", connId)
	return ctx
}

func (cm *ConnManager) CloseConn(ctx context.Context) {
	//根据connId 关闭连接,释放资源
	connId := cm.getConnIdFromCtx(ctx)
	if connId == "" {
		return
	}

	cm.logger.DebugF(nil, "ConnManage.CloseConn, connId=%+v", connId)

	cm.connRequests.RemoveCb(connId, func(key string, v *Request, exists bool) bool {
		if exists {
			return false
		}
		cm.logger.DebugF(nil, "ConnManage.CloseConn, remove connRequests, connId=%+v", connId)
		v.Close()

		keys := v.Keys()
		for _, reqKey := range keys {
			cm.reqKeys.Remove(reqKey)
			cm.logger.DebugF(nil, "ConnManage.CloseConn, remove reqKeys, connId=%+v, reqKey=%+v", connId, reqKey)
		}

		return true
	})
}

// NewRequest 新请求, 保存reqKey与connId的关系
func (cm *ConnManager) NewRequest(ctx context.Context, reqKey string) {
	connId := cm.getConnIdFromCtx(ctx)
	if connId == "" {
		return
	}

	cm.logger.DebugF(nil, "ConnManage.NewRequest, connId=%+v, reqKey=%+v", connId, reqKey)

	req, ok := cm.connRequests.Get(connId)
	if !ok || req == nil {
		return
	}
	cm.reqKeys.Set(reqKey, connId)
	req.NewKey(reqKey)
}

func (cm *ConnManager) GetRequest(reqKey string) (*Request, error) {
	connId, ok := cm.reqKeys.Get(reqKey)
	if !ok {
		return nil, errors.New("connId not found")
	}
	cm.logger.DebugF(nil, "ConnManage.GetRequest, connId=%+v, reqKey=%+v", connId, reqKey)
	cm.reqKeys.Remove(reqKey)

	req, ok := cm.connRequests.Get(connId)
	if !ok || req == nil {
		return nil, errors.New("request not found")
	}
	req.RemoveKey(reqKey)

	return req, nil
}

func (cm *ConnManager) getConnIdFromCtx(ctx context.Context) string {
	v := ctx.Value("connId")
	if v == nil {
		return ""
	}
	connId, ok := v.(string)
	if !ok {
		return ""
	}
	return connId
}
