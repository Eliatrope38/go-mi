package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/go-ble/ble/linux"
	"github.com/go-ble/ble/linux/hci/cmd"
)

var (
	device = flag.String("device", "default", "implementation of ble hci0-hci1-etc...")
	name   = flag.String("name", "MJ_HT_V1", "name of Xiaomi Sensor")
)

func main() {
	// parse cmd line argument
	flag.Parse()
	// create channel for Ctrl+c or other signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Setup BLE
	d, err := dev.NewDevice(*device)
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)

	// This part is requested only on certain device
	// where MAC Address is not set by Hardware
	if dev, ok := d.(*linux.Device); ok {
		if err := dev.HCI.Send(&cmd.LESetRandomAddress{
			RandomAddress: [6]byte{0xFF, 0x11, 0x22, 0x33, 0x44, 0x55},
		}, nil); err != nil {
			log.Fatalln(err, "can't set random address")
		}
	}

	// Create Cancellation Context
	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))

	// Default to search device with name of MJ_HT_V1 (or specified by user).
	filter := func(a ble.Advertisement) bool {
		return strings.ToUpper(a.LocalName()) == strings.ToUpper(*name)
	}

	log.Printf("Scanning for devices : %s", *name)
	// start Scannig in another Thread
	go ble.Scan(ctx, true, scanAdvertissement, filter)

	// wait for end of program
	c := <-sigs
	log.Printf("Signal Received - %s - Wait Few Seconds: ", c)
	ctx.Done()
	time.Sleep(5 * time.Second)
}

// Advertissement frame call baclk
func scanAdvertissement(a ble.Advertisement) {
	// check all Service Data for 0xfe95 as describe here
	// https://github.com/hannseman/homebridge-mi-hygrothermograph
	for _, chara := range a.ServiceData() {
		if chara.UUID.Equal(ble.MustParse("fe95")) {
			// if not 18 it's not a comple packet
			if len(chara.Data) != 18 {
				return
			}
			// extract only end of frame
			info := chara.Data[len(chara.Data)-4:]
			// extract temp
			temp := float64(int16(info[1])<<8+int16(info[0])) / 10.0
			// extract humidity
			humidity := float64(int16(info[3])<<8+int16(info[2])) / 10.0
			// log it
			log.Printf("%s - %s [ %v °C - %v %%]\n", a.LocalName(), a.Addr().String(), temp, humidity)
			//log.Printf("%s - %s [ %v °C - %v %%]\n [% X] %d\n", a.LocalName(), a.Addr().String(), temp, humidity, chara.Data, len(chara.Data))
		}
	}

}
