package main

import (
	"corpc"
	pb "corpc/proto"
	"fmt"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", ":8000")
	if err != nil {
		log.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	corpc := corpc.NewClient(conn)
	greeterClient := pb.NewGreeterClient(corpc)
	reply, err := greeterClient.SayHello(&pb.HelloRequest{Name: "cong"})
	if err != nil {
		log.Fatalf("SayHello error: %v", err)
	}
	fmt.Println(reply)
}
