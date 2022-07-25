# corpc

# Usage
Server:
```go
listener, err := net.Listen("tcp", ":8000")
if err != nil {
    log.Fatalf("error listening: %v", err)
}

corpc := corpc.NewServer()
pb.RegisterGreeterServer(corpc, &pb.Greeter{})
corpc.Serve(listener)
```

Client:
```go
// conn, err := net.Dial("tcp", ":8000")
conn, err := net.DialTimeout("tcp", ":8000", time.Second)
if err != nil {
	  log.Fatalf("dial error: %v", err)
}
defer conn.Close()

corpc := corpc.NewClient(conn, corpc.WithTimeout(time.Second))

greeterClient := pb.NewGreeterClient(corpc)

reply, err := greeterClient.SayHello(&pb.HelloRequest{Name: "cong"})
if err != nil {
	  log.Fatalf("SayHello error: %v", err)
}
```
