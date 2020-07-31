package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

var (
	fn          string
	random      bool
	concurrency int
	interval    int
)

func main() {
	flag.StringVar(&fn, "f", "", "File of domains to resolve domain")
	flag.BoolVar(&random, "r", false, "Generate random domains")
	flag.IntVar(&concurrency, "c", 10, "Concurrency to resolve domain")
	flag.IntVar(&interval, "i", 100, "Interval of resolve")
	flag.Parse()
	fmt.Printf("file:%s, random:%+v, concurrency:%d, interval:%d \n", fn, random, concurrency, interval)
	domains := []string{}
	if len(fn) != 0 {
		fi, err := os.Open(fn)
		if err != nil {
			panic(err)
		}
		br := bufio.NewReader(fi)
		for {
			line, _, err := br.ReadLine()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("ReadLine error:%+v", err)
				continue
			}
			domains = append(domains, string(line))
		}
		fi.Close()
	}
	if len(domains) == 0 {
		domains = append(domains, "baidu.com")
		domains = append(domains, "qq.com")
		domains = append(domains, "github.com")
		domains = append(domains, "taobao.com")
	}
	for i := 0; i < concurrency; i++ {
		go func() {
			for {

				rand.Seed(time.Now().Unix())
				index := rand.Intn(len(domains))
				dom := domains[index]
				if random {
					rdom := strconv.Itoa(rand.Intn(1000000000)) + "." + dom
					start := time.Now()
					address, err := net.LookupHost(rdom)
					fmt.Printf("ts:%d|dom:%s|address length:%d|duration:%d|err:%+v \n", time.Now().Unix(), rdom, len(address), time.Since(start)/1e6, err)
				}
				start := time.Now()
				address, err := net.LookupHost(dom)
				fmt.Printf("ts:%d|dom:%s|address length:%d|duration:%d|err:%+v \n", time.Now().Unix(), dom, len(address), time.Since(start)/1e6, err)
				time.Sleep(time.Duration(interval) * time.Millisecond)
			}
		}()
	}
	select {}
}