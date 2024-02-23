package main

import (
	"bufio"
	"io"
	"log"
	"net/http"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

// httpStream 真正地处理 HTTP 请求的解码
type httpStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
}

func (h *httpStream) run() {
	buf := bufio.NewReader(&h.r)
	for {
		req, err := http.ReadRequest(buf)
		if err == io.EOF {
			// 必须读到 EOF...非常重要！
			// log.Println("EOF end...")
			return
		} else if err != nil {
			log.Printf("Error reading stream, %s", err)
			continue
		}
		if req != nil {
			body, _ := io.ReadAll(req.Body)
			_ = req.Body.Close()
			// log.Printf("http request: %#v\n", *req)
			// log.Println(string(body))
			// 发送 Request 到 waap
			go sendToWaap(req, body)
		}
	}
}

// 使用 tcpassembly.StreamFactory 和 tcpassembly.Stream 接口，构建 HTTP 请求解析器。
// httpStreamFactory 实现 tcpassembly.StreamFactory
type httpStreamFactory struct{}

func (h *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	hs := &httpStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
	}
	// 重要...必须保证读取 reader 流的数据
	go hs.run()
	// ReaderStream 实现 tcpassembly.Stream, 因此可以返回指向它的指针
	return &hs.r
}

func newAssembler() *tcpassembly.Assembler {
	// 设置 TCP 流重组
	// 1. 创建 httpStreamFactory 结构体，实现 tcpassembly.StreamFactory 接口
	streamFactory := &httpStreamFactory{}
	// 2. 创建连接池
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	// 3. 创建重组器
	assembler := tcpassembly.NewAssembler(streamPool)
	return assembler
}
