package main

import (
	"bluego/discovery"
	"fmt"
	"rsc.io/quote"
)

func main() {
	fmt.Println(quote.Hello())
	discovery.Run("11234", false)
}
