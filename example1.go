package main

// Not an actual main. It's more of a sandbox/example of use.

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jpmondet/netmiGo/pkg/devices"
	"github.com/jpmondet/netmiGo/pkg/push"
	"github.com/jpmondet/netmiGo/pkg/session"

	"github.com/google/goterm/term"
)

var (
	addr       = flag.String("address", "sbx-nxos-mgmt.cisco.com", "Address of the remote device")
	port       = flag.String("port", "8181", "Port on which ssh is listening on  the remote device")
	user       = flag.String("user", "admin", "Username to use")
	pass1      = flag.String("pass1", "Admin_1234!", "Password to use")
	pass2      = flag.String("pass2", "pass2", "Alternate password to use (for example 'enable pass' on cisco devices)")
	cmd        = flag.String("cmd", "show vlan", "show routing table")
	deviceType = flag.String("dev_t", "cisco_nxos", "Type/model of the device. Currently, it supports legacy ios cli (thus cisco, arista, etc)")
)

// Func used in Example 1
func parallelMultiDev(dev devices.Device, wg *sync.WaitGroup, cmdsToSend ...string) {
	defer wg.Done()

	bc, err := session.ConnectionHandler(dev)
	if err != nil {
		log.Fatal("Can't connect to the device: ", dev.Addr, err)
	}

	res, err := push.SendCommands(bc, cmdsToSend...)
	if err != nil {
		log.Printf("There was an error sending the cmd %s to the device: %s", cmdsToSend, dev.Addr)
		log.Println("The error was:", err)
		os.Exit(1)
	}
	fmt.Println(term.Greenf("result: %s\n", res))

}

func main() {
	// Profile testing
	// defer profile.Start(profile.MemProfile).Stop()

	// Example of getting input from cli
	flag.Parse()

	fmt.Println(term.Bluef("SSH to %q:%q", *addr, *port))

	// Example of sending cmd to user defined + incode defined devices
	remoteDevice := devices.Device{Addr: *addr, Port: *port, User: *user, Pass1: *pass1, Pass2: *pass2, DeviceType: *deviceType}
	remoteDevice2 := devices.Device{
		Addr:       "rviews.kanren.net",
		Port:       "22",
		User:       "rviews",
		Pass1:      "rviews",
		DeviceType: "junos"}

	var devSlice = map[devices.Device][]string{
		remoteDevice: []string{
			"sh ip route", "sh version",
		}, remoteDevice2: []string{
			"show route 1.1.1.0/24", "traceroute 1.1.1.1",
		},
	}

	// Example 1 :
	// Sending cmds in parallel to multiple netdevices
	var wg sync.WaitGroup
	wg.Add(len(devSlice))
	for dev, cmdsToSend := range devSlice {
		go parallelMultiDev(dev, &wg, cmdsToSend...)
	}
	wg.Wait()

	fmt.Println(term.Greenf("All done"))
	os.Exit(0)
}
