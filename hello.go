package main

import (
	"bluego/discovery"
	"fmt"
	"rsc.io/quote"
)

func main() {
	fmt.Println(quote.Hello())
	_ = discovery.Run("hci0", false)
	select {}
}
