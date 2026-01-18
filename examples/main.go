package main

import (
	"log"
	"net/http"

	"github.com/rbrick/go-ccip-read"
)

func main() {
	resolver := ccip.NewCCIPReadResolver()

	resolver.Handle("function addr(bytes32 namehash) view returns (bytes)", func(request *ccip.CCIPReadRequest) ([]interface{}, error) {
		namehash, ok := request.Var("namehash")
		if !ok {
			return nil, nil
		}

		log.Println(namehash)
		return nil, nil
	})

	http.ListenAndServe(":8080", resolver)
}
