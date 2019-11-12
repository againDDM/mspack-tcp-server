package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

func handleConnection(c net.Conn, workData *timeMap, deadline int64) {
	defer c.Close()
	client := c.RemoteAddr().String()

	scanner := bufio.NewScanner(c)
	for {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				log.Printf("Error reading from %v :: %v \n", client, err)
			}
			break
		}

		err := c.SetDeadline(time.Now().Add(time.Duration(deadline)))
		if err != nil {
			log.Printf("Fail to set connection deadline %v :: %v \n", client, err)
			break
		}
		notice, err := newNotice(scanner.Bytes())
		if err != nil {
			log.Printf("Wrong data from %v :: %v \n", client, err)
			break
		}
		workData.add(notice)
	}
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
