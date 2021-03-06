package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
)

func main() {
	var err error
	var addrs []net.Addr

	var ifaces []net.Interface

	ifaces, err = net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, err = i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			log.Printf("addr: %v", ip)
		}
	}

	fs := http.FileServer(http.Dir("./html"))
	http.Handle("/", fs)

	log.Println("Listening on :3000...")
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
