package main

import (
	// _ "net/http/pprof"

	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/haierspi/pt-gateway/utils/config"
	"github.com/haierspi/pt-gateway/utils/ptjson"
	"github.com/haierspi/pt-gateway/utils/rpc"
)

const (
	layoutInt = "20060102150405"
)

var (
	client     *rpc.Client
	listenPort string
	isDebug    bool
	signKey    = "signKey"
)

// Resp 响应
type Resp struct {
	ErrorCode int64
	ErrorMsg  string
}

func init() {
	var err error
	signKey = config.String("./config.cfg", "gateway", "signKey")
	listenPort = config.String("./config.cfg", "gateway", "listen")
	mqURL := config.String("./config.cfg", "mq", "url")
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	client, err = rpc.Dial(mqURL)
	if err != nil {
		log.Fatal("rpc Dial:", mqURL, err)
	}
	if client == nil {
		log.Fatal("rpc Dial: client is nil,", mqURL)
	}
	client.Timeout = config.Int64("./config.cfg", "gateway", "timeout")
	isDebug = config.Bool("./config.cfg", "gateway", "debug")
}

func main() {
	fmt.Println(listenPort)
	fmt.Println(http.ListenAndServe(listenPort, http.HandlerFunc(gateway)))
}

func gateway(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/gateway" || path == "/gateway/" {
		gatewayDefault(w, r)
		return
	}
	if strings.Index(path, "/gateway/b/") == 0 { // 场景微信支付
		gatewayBody(w, r, path[11:])
		return
	}
	if strings.Index(path, "/gateway/f/") == 0 { // 场景支付宝支付
		gatewayForm(w, r, path[11:])
		return
	}
	if strings.Index(path, "/gateway/r/") == 0 {
		gatewayRaw(w, r, path[11:])
		return
	}
	if strings.Index(path, "/gateway/u/") == 0 { // 场景如图片
		gatewayURL(w, r, path[11:])
		return
	}

}

// body字符串，成为bizContent的Body字段，用于post body回调，比如微信支付
//
// url: /gateway/b/m/examples_1.0_Examples.Echo
//
// body: 任意文本
//
// bizContent
//
//	{"Body":"任意文本"}
//
// 必须返回
//
//	{
//	   "Body": "<xml></xml>",
//	   "ContentType": "text/xml"
//	}
func gatewayBody(w http.ResponseWriter, r *http.Request, path string) {
	// 公共参数
	module, version, method, callBack, _ := _handPath(path)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	} else {
		r.Body.Close()
	}

	bizContentData := map[string]interface{}{
		"Body":     string(body),
		"ClientIP": getClientIP(r),
	}
	_callAPI(module, version, method, callBack, "", "", true, bizContentData, w, nil)
}

// bizContent在form中，用于post Form回调，比如支付宝支付
//
// url: /gateway/f/m/examples_1.0_Examples.Echo
//
// 表单: kv
//
// bizContent,kv都是字符串
//
//	{"key":"value"}
//
// 必须返回结构
//
//	{
//	   "Body": "<xml></xml>",
//	   "ContentType": "text/xml"
//	}
func gatewayForm(w http.ResponseWriter, r *http.Request, path string) {
	// 公共参数
	module, version, method, callBack, _ := _handPath(path)

	// bizContent
	var err error
	requestContentType := r.Header.Get("Content-Type")
	if strings.Index(requestContentType, "multipart/form-data") != -1 {
		err = r.ParseMultipartForm(32 << 20)
	} else {
		err = r.ParseForm()
	}

	formValues := r.Form
	bizContentData := map[string]interface{}{}
	for key, val := range formValues {
		bizContentData[key] = val[0]
	}
	bizContentData["ClientIP"] = getClientIP(r)
	_callAPI(module, version, method, callBack, "", "", true, bizContentData, w, err)
}

// body字符串是json，解析为bizContent
//
// url: /gateway/r/m/examples_1.0_Examples.Echo
//
// body: json object 字符串
//
// 返回 任意数据
func gatewayRaw(w http.ResponseWriter, r *http.Request, path string) {
	// 公共参数
	module, version, method, callBack, _ := _handPath(path)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	} else {
		r.Body.Close()
	}

	// bizContent
	var bizContentData map[string]interface{}
	err = ptjson.Unmarshal(body, &bizContentData)
	if bizContentData == nil {
		bizContentData = map[string]interface{}{}
	}

	bizContentData["ClientIP"] = getClientIP(r)
	_callAPI(module, version, method, callBack, "", "", false, bizContentData, w, err)
}

// body字符串是json，解析为bizContent
//
// url: /gateway/u/m/examples_1.0_Examples.Echo/b/{"Body":"hahaha"}
//
// body: json object 字符串
//
// 返回 任意数据
func gatewayURL(w http.ResponseWriter, r *http.Request, path string) {
	// 公共参数
	module, version, method, callBack, bizContent := _handPath(path)

	// bizContent
	var bizContentData map[string]interface{}
	err := ptjson.Unmarshal([]byte(bizContent), &bizContentData)
	if bizContentData == nil {
		bizContentData = map[string]interface{}{}
	}

	bizContentData["ClientIP"] = getClientIP(r)
	_callAPI(module, version, method, callBack, "", "", true, bizContentData, w, err)
}

