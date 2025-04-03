package xlsx

import (
	"github.com/qiafan666/gotato/commons/gcommon"
	"testing"
)

// 只支持int string bool float64 类型，不支持slice map struct等复杂类型
type User struct {
	UserID      string `column:"user_id"`
	Nickname    string `column:"nickname"`
	FaceURL     string `column:"face_url"`
	Birth       string `column:"birth"`
	Gender      string `column:"gender"`
	AreaCode    string `column:"area_code"`
	PhoneNumber string `column:"phone_number"`
	Email       string `column:"email"`
	Account     string `column:"account"`
	Password    string `column:"password"`
}

func (User) SheetName() string {
	return "user"
}

type User1 struct {
	UserID   string `column:"user_id"`
	Nickname string `column:"nickname"`
	FaceURL  string `column:"face_url"`
	Birth    string `column:"birth"`
	Gender   string `column:"gender"`
}

func (User1) SheetName() string {
	return "user1"
}

func TestReadXlsx(t *testing.T) {
	//formFile, err := c.FormFile("data")
	//if err != nil {
	//	return
	//}

	//file, err := formFile.Open()
	//if err != nil {
	//	return
	//}
	//defer file.Close()
	filePath := "./template.xlsx"
	file, err := gcommon.NewFileReader(filePath)
	if err != nil {
		t.Error(err)
	}
	//var file multipart.File
	var users []User
	var users1 []User1
	if err = Import(file, &users, &users1); err != nil {
		t.Error(err)
	}
	t.Log(users)
	t.Log(users1)
}
