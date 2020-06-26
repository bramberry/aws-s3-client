package simplerest

import (
	"fmt"
	"net/http"
)

func Start() error {

	srv := newServer()
	fmt.Printf("Started on: %s port", Config.BindPort)
	return http.ListenAndServe(Config.BindPort, srv)
}
