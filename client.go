package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "distributed-url-shortener/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewURLShortenerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Shorten a URL
	resp, err := client.Shorten(ctx, &pb.ShortenRequest{LongUrl: "https://google.com"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("🔗 Shortened: https://google.com -> Code: %s\n", resp.ShortCode)

	// 2. Resolve it back
	resolveResp, err := client.Resolve(ctx, &pb.ResolveRequest{ShortCode: resp.ShortCode})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("🔍 Resolved: Code %s -> URL: %s\n", resp.ShortCode, resolveResp.LongUrl)
}