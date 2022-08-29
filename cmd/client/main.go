package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	pb "github.com/andrey-berenda/go-filestorage/gen/api/v1"
	"github.com/andrey-berenda/go-filestorage/internal/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var addr string
var path string
var version int

func init() {
	flag.StringVar(&addr, "address", "localhost:8000", "filestorage address")
	flag.StringVar(&path, "path", "", "Path to store a file")
	flag.IntVar(&version, "version", 1, "Version of client")
}

func main() {
	flag.Parse()
	ctx := context.Background()
	connCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	conn, err := grpc.DialContext(
		connCtx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("grpc.DialContext: %v", err)
	}

	c := client.New(pb.NewFileServiceClient(conn))

	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("os.Create(%#v): %v", path, err)
	}
	defer f.Close()

	file, err := c.GetFile(ctx, "gopher")
	if err != nil {
		log.Fatalf("GetFile: %v", err)
	}
	_, _ = f.ReadFrom(file)
}
