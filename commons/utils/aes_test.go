package utils

import (
	"fmt"
	"testing"
)

const testText string = "{\"userId\":1,\"amount\"}"
const testKey string = "123456789012345678901234"

var encryptOut string

func TestAesEncrypt(t *testing.T) {
	type args struct {
		origData []byte
		key      []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{name: "test 1", args: args{
			origData: []byte(testText),
			key:      []byte(testKey),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AesEncrypt(tt.args.origData, tt.args.key)
			if err != nil {
				t.Errorf("AesEncrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			encryptOut = got
		})
	}
}
func TestAesDecrypt(t *testing.T) {
	type args struct {
		crypted string
		key     []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "test_1", args: args{
			crypted: encryptOut,
			key:     []byte(testKey),
		}, want: testText},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AesDecrypt(tt.args.crypted, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("AesDecrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AesDecrypt() got = %v, want %v", got, tt.want)
			}
			fmt.Println(string(got))
		})
	}
}
