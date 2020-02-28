package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/kravcs/jff-say-grpc/api"
)

func main() {
	// define a flag and parse it
	port := flag.Int("p", 8080, "port to listen to")
	flag.Parse()

	logrus.Infof("listening to port %d", *port)

	// create a listener of that port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logrus.Fatalf("could not listen to port %d: %v", *port, err)
	}

	s := grpc.NewServer()
	pb.RegisterTextToSpeechServer(s, server{})
	err = s.Serve(lis)
	if err != nil {
		logrus.Fatalf("could not serve: %v", err)
	}
}

type server struct{}

func (server) Say(ctx context.Context, text *pb.Text) (*pb.Speech, error) {
	// create temp file to write the audio
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, fmt.Errorf("could not create temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("could not close %s: %v", f.Name(), err)
	}

	// try to write the audio
	cmd := exec.Command("flite", "-t", text.Text, "-o", f.Name())
	if data, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("flite failed %s:", data)
	}

	// read the temp file
	data, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("could not read temp file: %v", err)
	}

	// return the audio data
	return &pb.Speech{Audio: data}, nil
}
