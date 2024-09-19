package main

import (
	_ "github.com/zoninnik89/messenger/common"
)

var (
	serviceName   = "pub-sub"
	httpAddress   = common.EnvString("HTTP_ADDR", ":8080")
	consulAddress = common.EnvString("CONSUL_ADDR", ":8500")
)

func main() {

}
