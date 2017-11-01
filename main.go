package main

import (
	"fmt"
	// "net"
	"github.com/pokidovea/mimicro/config"
	// "strconv"
)

// func GetFreePort() (int, error) {
// 	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
// 	if err != nil {
// 		return 0, err
// 	}
// 	fmt.Printf(addr.String() + "\n")
// 	l, err := net.ListenTCP("tcp", addr)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer l.Close()
// 	return l.Addr().(*net.TCPAddr).Port, nil
// }

func main() {
	// port, _ := GetFreePort()
	// fmt.Printf(strconv.Itoa(port))

	servers, err := config.Load("/home/pokidovea/Dropbox/projects/go/src/github.com/pokidovea/mimicro/config/main.yml")

	if err != nil {
		panic(err)
	}
	fmt.Printf(servers.Servers[0].Endpoints[0].Url)
}
