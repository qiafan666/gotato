package ggin

import (
	"github.com/qiafan666/gotato/commons/ggin/jsonutil"
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
	data, err := resp.MarshalJSON()
	if err != nil {
		panic(err)
	}
	t.Log(string(data))

	var rReso ApiResponse
	rReso.Data = &UpdateFriendsReq{}

	if err = jsonutil.JsonUnmarshal(data, &rReso); err != nil {
		panic(err)
	}

	t.Logf("%+v\n", rReso)

}
