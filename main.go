package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"

	_ "github.com/google/gopacket/layers"
)

var (
	iface      = flag.String("i", "any", "Interface to read from")
	cpuprofile = flag.String("cpuprofile", "", "If non-empty, write CPU profile here")
	snaplen    = flag.Int("s", 0, "Snaplen, if <= 0, use 65535")
	bufferSize = flag.Int("b", 8, "Interface buffersize (MB)")
	filter     = flag.String("f", "port not 22", "BPF filter")
	addVLAN    = flag.Bool("add_vlan", false, "If true, add VLAN header")
	// 转发给的waap地址
	targetUrl = flag.String("t", "http://127.0.0.1:8080", "filter dst Url")
)

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		log.Printf("Writing CPU profile to %q", *cpuprofile)
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	log.Printf("Starting on interface %q", *iface)
	if *snaplen <= 0 {
		*snaplen = 65535
	}
	szFrame, szBlock, numBlocks, err := afpacketComputeSize(*bufferSize, *snaplen, os.Getpagesize())
	if err != nil {
		log.Fatal(err)
	}
	afpacketHandle, err := newAfpacketHandle(*iface, szFrame, szBlock, numBlocks, *addVLAN, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	err = afpacketHandle.SetBPFFilter(*filter, *snaplen)
	if err != nil {
		log.Fatal(err)
	}
	source := gopacket.ZeroCopyPacketDataSource(afpacketHandle)
	dlc := gopacket.DecodingLayerContainer(gopacket.DecodingLayerMap{})
	defer afpacketHandle.Close()

	// 创建重组器
	assembler := newAssembler()
	// 一分钟定时器
	ticker := time.Tick(time.Minute)

	for {
		d, _, err := source.ZeroCopyReadPacketData()
		if err != nil {
			log.Fatal(err)
		}

		dealWithData(assembler, dlc, d, time.Now(), ticker)
	}
}

func dealWithData(assembler *tcpassembly.Assembler, dlc gopacket.DecodingLayerContainer, d []byte, timestamp time.Time, ticker <-chan time.Time) {
	data := make([]byte, len(d))
	copy(data, d)
	var eth layers.Ethernet
	var ip4 layers.IPv4
	var tcp layers.TCP
	dlc = dlc.Put(&eth)
	dlc = dlc.Put(&ip4)
	dlc = dlc.Put(&tcp)
	decoderFunc := dlc.LayersDecoder(layers.LayerTypeEthernet, gopacket.NilDecodeFeedback)
	decodedLayers := make([]gopacket.LayerType, 0, 10)
	decoderFunc(data, &decodedLayers)
	assembler.AssembleWithTimestamp(tcp.TransportFlow(), &tcp, timestamp)

	if ticker != nil {
		select {
		case <-ticker:
			// 每隔 1 分钟，刷新之前 2 分钟内不活跃的连接
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		default:
		}
	}
}
