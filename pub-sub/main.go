package main

import "messenger/common"

func main() {
	var (
		serviceName   = "gateway"
		httpAddress   = common.EnvString("HTTP_ADDR", ":8080")
		consulAddress = common.EnvString("CONSUL_ADDR", ":8500")
	)
}
