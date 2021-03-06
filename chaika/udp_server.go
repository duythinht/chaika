package chaika

import (
	"fmt"
	"net"
	"strconv"

	"github.com/duythinht/chaika/config"
	"github.com/duythinht/chaika/courier"
)

func RunServer() {
	cfg := config.GetConfig()

	listenAddr := ":" + strconv.FormatInt(cfg.Port, 10)
	ServerAddr, err := net.ResolveUDPAddr("udp", listenAddr)

	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	fmt.Println("Server is up and listen on " + listenAddr)

	CheckError(err)

	defer ServerConn.Close()

	go RunMonitor()

	courier.Setup()

	for {
		// n, add, err
		buffer := make([]byte, 32678)
		length, _, err := ServerConn.ReadFromUDP(buffer)

		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}

		log, parseError := ParseLog(buffer[:length])

		if parseError != nil {
			fmt.Println("Error: ", err)
			continue
		}

		g := courier.Get(log.Service)

		fmt.Println(log.Service, ":", log.Message)
		SendOverMonitor(log.Service + ":" + log.Message + "\n")
		g.Send(log.Service, log.Catalog, log.Level, log.Message)
	}
}
