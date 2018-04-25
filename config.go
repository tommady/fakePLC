package main

import (
	"flag"
	"strings"
	"time"
)

type configuration struct {
	mode string

	// bench mode settings
	roundTimes  int
	roundPeriod time.Duration

	// normal settings
	barcodes    []string
	basketID    string
	connectAddr string
	endian      string
	ssid        string
	localPort   int
}

func newConfig() (*configuration, error) {
	c := new(configuration)

	flag.StringVar(&c.mode, "mode", "test-once", "test-once run for normal single test, test-many run for long-term")
	flag.StringVar(&c.basketID, "basket_id", "3345678", "basket id send to PLC socket server")
	flag.StringVar(&c.connectAddr, "addr", "localhost:10010", "PLC socker server address")
	flag.StringVar(&c.endian, "endian", "BigEndian", "packet endian BigEndian or LittleEndian")
	flag.StringVar(&c.ssid, "ssid", "world-wild-only-SSID", "the plc device connected wifi SSID")
	flag.IntVar(&c.localPort, "local_port", 33456, "the plc device port to dial plc server")
	bs := flag.String("barcodes", "8888351100042,9556166090085,8850025001023", "barcodes send to PLC socket server")
	flag.IntVar(&c.roundTimes, "round_times", 1000, "run how many times")
	rd := flag.Int("round_period_sec", 1, "run period second")

	flag.Parse()
	c.barcodes = strings.Split(*bs, ",")
	c.roundPeriod = time.Duration(*rd) * time.Second

	return c, nil
}
