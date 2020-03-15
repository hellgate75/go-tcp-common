package common

import (
	"fmt"
	"net"
	"testing"
)

func TestAddressCount(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
			panic(err.Error())
	}
	var expectedAddressCount uint64 = 256
	addressCount := AddressCount(ipnet)
	if expectedAddressCount != addressCount {
		t.Fatalf("TestAddressCount - net/common.AddressCount - Expected: %v but Given: %v", expectedAddressCount, addressCount)
	}
	_, ipnet, err = net.ParseCIDR("192.168.0.0/16")
	if err != nil {
		panic(err.Error())
	}
	expectedAddressCount = 65536
	addressCount = AddressCount(ipnet)
	if expectedAddressCount != addressCount {
		t.Fatalf("TestAddressCount - net/common.AddressCount - Expected: %v but Given: %v", expectedAddressCount, addressCount)
	}
}


func TestListAddresses(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		panic(err.Error())
	}
	addresses := ListAddresses(ipnet)
	var expectedAddressCount int = 256
	addressCount := len(addresses)
	if expectedAddressCount != addressCount {
		t.Fatalf("TestListAddresses - net/common.ListAddresses - Expected Len: %v but Given: %v", expectedAddressCount, addressCount)
	}
	expectedFirstAddress:="192.168.1.0"
	firstAddress:=fmt.Sprintf("%s", addresses[0].String())
	if expectedFirstAddress != firstAddress {
		t.Fatalf("TestListAddresses - net/common.ListAddresses - Expected First Address: %s but Given: %s", expectedFirstAddress, firstAddress)
	}
	expectedLastAddress:="192.168.1.255"
	lastAddress:=fmt.Sprintf("%s", addresses[len(addresses)-1].String())
	if expectedLastAddress != lastAddress {
		t.Fatalf("TestListAddresses - net/common.ListAddresses - Expected Last Address: %s but Given: %s", expectedLastAddress, lastAddress)
	}

	_, ipnet, err = net.ParseCIDR("192.168.0.0/16")
	if err != nil {
		panic(err.Error())
	}
	addresses = ListAddresses(ipnet)
	expectedAddressCount = 65536
	addressCount = len(addresses)
	if expectedAddressCount != addressCount {
		t.Fatalf("TestListAddresses - net/common.ListAddresses - Expected Len: %v but Given: %v", expectedAddressCount, addressCount)
	}
	expectedFirstAddress="192.168.0.0"
	firstAddress=fmt.Sprintf("%s", addresses[0].String())
	if expectedFirstAddress != firstAddress {
		t.Fatalf("TestListAddresses - net/common.ListAddresses - Expected First Address: %s but Given: %s", expectedFirstAddress, firstAddress)
	}
	expectedLastAddress="192.168.255.255"
	lastAddress=fmt.Sprintf("%s", addresses[len(addresses)-1].String())
	if expectedLastAddress != lastAddress {
		t.Fatalf("TestListAddresses - net/common.ListAddresses - Expected Last Address: %s but Given: %s", expectedLastAddress, lastAddress)
	}
}

//func main() {
//	ip, ipnet, err := net.ParseCIDR("192.168.1.0/24")
//	if err != nil {
//		panic(err.Error())
//	}
//	fmt.Printf("IP: %s\n", ip.String())
//	fmt.Printf("IP Net: %s\n" , (*ipnet).String())
//	list := ListAddresses(ipnet)
//	fmt.Printf("Expected IP Len: %v\n" , AddressCount(ipnet))
//	fmt.Printf("Found IP Len: %v\n" , len(list))
//	fmt.Printf("First IP Addr: %v\n" , list[0])
//	fmt.Printf("Last IP Addr: %v\n" , list[len(list)-1])
//	fmt.Println()
//	ip, ipnet, err = net.ParseCIDR("192.168.0.0/16")
//	if err != nil {
//		panic(err.Error())
//	}
//	fmt.Printf("IP: %s\n", ip.String())
//	fmt.Printf("IP Net: %s\n" , (*ipnet).String())
//	list = ListAddresses(ipnet)
//	fmt.Printf("Expected IP Len: %v\n" , AddressCount(ipnet))
//	fmt.Printf("Found IP Len: %v\n" , len(list))
//	fmt.Printf("First IP Addr: %v\n" , list[0])
//	fmt.Printf("Last IP Addr: %v\n" , list[len(list)-1])
//
//}