func _handPath(path string) (module, version, method, callBack, bizContent string) {
	paths := strings.Split(path, "/")
	lenPaths := len(paths)
	params := map[string]string{}
	if lenPaths%2 == 0 {
		for i := 0; i < lenPaths/2; i++ {
			params[paths[i*2]] = paths[i*2+1]
		}
	}
	m := params["m"]
	methodInfos := strings.Split(m, "|")
	if len(methodInfos) == 3 {
		module = methodInfos[0]
		version = methodInfos[1]
		method = methodInfos[2]
	}
	callBack = params["c"]
	bizContent = params["b"]
	return
}

// 所有数据均在query中
//
// url: /gateway/
//
// 表单或者query:
//
// module:examples
//
// version:1.0
//
// method:Examples.Echo
//
// bizContent:{"Body":"hahaha"}
//
// 返回 任意数据
func gatewayDefault(w http.ResponseWriter, r *http.Request) {
	var err error
	requestContentType := r.Header.Get("Content-Type")
	if strings.Index(requestContentType, "multipart/form-data") != -1 {
		err = r.ParseMultipartForm(32 << 20)
	} else {
		err = r.ParseForm()
	}
	if err != nil {
		log.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	formValues := r.Form

	// 公共参数
	module := formValues.Get("module")
	version := formValues.Get("version")
	method := formValues.Get("method")
	sign := formValues.Get("sign")
	callBack := formValues.Get("callback")

	// bizContent
	var bizContentData map[string]interface{}
	err = ptjson.Unmarshal([]byte(formValues.Get("bizContent")), &bizContentData)
	if bizContentData == nil {
		bizContentData = map[string]interface{}{}
	}
	bizContentData["ClientIP"] = getClientIP(r)
	signMessage := ""
	if sign != "" {
		signMessage = verifySign(r)
	}

	_callAPI(module, version, method, callBack, sign, signMessage, false, bizContentData, w, err)
}

func _callAPI(module, version, method, callBack, sign, signMessage string, isBody bool, bizContent map[string]interface{}, w http.ResponseWriter, err error) {
	var reply = &[]byte{}  //存正确的返回
	var result = new(Resp) //存错误的返回
	var start = time.Now()

	defer func(errRelsult *Resp, rightResult *[]byte) {
		var r1 []byte
		if result.ErrorCode != 0 {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			r1, _ = ptjson.PrettyMarshal(errRelsult)
		} else if isBody {
			var bodyReply rpc.BodyReply
			err := ptjson.Unmarshal(*rightResult, &bodyReply)
			if err != nil {
				log.Println(err)
			}
			if bodyReply.ContentType == "" {
				bodyReply.ContentType = "text/plain"
			}
			w.Header().Set("Content-Type", bodyReply.ContentType+"; charset=UTF-8")
			r1 = bodyReply.Body
		} else {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			r1 = *rightResult
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if callBack != "" {
			w.Write([]byte(callBack + "(" + string(r1) + ")"))
		} else {
			w.Write(r1)
		}

		if isDebug {
			log.Println(time.Now().Sub(start), module, method, version, bizContent, " replay:", string(*reply))
		} else {
			log.Println(time.Now().Sub(start), module, method, version, bizContent)
		}
	}(result, reply)

	if err != nil {
		result.ErrorCode = 5000
		result.ErrorMsg = "form表单错误:" + err.Error()
		return
	}

	if strings.Contains(method, "WithSign") {
		result.ErrorCode = 5001
		result.ErrorMsg = fmt.Sprintf("请求方法错误:%s", method)
		return
	}
	if sign != "" {
		if signMessage != "" {
			result.ErrorCode = 5002
			result.ErrorMsg = "签名错误:" + signMessage
			return
		}
		method = method + "WithSign"
	}

	b, _ := ptjson.Marshal(bizContent)
	err = client.JSONCall(fmt.Sprintf("%s_%s", module, version), method, &b, reply)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			log.Println(module, method, version, bizContent, err.Error())
		}
		result.ErrorCode = 5003
		result.ErrorMsg = strings.Replace(err.Error(), "WithSign", "", -1)
	}
}

func getClientIP(req *http.Request) string {
	ip := req.Header.Get("X-Real-IP")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = req.RemoteAddr
			ips := strings.Split(ip, ":")
			if len(ips) > 0 {
				ip = ips[0]
			}
		} else {
			ips := strings.Split(ip, ",")
			if len(ips) > 0 {
				ip = ips[0]
			}
		}
	}
	return ip
}

func verifySign(req *http.Request) (message string) {
	req.ParseForm()
	var urls url.Values
	if req.Method == "POST" {
		urls = req.Form
	} else {
		urls = req.URL.Query()
	}

	sign := urls.Get("sign")
	timeStamp := urls.Get("timestamp")
	t0, _ := time.ParseInLocation(layoutInt, timeStamp, time.Local)
	subs := time.Now().Sub(t0).Minutes()
	if subs < -5 || subs > 5 {
		message = "请求已过期"
	} else {
		urls.Del("sign")
		data, _ := url.QueryUnescape(urls.Encode())
		data = fmt.Sprintf("%s&key=%s", data, signKey)
		h := md5.New()
		h.Write([]byte(data))
		result := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
		if sign != result {
			message = "签名失败,拿掉签名试试"
		}
	}
	return
}
