package main

import (
	"log"
	"time"
)

func main() {
	conf, err := newConfig()
	if err != nil {
		log.Fatalf("new config failed:%v", err)
	}

	client, err := newClient(conf.connectAddr, conf.ssid, conf.endian, conf.localPort)
	if err != nil {
		log.Fatalf("new client failed:%v", err)
	}
	defer client.close()

	switch conf.mode {
	case "test-once":
		if err = client.purchase(conf.basketID, conf.barcodes); err != nil {
			log.Fatalf("write purchase failed:%v", err)
		}
	case "test-many":
		for i := 0; i < conf.roundTimes; i++ {
			if err = client.purchase(conf.basketID, conf.barcodes); err != nil {
				log.Fatalf("write purchase failed:%v", err)
			}
			time.Sleep(conf.roundPeriod)
		}
	}
}
