package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	netUrl "net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	methodGet    = "GET"
	methodPost   = "POST"
	methodPut    = "PUT"
	methodDelete = "DELETE"

	contentLength = "Content-Length"

	contentType       = "Content-Type"
	contentTypeStream = "application/octet-stream"
	contentTypeJson   = "application/json"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

type param interface {
	Add(key string, value string)
}

type File struct {
	Filename  string
	Fieldname string
	Data      []byte
}

type Multipart struct {
	Files []File
	Form  map[string]string
}

type httpClient struct {
	url            *netUrl.URL
	method         string
	header         http.Header
	body           bytes.Buffer
	err            error
	timeout        time.Duration
	recursionCount int
	isDebug        bool
}

func DefReq(method string, url string, args ...interface{}) *httpClient {
	hc := &httpClient{
		url:            nil,
		method:         method,
		header:         make(http.Header),
		body:           bytes.Buffer{},
		err:            nil,
		timeout:        time.Second * 2,
		recursionCount: 1 << 4,
		isDebug:        false,
	}
	parse, err := netUrl.Parse(fmt.Sprintf(url, args...))
	if err != nil {
		hc.err = fmt.Errorf("DefReq() err:%s\n", err)
		return hc
	}
	hc.url = parse
	return hc
}

func Get(url string, args ...interface{}) *httpClient {
	return DefReq(methodGet, url, args...)
}

func Post(url string, args ...interface{}) *httpClient {
	return DefReq(methodPost, url, args...)
}

func Put(url string, args ...interface{}) *httpClient {
	return DefReq(methodPut, url, args...)
}

func Delete(url string, args ...interface{}) *httpClient {
	return DefReq(methodDelete, url, args...)
}

func (hc *httpClient) Debug() *httpClient {
	hc.isDebug = true
	return hc
}

func (hc *httpClient) Timeout(duration time.Duration) *httpClient {
	hc.timeout = duration
	return hc
}

func (hc *httpClient) Header(header interface{}) *httpClient {
	if hc.err != nil || header == nil {
		return hc
	}
	err := hc.conv(hc.header, "", header, hc.recursionCount)
	if err != nil {
		hc.err = fmt.Errorf("Header() err:%s ", err)
	}
	return hc
}

func (hc *httpClient) Param(param interface{}) *httpClient {
	if hc.err != nil || param == nil {
		return hc
	}
	params := make(netUrl.Values)
	err := hc.conv(params, "", param, hc.recursionCount)
	if err != nil {
		hc.err = fmt.Errorf("Param() err:%s ", err)
		return hc
	}
	hc.url.RawQuery = params.Encode()
	return hc
}

func (hc *httpClient) Stream(stream []byte) *httpClient {
	if hc.err != nil || stream == nil {
		return hc
	}
	hc.body.Reset()
	hc.body.Write(stream)
	hc.header.Add(contentType, contentTypeStream)
	hc.header.Add(contentLength, strconv.Itoa(hc.body.Len()))
	return hc
}

func (hc *httpClient) Json(str interface{}) *httpClient {
	if hc.err != nil || str == nil {
		return hc
	}
	jsonBytes, err := json.Marshal(str)
	if err != nil {
		hc.err = fmt.Errorf("Json() err:%s ", err)
		return hc
	}
	hc.body.Reset()
	hc.body.Write(jsonBytes)
	hc.header.Add(contentType, contentTypeJson)
	hc.header.Add(contentLength, strconv.Itoa(hc.body.Len()))
	return hc
}

func (hc *httpClient) Form(form interface{}) *httpClient {
	if hc.err != nil || form == nil {
		return hc
	}
	forms := make(netUrl.Values)
	err := hc.conv(forms, "", form, hc.recursionCount)
	if err != nil {
		hc.err = fmt.Errorf("Form() err:%s ", err)
		return hc
	}
	hc.body.Reset()
	hc.body.WriteString(forms.Encode())
	hc.header.Add(contentType, contentTypeForm)
	hc.header.Add(contentLength, strconv.Itoa(hc.body.Len()))
	return hc
}

func (hc *httpClient) Multipart(multi *Multipart) *httpClient {
	if hc.err != nil || multi == nil {
		return hc
	}
	hc.body.Reset()
	writer := multipart.NewWriter(&hc.body)
	for k, v := range multi.Form {
		e := writer.WriteField(k, v)
		if e != nil {
			hc.err = fmt.Errorf("Multipart() err:%s ", e)
			return hc
		}
	}
	for _, file := range multi.Files {
		formFile, err := writer.CreateFormFile(file.Fieldname, file.Filename)
		if err != nil {
			hc.err = fmt.Errorf("Multipart() err:%s ", err)
			return hc
		}
		_, err = formFile.Write(file.Data)
		if err != nil {
			hc.err = fmt.Errorf("Multipart() err:%s ", err)
			return hc
		}
	}
	err := writer.Close()
	if err != nil {
		hc.err = fmt.Errorf("Multipart() err:%s ", err)
		return hc
	}
	hc.header.Add(contentType, writer.FormDataContentType())
	hc.header.Add(contentLength, strconv.Itoa(hc.body.Len()))
	return hc
}

