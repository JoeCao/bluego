package main

import (
	"bluego/discovery"
	"fmt"
	"rsc.io/quote"
)

func main() {
	fmt.Println(quote.Hello())
	var list, _ = discovery.Run("hci0", false)
	for _, pro := range list {
		fmt.Printf("获取到设备名%s 设备地址%s ,设备类型%s, 强度%d",
			pro.Name, pro.Address, pro.AddressType, pro.RSSI)
	}
}
