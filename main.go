package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

var COOKIE []*http.Cookie
var TOKEN, USER, PASS, URL, IPFILTER, PATTERN string

func HttpGetURL(url string) []byte {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	request.AddCookie(COOKIE[0])
	request.SetBasicAuth(USER, PASS)
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("请求错误,确认uTorrent是否运行,并开启网页界面")
		os.Exit(3)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return body
}

func GetToken() (string, []*http.Cookie) {
	newurl := URL + "token.html"
	request, err := http.NewRequest("GET", newurl, nil)
	if err != nil {
		panic(err)
	}
	request.SetBasicAuth(USER, PASS)
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("请求错误,请确认URL是否正确。")
		os.Exit(2)
	}
	defer res.Body.Close()
	cookie := res.Cookies()
	if res.StatusCode > 200 {
		fmt.Println("无法获取token信息,请确认帐号密码是否正确。")
		os.Exit(1)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	pattern, _ := regexp.Compile(`<html><div id='token' .*>(.*?)</div></html>`)
	result := pattern.FindStringSubmatch(string(body))
	return result[1], cookie
}

func ParseTorrents(data []byte) []string {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	v := result["torrents"].([]interface{})
	var s = []string{}
	for _, i := range v {
		a := i.([]interface{})
		s = append(s, a[0].(string))
	}
	return s
}

func GetHash() []string {
	newurl := URL + "?token=" + TOKEN + "&list=1"
	// fmt.Println(newurl)
	data := HttpGetURL(newurl)
	s := ParseTorrents(data)
	return s
}

func ParsePeers(data []byte, s *[][]string) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	v := result["peers"].([]interface{})
	for idx, i := range v {
		if idx == 1 {
			a := i.([]interface{})
			for _, j := range a {
				var si = []string{}
				ss := j.([]interface{})
				s1 := ss[1].(string)
				// s4 := fmt.Sprintf("%d", int(ss[4].(float64)))
				s5 := ss[5].(string)
				si = append(si, s1, s5)
				*s = append(*s, si)
			}
		}
	}
}

func GetPeers(hash string, s *[][]string) {
	newurl := URL + "?token=" + TOKEN + "&action=getpeers&hash=" + hash
	// fmt.Println(newurl)
	data := HttpGetURL(newurl)
	ParsePeers(data, s)
}

func GetAllPeers(hash []string) [][]string {
	var s = [][]string{}
	for _, j := range hash {
		// fmt.Println(j)
		GetPeers(j, &s)
	}
	return s
}

func ReloadUT() {
	newurl := URL + "?token=" + TOKEN + "&action=setsetting&s=ipfilter.enable&v=1"
	HttpGetURL(newurl)
	// fmt.Println(data)
}

func WriteIpfilter(s []string) {
	file, err := os.OpenFile(IPFILTER, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	write := bufio.NewWriter(file)
	for _, i := range s {
		_, err = write.WriteString(i + "\n")
		if err != nil {
			panic(err)
		}
		// fmt.Println("写入信息: ", i)
	}
	err = write.Flush()
	if err != nil {
		panic(err)
	}
	ReloadUT()
}

func TruncateFile() {
	file, err := os.OpenFile(IPFILTER, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.Truncate(0)
}

func Block(s *[][]string) {
	var block_list []string
	currentTime := time.Now()
	fmt.Println(currentTime.Format("2006-01-02 15:04:05"), "屏蔽信息:")
	for _, i := range *s {
		s := i[1]
		ok, err := regexp.MatchString(PATTERN, s)
		if err != nil {
			fmt.Println("regexp err: ", err)
		}
		if ok {
			block_list = append(block_list, i[0])
			fmt.Printf("    IP: %s, Client: %s\n", i[0], strings.Replace(s, "\r", "", -1))
		}
	}

	keys := make(map[string]bool)
	list := []string{}
	for _, s := range block_list {
		if _, value := keys[s]; !value {
			keys[s] = true
			list = append(list, s)
		}
	}

	if len(list) != 0 {
		WriteIpfilter(list)
	}
}

func run() {
	ticker := time.NewTicker(2*time.Hour + 10*time.Second)
	go func() {
		for s := range ticker.C {
			fmt.Println(s.Format("2006-01-02 15:04"), " 运行两小时,清空IP列表")
			TruncateFile()
			ReloadUT()
		}
	}()
	for {
		h := GetHash()
		s := GetAllPeers(h)
		Block(&s)
		time.Sleep(30 * time.Second)
	}
}

func cmdParse() map[string]string {
	flagset := pflag.NewFlagSet("parse", pflag.ExitOnError)
	var uurl = flagset.String("url", "http://127.0.0.1:1000/gui/", "uTorrent host default http://127.0.0.1:1000/gui/")
	var uuser = flagset.String("user", "", "uTorrent username")
	var upass = flagset.String("pass", "", "uTorrent password")
	var regex = flagset.String("key", "", "自定议额外要屏蔽的客户端关键字,正则表达式。默认屏蔽、Xunlei、QQDownload、../torrent、aria2")
	var file = flagset.String("path", "", "指定uTorrent所在目录,将本程序放到uTorrent程序所在目录下可不配置此参数,可自行指定位置，例如'D:\\utorrent'。")

	flagset.Parse(os.Args[1:])

	args := map[string]string{
		"user":  "",
		"pass":  "",
		"url":   "http://127.0.0.1:1000/gui/",
		"regex": "",
		"file":  "",
	}
	args["user"] = *uuser
	args["pass"] = *upass
	args["url"] = *uurl
	args["regex"] = *regex
	args["file"] = *file
	return args
}

func main() {
	m := cmdParse()
	USER = m["user"]
	PASS = m["pass"]
	URL = m["url"]
	IPFILTER = m["file"]
	regex := m["regex"]
	if len(USER) == 0 || len(PASS) == 0 {
		fmt.Println("帐号密码不能为空")
		return
	}

	filename := "\\ipfilter.dat"
	if len(IPFILTER) == 0 {
		fpath, err := filepath.Abs(os.Args[0])
		if err != nil {
			fpath = "."
		}
		basepath := filepath.Dir(fpath)
		IPFILTER = basepath + filename
	} else {
		IPFILTER = IPFILTER + filename
	}

	if len(regex) == 0 {
		PATTERN = `(?i)(-XL0012-|Xunlei|QQDownload|..\/torrent|aria2)`
	} else {
		PATTERN = `(?i)(-XL0012-|Xunlei|QQDownload|..\/torrent|aria2)|` + regex
	}

	TOKEN, COOKIE = GetToken()

	go run()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	func() {
		for sig := range c {
			fmt.Printf("\nCaptured %v. Exiting...\n", sig)
			TruncateFile()
			os.Exit(0)
		}
	}()
}
