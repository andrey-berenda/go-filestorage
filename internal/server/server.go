package server

import (
	"bytes"
	_ "embed"
	"io"

	pb "github.com/andrey-berenda/go-filestorage/gen/api/v1"
	"github.com/andrey-berenda/go-filestorage/internal/file"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:embed static/gopher.png
var gopher []byte

var chunkSize = 1024 * 3

type server struct {
	pb.UnimplementedFileServiceServer
}

func New() pb.FileServiceServer {
	return server{}
}

func (s server) Download(req *pb.DownloadRequest, server pb.FileService_DownloadServer) error {
	if req.GetId() == "" {
		return status.Error(codes.InvalidArgument, "id is required")
	}

	f, ok := getFile(req.Id)
	if !ok {
		return status.Error(codes.NotFound, "file is not found")
	}
	err := server.SendHeader(f.Metadata())
	if err != nil {
		return status.Error(codes.Internal, "error during sending header")
	}

	chunk := &pb.DownloadResponse{Chunk: make([]byte, chunkSize)}
	var n int

Loop:
	for {
		n, err = f.Read(chunk.Chunk)
		switch err {
		case nil:
		case io.EOF:
			break Loop
		default:
			return status.Errorf(codes.Internal, "io.ReadAll: %v", err)
		}
		chunk.Chunk = chunk.Chunk[:n]
		serverErr := server.Send(chunk)
		if serverErr != nil {
			return status.Errorf(codes.Internal, "server.Send: %v", serverErr)
		}
	}
	return nil
}

func getFile(fileID string) (*file.File, bool) {
	if fileID != "gopher" {
		return nil, false
	}
	return file.NewFile("gopher", "png", len(gopher), bytes.NewReader(gopher)), true
}
