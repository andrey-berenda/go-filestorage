package client

import (
	"context"
	"fmt"
	"io"

	pb "github.com/andrey-berenda/go-filestorage/gen/api/v1"
	"github.com/andrey-berenda/go-filestorage/internal/file"
)

type Client struct {
	client pb.FileServiceClient
}

func New(client pb.FileServiceClient) Client {
	return Client{client: client}
}

func copyFromResponse(w *io.PipeWriter, res pb.FileService_DownloadClient) {
	message := new(pb.DownloadResponse)
	var err error
	for {
		err = res.RecvMsg(message)
		if err == io.EOF {
			_ = w.Close()
			break
		}
		if err != nil {
			_ = w.CloseWithError(err)
			break
		}
		if len(message.GetChunk()) > 0 {
			_, err = w.Write(message.Chunk)
			if err != nil {
				_ = res.CloseSend()
				break
			}
		}
		message.Chunk = message.Chunk[:0]
	}
}

func (c Client) GetFile(ctx context.Context, fileID string) (*file.File, error) {
	response, err := c.client.Download(
		ctx,
		&pb.DownloadRequest{Id: fileID},
	)
	if err != nil {
		return nil, fmt.Errorf("client.LoadFile: %w", err)
	}
	md, err := response.Header()
	if err != nil {
		return nil, fmt.Errorf("response.Header: %w", err)
	}
	r, w := io.Pipe()
	f := file.NewFromMetadata(md, r)
	go copyFromResponse(w, response)
	return f, nil
}
