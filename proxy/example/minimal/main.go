package main

import (
	"github.com/hossinasaadi/warp-plus/proxy/pkg/mixed"
)

func main() {
	proxy := mixed.NewProxy()
	_ = proxy.ListenAndServe()
}
