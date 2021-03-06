package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

var (
	Info       *log.Logger
	Warning    *log.Logger
	Error      *log.Logger
	HostRecord string
)

const (
	ACCESS_KEY_ID     = ""
	ACCESS_KEY_SECRET = ""
	DOMAIN_NAME       = "fuzuchang.com"
	JSON_IP           = "https://jsonip.com"
)

// 获取解析记录列表
func describeDomainRecords() {
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", ACCESS_KEY_ID, ACCESS_KEY_SECRET)

	request := alidns.CreateDescribeDomainRecordsRequest()
	request.DomainName = DOMAIN_NAME

	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		Error.Println(err)
	}
	Info.Println(response)
}

//修改解析记录
func updateDomainRecord(rr string, ip string) {
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", ACCESS_KEY_ID, ACCESS_KEY_SECRET)
	request := alidns.CreateUpdateDomainRecordRequest()
	request.RecordId = "20091037860762624"
	request.Type = "A"
	request.RR = rr
	request.Value = ip

	response, err := client.UpdateDomainRecord(request)
	if err != nil {
		Error.Println(err)
	}
	Info.Println(response)
}

type ClientIp struct {
	Ip    string `json:"ip"`
	GeoIp string `json:"geo-ip"`
}

//获取公网IP
func getPublicIp() (string, error) {
	res, err := http.Get(JSON_IP)
	if err != nil {
		Error.Println(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Error.Println(err)
	}
	clientIp := &ClientIp{}
	json.Unmarshal(body, &clientIp)
	return clientIp.Ip, err
}

//解析命令行参数
func parseVar() {
	flag.StringVar(&HostRecord, "r", "demo", "指定主机记录")
}

//初始化
func init() {
	errFile, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}

	Info = log.New(os.Stdout, "Info:", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(os.Stdout, "Warning:", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(os.Stderr, errFile), "Error:", log.Ldate|log.Ltime|log.Lshortfile)

	parseVar()
}

func updateDDNS() {
	//获取公网IP
	// ip, _ := getPublicIp()
	ip := parseIp()
	Info.Println("ip:" + ip)
	if ip != oldIp || HostRecord != oldHostRecord {
		oldIp = ip
		oldHostRecord = HostRecord
		updateDomainRecord(HostRecord, ip)
	}
}

var oldIp, oldHostRecord string

//获取IP
func parseIp() string {
	// searchUrl := "https://www.baidu.com/s?wd=ip"
	searchUrl := "http://www.net.cn/static/customercare/yourip.asp"

	res, err := http.Get(searchUrl)

	if err != nil {
		Error.Println(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		Error.Println("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		Error.Println(err)
	}

	// Find the review items
	ip := doc.Find("h2").Text()

	return ip
}

func main() {
	//获取命令行参数
	flag.Parse()
	updateDDNS()
	// 每10分钟执行一次
	for range time.Tick(5 * 60 * time.Second) {
		updateDDNS()
	}
}
