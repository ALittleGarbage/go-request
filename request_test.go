package main

import (
	"fmt"
	"os"
	"request/request"
	"testing"
)

func Test_GetParam(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	resp, err := request.Get("https://baidu.com").Debug().
		Header(map[string]interface{}{
			"token": "123456789",
		}).
		Param(map[string]interface{}{
			"a": 1,
			"b": 2,
			"c": []int{1, 2, 3},
			"d": &User{
				Name: "大黄",
				Age:  18,
			},
		}).Sync2String()
	if err != nil {
		fmt.Printf("err:%s\n", err)
		return
	}

	fmt.Printf("resp:%s\n", resp)
}

func Test_PostJson(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	resp, err := request.Post("https://baidu.com").Debug().
		Header(map[string]interface{}{
			"token": "123456789",
		}).
		Json(User{
			Name: "大黄",
			Age:  18,
		}).Sync2String()
	if err != nil {
		fmt.Printf("err:%s\n", err)
		return
	}

	fmt.Printf("resp:%s\n", resp)
}

func Test_Multipart(t *testing.T) {
	var filepath string
	file, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	resp, err := request.Post("https://baidu.com").Debug().
		Header(map[string]interface{}{
			"token": "123456789",
		}).
		Multipart(&request.Multipart{
			Files: []request.File{{
				Filename:  "img.png",
				Fieldname: "img",
				Data:      file,
			}},
			Form: map[string]string{"id": "1"},
		}).Sync2String()
	if err != nil {
		fmt.Printf("err:%s\n", err)
		return
	}

	fmt.Printf("resp:%s\n", resp)
}