func (hc *httpClient) Sync() ([]byte, error) {
	if hc.err != nil {
		return nil, hc.err
	}
	req, err := http.NewRequest(hc.method, hc.url.String(), &hc.body)
	if err != nil {
		return nil, fmt.Errorf("Sync() err:%s ", err)
	}
	req.Header = hc.header
	hc.log()
	client := http.Client{Timeout: hc.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Sync() err:%s ", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Sync() err:%s ", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sync() StatusCode:%s ErrMsg:%s ", resp.Status, string(data))
	}
	return data, nil
}

func (hc *httpClient) Sync2String() (string, error) {
	data, err := hc.Sync()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (hc *httpClient) Sync2Struct(str interface{}) error {
	data, err := hc.Sync()
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, str)
	return err
}

func (hc *httpClient) Async(callback func([]byte), errCallback func(err error)) error {
	if hc.err != nil {
		return hc.err
	}
	req, err := http.NewRequest(hc.method, hc.url.String(), &hc.body)
	if err != nil {
		return fmt.Errorf("Async() err:%s ", err)
	}
	req.Header = hc.header
	hc.log()
	client := http.Client{Timeout: hc.timeout}
	go func() {
		resp, e := client.Do(req)
		if e != nil {
			errCallback(e)
			return
		}
		defer resp.Body.Close()
		data, e := io.ReadAll(resp.Body)
		if e != nil {
			errCallback(e)
			return
		}
		if resp.StatusCode != http.StatusOK {
			e = fmt.Errorf("Async() StatusCode:%s ErrMsg:%s ", resp.Status, string(data))
			errCallback(e)
		}
		callback(data)
	}()
	return nil
}

func (hc *httpClient) conv(param param, key string, value interface{}, count int) error {
	if count == 0 {
		return errors.New("传入的数据结构过于复杂，请简化数据结构")
	}
	count--

	typeOf := reflect.TypeOf(value)
	valueOf := reflect.ValueOf(value)

	switch {
	case typeOf.Kind() == reflect.String:
		if key == "" {
			return errors.New("不支持传入string类型")
		}
		param.Add(key, value.(string))
	case typeOf.Kind() == reflect.Slice && typeOf.Elem().Kind() == reflect.Uint8:
		if key == "" {
			return errors.New("不支持传入[]byte类型")
		}
		param.Add(key, string(value.([]byte)))
	case typeOf.Kind() == reflect.Array || typeOf.Kind() == reflect.Slice:
		if key == "" {
			return errors.New("不支持传入slice或者array类型")
		}
		for i := 0; i < valueOf.Len(); i++ {
			val := valueOf.Index(i).Interface()
			err := hc.conv(param, key+"["+strconv.Itoa(i)+"]", val, count)
			if err != nil {
				return err
			}
		}
	case typeOf.Kind() == reflect.Map:
		kind := typeOf.Key().Kind()
		if kind != reflect.String {
			return errors.New("仅支持传入的map类型的key值为string类型")
		}
		if key != "" {
			key = key + "."
		}
		for _, k := range valueOf.MapKeys() {
			v := valueOf.MapIndex(k).Interface()
			err := hc.conv(param, key+k.Interface().(string), v, count)
			if err != nil {
				return err
			}
		}
	case typeOf.Kind() == reflect.Struct:
		if key != "" {
			key = key + "."
		}
		for i := 0; i < typeOf.NumField(); i++ {
			field := typeOf.Field(i)
			k := field.Tag.Get("json")
			if k == "" {
				k = field.Name
			} else {
				k = strings.Split(k, ",")[0]
			}
			v := valueOf.Field(i).Interface()
			err := hc.conv(param, key+k, v, count)
			if err != nil {
				return err
			}
		}
	case typeOf.Kind() == reflect.Ptr:
		if value == nil {
			return nil
		}
		err := hc.conv(param, key, valueOf.Elem().Interface(), count)
		return err
	default:
		param.Add(key, fmt.Sprintf("%v", value))
	}
	return nil
}

func (hc *httpClient) log() {
	if !hc.isDebug {
		return
	}

	fmt.Printf("+ [%s] %s\n", hc.method, hc.url.String())
	fmt.Printf("+ Header:\n")
	for k := range hc.header {
		fmt.Printf("+--- %s : %s\n", k, hc.header.Get(k))
	}
	fmt.Printf("+ Body:\n")
	if hc.body.Len() != 0 {
		fmt.Printf("+--- %s\n", strings.Replace(hc.body.String(), "\n", "\n+--- ", -1))
	}
}
