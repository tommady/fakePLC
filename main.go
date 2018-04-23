package main

import (
	"flag"
	"log"
	"strings"
)

var (
	barcodes []string
	basketID string
	addr     string
)

func main() {
	bs := flag.String("barcodes", "8888351100042,9556166090085,8850025001023", "barcodes send to PLC socket server")
	flag.StringVar(&basketID, "basket_id", "3345678", "basket id send to PLC socket server")
	flag.StringVar(&addr, "addr", "localhost:10010", "PLC socker server address")
	endian := flag.String("endian", "BigEndian", "packet endian BigEndian or LittleEndian")
	ssid := flag.String("ssid", "world-wild-only-SSID", "the plc device connected wifi SSID")
	localPort := flag.Int("local_port", 33456, "the plc device port to dial plc server")
	flag.Parse()

	barcodes = strings.Split(*bs, ",")

	client, err := newClient(addr, *ssid, *endian, *localPort)
	if err != nil {
		log.Fatalf("new client failed:%v", err)
	}
	defer client.close()

	if err = client.purchase(basketID, barcodes); err != nil {
		log.Fatalf("write purchase failed:%v", err)
	}
}
