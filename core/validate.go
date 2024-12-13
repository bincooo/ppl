package core

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"ppl/core/logger"
	"strings"
	"time"
)

func verifyMap() {
	ticker := time.NewTicker(120 * time.Second)
	for range ticker.C {
		for key := range ElemMap {
			elem := ElemMap[key]
			switch elem.T {
			case "HTTP":
				if !h1(elem.Addr, elem.Port) {
					delete(ElemMap, key)
				}
			case "HTTPS":
				if !h2(elem.Addr, elem.Port) {
					delete(ElemMap, key)
				}
			case "SOCKS5":
				if !socks5(elem.Addr, elem.Port) {
					delete(ElemMap, key)
				}
			}
		}

		chunk, err := json.Marshal(ElemMap)
		if err != nil {
			logger.Error(err)
			continue
		}

		err = os.WriteFile("data.json", chunk, 0644)
		if err != nil {
			logger.Error(err)
		}
	}
}

func verify(addr string, port int) (ok bool, elem Elem) {
	if ok = h1(addr, port); ok {
		elem = Elem{
			Addr: addr,
			Port: port,
			T:    "HTTP",
		}
		return
	}

	if ok = h2(addr, port); ok {
		elem = Elem{
			Addr: addr,
			Port: port,
			T:    "HTTPS",
		}
		return
	}

	if ok = socks5(addr, port); ok {
		elem = Elem{
			Addr: addr,
			Port: port,
			T:    "SOCKS5",
		}
		return
	}

	return
}

func h1(addr string, port int) (ok bool) {
	u, err := url.Parse(fmt.Sprintf("http://%s:%d", addr, port))
	if err != nil {
		logger.Debugf("[H1] : %v", err)
		return
	}

	tr := http.Transport{
		Proxy: http.ProxyURL(u),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := http.Client{Timeout: 10 * time.Second, Transport: &tr}
	request, err := http.NewRequest("GET", "http://baidu.com", nil)
	res, err := client.Do(request)
	if err != nil {
		logger.Debugf("[H1] : %v", err)
		return
	}

	defer res.Body.Close()
	dataBytes, _ := io.ReadAll(res.Body)
	result := string(dataBytes)
	ok = strings.Contains(result, "0;url=http://www.baidu.com")
	return
}

func h2(addr string, port int) (ok bool) {
	destConn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", addr, port), 10*time.Second)
	if err != nil {
		return
	}
	defer destConn.Close()
	req := []byte{67, 79, 78, 78, 69, 67, 84, 32, 119, 119, 119, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 58, 52, 52, 51, 32, 72, 84, 84, 80, 47, 49, 46, 49, 13, 10, 72, 111, 115, 116, 58, 32, 119, 119, 119, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 58, 52, 52, 51, 13, 10, 85, 115, 101, 114, 45, 65, 103, 101, 110, 116, 58, 32, 71, 111, 45, 104, 116, 116, 112, 45, 99, 108, 105, 101, 110, 116, 47, 49, 46, 49, 13, 10, 13, 10}
	destConn.Write(req)
	bytes := make([]byte, 1024)
	err = destConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		logger.Debugf("[H2] : %v", err)
		return
	}

	read, err := destConn.Read(bytes)
	if err != nil {
		logger.Debugf("[H2] : %v", err)
		return
	}

	ok = strings.Contains(string(bytes[:read]), "200 Connection established")
	return
}

func socks5(addr string, port int) (ok bool) {
	destConn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", addr, port), 10*time.Second)
	if err != nil {
		logger.Debugf("[S5] : %v", err)
		return
	}
	defer destConn.Close()
	req := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	destConn.Write(req)
	bytes := make([]byte, 1024)
	err = destConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		logger.Debugf("[S5] : %v", err)
		return
	}

	_, err = destConn.Read(bytes)
	if err != nil {
		logger.Debugf("[S5] : %v", err)
		return
	}

	ok = bytes[0] == 5 && bytes[1] == 255
	return
}
