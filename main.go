package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"golang.org/x/sync/semaphore"
)

type Request struct {
	Protocol string
	Version uint8
    Method   string
}

type ClientJob struct {
	Client *net.UDPAddr
	Data []byte
}

var port int = 23500
var sentinelProtocolVersion uint8 = 0
var maxWorkers int64 = 4
var clientJobChan = make(chan ClientJob)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)

	addr, err := net.ResolveUDPAddr("udp4", ":" + strconv.Itoa(port))
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	log.Println("Sentinel service running on " + addr.IP.String() + ":" + strconv.Itoa(addr.Port))

	buffer := make([]byte, 1024)

	go handleClients(conn, &clientJobChan)

	go func() {
		for {
			n, client, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Println("Failed to receive packet", err)
				continue
			}

			data := make([]byte, n)
			copy(data, buffer[:n])

			clientJobChan <- ClientJob{Client: client, Data: data}
		}
	}()

	<-quit
	
	log.Println("Shutting down Sentinel service...")
	
	close(clientJobChan)
}

func handleClients(conn *net.UDPConn, jobs *chan ClientJob) {
	sem := semaphore.NewWeighted(maxWorkers)
	for job := range *jobs {
		go func() {
			sem.Acquire(context.Background(), 1)
			defer sem.Release(1)

			handleRequest(conn, job)
		}()
	}
}

func handleRequest(conn *net.UDPConn, job ClientJob) {
	data := job.Data

	req, err := parseRequest(data)
	if err != nil {
		log.Println("Failed to parse request!")
		return
	}

	log.Println("[" + job.Client.String() + "] " + req.Protocol + "/" + strconv.Itoa(int(req.Version)) + " " + req.Method)

	if req.Version != sentinelProtocolVersion {
		log.Println("Protocol version mismatch detected!")
	}

	var fn func(*net.UDPConn, Request, *net.UDPAddr)
	switch req.Method {
	case "#external_address":
		fn = externalAddressHandler
	}

	fn(conn, req, job.Client)
}

func externalAddressHandler(conn *net.UDPConn, req Request, addr *net.UDPAddr) {
	conn.WriteToUDP(addr.IP.To4(), addr)
}

func parseRequest(data []byte) (Request, error) {
	str := string(data)
	parts := strings.Fields(str)

	protocolParts := strings.Split(parts[0], "/")

	protocol := protocolParts[0]

	version, err := strconv.Atoi(protocolParts[1])
	if err != nil {
		return Request{}, err
	}
	method := parts[1]

	return Request{
		Protocol: protocol,
		Version: uint8(version),
		Method: method,
	}, nil
}