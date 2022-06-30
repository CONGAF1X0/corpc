package main

import (
	"corpc"
	pb "corpc/proto"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("error listening: %v", err)
	}
	corpc := corpc.NewServer()

	pb.RegisterGreeterServer(corpc, &pb.Greeter{})

	corpc.Serve(listener)
}
