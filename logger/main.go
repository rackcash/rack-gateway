package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	pb "infra/pkg/protos/gen/go"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedLogServer
}

const PORT = "11111"

func (s *server) SendLog(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {

	var logstream string

	var bearer = base64.RawStdEncoding.EncodeToString([]byte(os.Getenv("PARSEABLE_USERNAME") + ":" + os.Getenv("PARSEABLE_PASSWORD")))

	// invoices, nats
	switch req.GetLogstream() {
	case pb.Logstream_INVOICES:
		logstream = "invoices"
	case pb.Logstream_FATAL:
		logstream = "fatal"
	case pb.Logstream_NATS:
		logstream = "nats"
	case pb.Logstream_WEBHOOKS:
		logstream = "webhooks"
	default:
		return &pb.LogResponse{
			Msg:     "invalid logstream: " + req.GetLogstream().String(),
			IsError: true,
		}, nil
	}

	sendLog(os.Getenv("PARSEABLE_URL")+"/api/v1/logstream/"+logstream, bearer, req.GetPayload())

	return &pb.LogResponse{
		Msg:     "ok",
		IsError: false,
	}, nil
}

func Run(port string) {
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Logger already running")
		return
	}

	grpcServer := grpc.NewServer()

	pb.RegisterLogServer(grpcServer, &server{})

	fmt.Println("Log server started")
	if err := grpcServer.Serve(l); err != nil {
		panic(err)
	}

}
func main() {
	err := godotenv.Load(os.Getenv("ENVPATH"))
	if err != nil {
		panic(err)
	}

	Run(PORT)
}

func sendLog(url, bearer string, log []byte) error {
	fmt.Println(url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(log))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if string(body) != "" {
		return errors.New(string(body))
	}
	return nil

}
