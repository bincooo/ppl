package core

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bincooo/emit.io"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/iocgo/sdk/env"
	"github.com/iocgo/sdk/inited"
	"ppl/core/logger"
)

type Spider struct {
	Label    string                 `mapstructure:"label"`
	Targets  []string               `mapstructure:"target"`
	Method   string                 `mapstructure:"method"`
	Interval int                    `mapstructure:"interval"`
	Headers  map[string]string      `mapstructure:"headers"`
	Body     map[string]interface{} `mapstructure:"body"`
	Match    struct {
		Address string `mapStructure:"address"`
		Port    string `mapStructure:"port"`
	} `mapStructure:"match"`
}

type Elem struct {
	T string `json:"t"` //类型 http / https / socks5
	A string `json:"a"` //代理匿名度, 透明：显示真实IP, 普匿：显示假的IP, 高匿：无代理IP特征
	I string `json:"i"` //IP提供商

	Addr string `json:"addr"`
	Port int    `json:"port"`

	Country  string `json:"country"`  //代理国家
	Province string `json:"province"` //代理省份
	City     string `json:"city"`     //代理城市
	Source   string `json:"source"`   //代理源
}

func (elem Elem) String() string {
	return fmt.Sprintf(`Elem( T="%s", Addr="%s", Port=%d )`, elem.T, elem.Addr, elem.Port)
}

var (
	HTTPClient *emit.Session

	ElemMap map[string]Elem
)

func init() {
	inited.AddInitialized(func(args ...interface{}) {
		environ := args[0].(*env.Environment)
		proxied := environ.GetString("server.proxied")
		{
			session, err := emit.NewSession(proxied, warp("127.0.0.1"),
				emit.IdleConnTimeoutHelper(60*time.Second),
				emit.MaxIdleConnectsHelper(10),
				emit.Ja3Helper(emit.Echo{
					RandomTLSExtension: true,
					HelloID:            profiles.Chrome_124,
				}, 60),
				emit.TLSConfigHelper(&tls.Config{
					InsecureSkipVerify: true,
				}),
			)
			if err != nil {
				logger.Fatal(err)
			}
			HTTPClient = session
		}

		ElemMap = make(map[string]Elem)
		if fileExists("data.json") {
			chunk, err := os.ReadFile("data.json")
			if err != nil {
				logger.Fatal(err)
			}
			err = json.Unmarshal(chunk, &ElemMap)
			if err != nil {
				logger.Fatal(err)
			}
		}
	})
}

func warp(i ...string) emit.Wip {
	return func() []string {
		return i
	}
}

func Run(env *env.Environment) {
	time.Sleep(5 * time.Second)
	var slice []Spider
	{
		err := env.UnmarshalKey("spider", &slice)
		if err != nil {
			logger.Error(err)
			return
		}
	}

	count := len(slice)
	if count == 0 {
		logger.Fatal("nothing to crawl !!!")
		return
	}

	// 自检池子
	go verifyMap()

	ticker := time.NewTicker(60 * time.Second)
	for range ticker.C {
		var wg sync.WaitGroup
		for _, spi := range slice {
			wg.Add(1)
			go run(spi, func() { wg.Done() })
		}
		wg.Wait()
	}
}

func run(spi Spider, executor func()) {
	defer executor()
	for _, u := range spi.Targets {
		response, err := emit.ClientBuilder(HTTPClient).
			Ja3().
			URL(u).
			Method(spi.Method).
			Headers(spi.Headers).
			Body(spi.Body).
			DoS(http.StatusOK)
		if err != nil {
			logger.Errorf("[%s] : %v", spi.Label, err)
			continue
		}
		defer response.Body.Close()
		value := emit.TextResponse(response)

		address := regexp.MustCompile(spi.Match.Address).FindAllStringSubmatch(value, -1)
		port := regexp.MustCompile(spi.Match.Port).FindAllStringSubmatch(value, -1)
		if len(address) == 0 {
			continue
		}

		for index, addr := range address {
			var (
				i string
				p int
			)
			i, err = url.QueryUnescape(addr[1])
			if err != nil {
				logger.Error(err)
				continue
			}

			var q string
			q, err = url.QueryUnescape(port[index][1])
			if err != nil {
				logger.Error(err)
				continue
			}
			p, err = strconv.Atoi(q)
			if err != nil {
				logger.Error(err)
				continue
			}

			ok, elem := verify(i, p)
			if !ok {
				logger.Debugf("[%s:%d] : verify failed", i, p)
				// delete(ElemMap, Elem{Source: spi.Label, Addr: i, Port: p}.String())
				continue
			}

			meta(&elem)
			anonymity(&elem)
			elem.Source = spi.Label
			logger.Infof("verify success [%s] : %s", spi.Label, elem.String())
			ElemMap[elem.String()] = elem
		}

		time.Sleep(time.Duration(spi.Interval) * time.Second)
	}
}

func anonymity(elem *Elem) {
	host := "http://httpbin.org/get"
	proxy := ""
	if elem.T == "SOCKET5" {
		proxy = "socks5://" + elem.Addr + ":" + strconv.Itoa(elem.Port)
	} else {
		proxy = "http://" + elem.Addr + ":" + strconv.Itoa(elem.Port)
	}

	retry := 3
label:
	retry--

	pu, err := url.Parse(proxy)
	if err != nil {
		if retry > 0 {
			goto label
		}
		logger.Debugf("[anonymity] : %v", err)
		return
	}

	client := http.Client{Timeout: 15 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(pu),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	request, err := http.NewRequest("GET", host, nil)
	if err != nil {
		logger.Debugf("[anonymity] : %v", err)
		return
	}

	request.Header.Add("Proxy-Connection", "keep-alive")
	res, err := client.Do(request)
	if err != nil {
		if retry > 0 {
			goto label
		}
		logger.Debugf("[anonymity] : %v", err)
		return
	}

	defer res.Body.Close()
	dataBytes, _ := io.ReadAll(res.Body)
	result := string(dataBytes)
	if !strings.Contains(result, `"url": "http://httpbin.org/`) {
		if retry > 0 {
			goto label
		}
		return
	}

	origin := regexp.MustCompile("(\\d+?\\.\\d+?.\\d+?\\.\\d+?,.+\\d+?\\.\\d+?.\\d+?\\.\\d+?)").FindAllStringSubmatch(result, -1)
	if len(origin) != 0 {
		elem.A = "透明"
		return
	}
	if strings.Contains(result, "keep-alive") {
		elem.A = "普匿"
		return
	}
	elem.A = "高匿"
}

func meta(elem *Elem) {
	client := http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	retry := 3
label:
	retry--
	res, err := client.Get("https://searchplugin.csdn.net/api/v1/ip/get?ip=" + elem.Addr)
	if err != nil {
		if retry > 0 {
			goto label
		}
		logger.Debugf("[meta] : %v", err)
		return
	}

	defer res.Body.Close()
	dataBytes, _ := io.ReadAll(res.Body)
	result := string(dataBytes)
	address := regexp.MustCompile("\"address\":\"(.+?)\",").FindAllStringSubmatch(result, -1)
	if len(address) != 0 {
		slice := filter(strings.Split(address[0][1], " "))
		count := len(slice)
		elem.I = strings.Split(slice[count-1], "/")[0]
		for i := range slice {
			if i == count-1 {
				break
			}
			switch i {
			case 0:
				elem.Country = slice[0]
			case 1:
				elem.Province = slice[1]
			case 2:
				elem.City = slice[2]
			}
		}
	}
}

func filter(arr []string) []string {
	set := make(map[string]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}

	return arr[:j]
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
