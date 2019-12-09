package main

import (
	_ "github.com/nymtech/nym-directory/docs"
	"github.com/nymtech/nym-directory/server"
)

// @title Nym Directory API
// @version 0.0.4
// @description This is a temporarily centralized directory/PKI/metrics API to allow us to get the other Nym node types running. Its functionality will eventually be folded into other parts of Nym.
// @termsOfService http://swagger.io/terms/

// @license.name Apache 2.0
// @license.url https://github.com/nymtech/nym-directory/license
func main() {
	router := server.New()
	router.Run(":8080")
}
