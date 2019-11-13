package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

func handleConnection(c net.Conn, workData *timeMap, deadline int64) {
	defer c.Close()
	client := c.RemoteAddr().String()
	buf := make([]byte, 128)

	err := c.SetDeadline(time.Now().Add(time.Duration(deadline) * time.Second))
	if err != nil {
		log.Printf("Fail to set connection deadline %v :: %v \n", client, err)
		return
	}
	reqLen, err := c.Read(buf)
	if err != nil {
		log.Printf("Can`not read data from %v :: %v \n", client, err)
		return
	}
	notice, err := newNotice(buf[:reqLen])
	if err != nil {
		log.Printf("Wrong data from %v :: %v \n", client, err)
		return
	}
	workData.add(notice)
}

func main() {
	address := flag.String("tcpaddr", "0.0.0.0", "Set listening address")
	port := flag.Int64("tcpport", 6000, "Listening tcp port")
	ttl := flag.Int64("ttl", 10, "Row`s lifetime")
	cleanPeriod := flag.Int64("clean_period", 1, "data cleaning frequency")
	outPeriod := flag.Int64("out_period", 1, "Data output frequency")
	clientDeadline := flag.Int64("client_deadline", 10, "Deadline to receive a notice")
	flag.Parse()

	workData := newTimeMap(*ttl, *cleanPeriod)
	go func() {
		for range time.Tick(time.Duration(*outPeriod) * time.Second) {
			log.Println("workData: \n" + workData.String())
		}
	}()

	servString := fmt.Sprintf("%s:%d", *address, *port)
	listener, err := net.Listen("tcp", servString)
	defer listener.Close()
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Server run at %v", servString)

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		go handleConnection(connection, workData, *clientDeadline)
	}
}
