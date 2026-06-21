package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"distributed-url-shortener/internal/fsm"
	pb "distributed-url-shortener/proto"

	"github.com/hashicorp/raft"
)

type GRPCServer struct {
	pb.UnimplementedURLShortenerServer
	Raft *raft.Raft
	FSM  *fsm.URLStore
}

func (s *GRPCServer) Shorten(ctx context.Context, req *pb.ShortenRequest) (*pb.ShortenResponse, error) {
	if s.Raft.State() != raft.Leader {
		return nil, fmt.Errorf("not leader")
	}
	code := req.LongUrl
	if len(code) > 6 {
		code = code[len(code)-6:]
	} else {
		code = "abc123"
	}
	cmd := fsm.Command{Op: "shorten", Code: code, URL: req.LongUrl}
	data, _ := json.Marshal(cmd)
	future := s.Raft.Apply(data, 5*time.Second)
	if err := future.Error(); err != nil {
		return nil, err
	}
	return &pb.ShortenResponse{ShortCode: code}, nil
}

func (s *GRPCServer) Resolve(ctx context.Context, req *pb.ResolveRequest) (*pb.ResolveResponse, error) {
	if s.Raft.State() != raft.Leader {
		return nil, fmt.Errorf("not leader")
	}
	url, found := s.FSM.Resolve(req.ShortCode)
	return &pb.ResolveResponse{LongUrl: url, Found: found}, nil
}