package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/coredns/coredns/plugin/pkg/parse"
	"github.com/miekg/dns"
	"io"
	"log"
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
	outputPath  string
	writeFile   *os.File
	proto       string
	to          string
)

func main() {
	flag.StringVar(&fn, "f", "", "File of domains to resolve domain")
	flag.BoolVar(&random, "r", false, "Generate random domains")
	flag.IntVar(&concurrency, "c", 10, "Concurrency to resolve domain")
	flag.IntVar(&interval, "i", 100, "Interval of resolve")
	flag.StringVar(&outputPath, "o", "/tmp/dns-stress", "File output path")
	flag.StringVar(&proto, "p", "udp", "request protocal")
	flag.StringVar(&to, "t", "/etc/resolv.conf", "the destination endpoints to forward to")
	flag.Parse()
	initFile(outputPath)
	fmt.Printf("file:%s, random:%+v, concurrency:%d, interval:%d ,output path:%s \n", fn, random, concurrency, interval, outputPath)
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
	if proto == "udp" {
		udp(domains)
	} else {
		tcp(domains, to)
	}
	select {}
}

func initFile(output string) {
	file, err := os.OpenFile(output+"/result.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		log.Printf("open file error:%+v", err)
		panic("open file fail")
	}
	writeFile = file
}

func udp(domains []string) {
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
					if err != nil {
						writeFile.Write([]byte(fmt.Sprintf("%d|udp:%s|%d|%d|%+v\n", time.Now().Unix(), rdom, len(address), time.Since(start)/1e6, err)))
					} else {
						writeFile.Write([]byte(fmt.Sprintf("%d|udp:%s|%d|%d|\n", time.Now().Unix(), rdom, len(address), time.Since(start)/1e6)))
					}
				}
				start := time.Now()
				address, err := net.LookupHost(dom)
				if err != nil {
					writeFile.Write([]byte(fmt.Sprintf("%d|udp:%s|%d|%d|%+v\n", time.Now().Unix(), dom, len(address), time.Since(start)/1e6, err)))
				} else {
					writeFile.Write([]byte(fmt.Sprintf("%d|udp:%s|%d|%d|\n", time.Now().Unix(), dom, len(address), time.Since(start)/1e6)))
				}
				time.Sleep(time.Duration(interval) * time.Millisecond)
			}
		}()
	}
}

func tcp(domains []string, to string) {
	for i := 0; i < concurrency; i++ {
		go func() {
			conns := newConnect(to)
			for {
				rand.Seed(time.Now().Unix())
				index := rand.Intn(len(domains))
				dom := domains[index]
				if random {
					rdom := strconv.Itoa(rand.Intn(1000000000)) + "." + dom
					start := time.Now()
					resultLen, err := request(conns[rand.Intn(len(conns))], rdom)
					if err != nil {
						writeFile.Write([]byte(fmt.Sprintf("%d|tcp:%s|%d|%d|%+v\n", time.Now().Unix(), rdom, resultLen, time.Since(start)/1e6, err)))
					} else {
						writeFile.Write([]byte(fmt.Sprintf("%d|tcp:%s|%d|%d|\n", time.Now().Unix(), rdom, resultLen, time.Since(start)/1e6)))
					}
				}
				start := time.Now()
				resultLen, err := request(conns[rand.Intn(len(conns))], dom)
				if err != nil {
					writeFile.Write([]byte(fmt.Sprintf("%d|tcp:%s|%d|%d|%+v\n", time.Now().Unix(), dom, resultLen, time.Since(start)/1e6, err)))
				} else {
					writeFile.Write([]byte(fmt.Sprintf("%d|tcp:%s|%d|%d|\n", time.Now().Unix(), dom, resultLen, time.Since(start)/1e6)))
				}
				time.Sleep(time.Duration(interval) * time.Millisecond)
			}
		}()
	}
}

func request(conn *dns.Conn, domain string) (int, error) {
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	req := new(dns.Msg)
	req.SetQuestion(domain+".", dns.TypeA)
	err := conn.WriteMsg(req)
	if err != nil {
		return 0, err
	}
	var ret *dns.Msg
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		ret, err = conn.ReadMsg()
		if err != nil && err == io.EOF {
			return 0, err
		}
		if req.Id == ret.Id {
			break
		}
	}
	return len(ret.Answer), nil
}

func newConnect(to string) []*dns.Conn {
	toHosts, err := parse.HostPortOrFile(to)
	if err != nil {
		panic(err)
	}
	var conns []*dns.Conn
	for _, host := range toHosts {
		conn, err := dns.DialTimeout("tcp", host, 5*time.Second)
		if err != nil {
			panic(err)
		}
		conns = append(conns, conn)
	}
	return conns
}
