package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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
	fmt.Printf("response is %#v\n", response)
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
	fmt.Printf("response is %#v\n", response)
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
	flag.StringVar(&HostRecord, "r", "dt", "指定主机记录")
}

//初始化
func init() {
	errFile, err := os.OpenFile("errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}

	Info = log.New(os.Stdout, "Info:", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(os.Stdout, "Warning:", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(os.Stderr, errFile), "Error:", log.Ldate|log.Ltime|log.Lshortfile)

	parseVar()
}

func main() {
	//获取命令行参数
	flag.Parse()

	var oldIp, oldHostRecord string

	// 每10分钟执行一次
	for range time.Tick(10 * time.Second) {

		//获取公网IP
		ip, _ := getPublicIp()

		Info.Println("ip:" + ip)

		if ip != oldIp || HostRecord != oldHostRecord {
			oldIp = ip
			oldHostRecord = HostRecord
			updateDomainRecord(HostRecord, ip)
		}
	}

}
