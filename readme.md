## 1.介绍

* 原生HttpClient，没有使用任何第三方包
* 万能HttpClient，什么请求都能发送
* 简易HttpCleint，采用链式调用

## 2.快速入门

|             | 方法                                          |
| ----------- | --------------------------------------------- |
| 起始函数    | DefReq()、Get()、Post()、Put()、Delete()      |
| 设置请求头  | Header()                                      |
| 设置url参数 | Param()                                       |
| 设置请求体  | Stream()、Json()、Form()、Multipart()         |
| 终结方法    | Sync()、Sync2String()、Sync2Struct()、Async() |

## 3.示例

1. GetParam：

```go
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
```

结果：

```plaintext
+ [GET] https://baidu.com?a=1&b=2&c%5B0%5D=1&c%5B1%5D=2&c%5B2%5D=3&d.Age=18&d.Name=%E5%A4%A7%E9%BB%84
+ Header:
+--- Token : 123456789
+ Body:
resp:<!DOCTYPE html>.......
```

2. PostJson:

```go
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
```

结果：

```plaintext
+ [POST] https://baidu.com
+ Header:
+--- Content-Type : application/json
+--- Content-Length : 26
+--- Token : 123456789
+ Body:
+--- {"Name":"大黄","Age":18}
resp:<!DOCTYPE html>....
```

3. multipart

```go
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
```

结果：
```plaintext
+ [POST] https://baidu.com
+ Header:
+--- Token : 123456789
+--- Content-Type : multipart/form-data; boundary=d5b2228f9f5e9384797ae622df1df78ec0fa6994a0a09308aacb28edf667
+--- Content-Length : 1759
+ Body:
+--- --d5b2228f9f5e9384797ae622df1df78ec0fa6994a0a09308aacb28edf667
+--- Content-Disposition: form-data; name="id"
+--- 
+--- 1
+--- --d5b2228f9f5e9384797ae622df1df78ec0fa6994a0a09308aacb28edf667
+--- Content-Disposition: form-data; name="img"; filename="img.png"
+--- Content-Type: application/octet-stream
+--- 
+--- �PNG
+--- 
  ���CH��`�&}M�f�P0o��
+--- ?*��#��
+--- ?*��#��
...
+--- ?*��#��[Tњs<�    IEND�B`�
+--- --d5b2228f9f5e9384797ae622df1df78ec0fa6994a0a09308aacb28edf667--
+--- 
resp:<!DOCTYPE html>...
```
