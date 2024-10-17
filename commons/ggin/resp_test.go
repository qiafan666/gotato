package ggin

import (
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gerr"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"testing"
)

type UpdateFriendsReq struct {
	OwnerUserID   string                  `json:"owner_user_id"`
	FriendUserIDs []string                `json:"friend_user_ids"`
	Remark        *wrapperspb.StringValue `json:"remark,omitempty"`
}

func TestName(t *testing.T) {
	resp := &ApiResponse{
		Code: 1234,
		Msg:  "test",
		Dlt:  "4567",
		Data: &UpdateFriendsReq{
			OwnerUserID:   "123456",
			FriendUserIDs: []string{"1", "2", "3"},
			Remark:        wrapperspb.String("1234567"),
		},
	}
	t.Logf(gcast.ToString(resp))
	var newResp ApiResponse
	err := gerr.Unmarshal(gcast.ToByte(resp), &newResp)
	if err != nil {
		return
	}
	t.Log(newResp)
	data, err := resp.MarshalJSON()
	if err != nil {
		panic(err)
	}
	t.Log(string(data))

	err = gerr.Unmarshal(data, &newResp)
	if err != nil {
		return
	}
	t.Log(newResp)

	var rReso ApiResponse
	rReso.Data = &UpdateFriendsReq{}

	if err = gerr.Unmarshal(data, &rReso); err != nil {
		panic(err)
	}

	t.Log(rReso)

	var updateFriendsReq UpdateFriendsReq
	if err = gerr.Unmarshal(gcast.ToByte(rReso.Data), &updateFriendsReq); err != nil {
		panic(err)
	}
	t.Log(updateFriendsReq)
}
