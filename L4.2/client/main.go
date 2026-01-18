package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	pb "grpc-grep/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type result struct {
	rank int
	resp *pb.GrepResponse
	err  error
}

func main() {
	after := flag.Int("A", 0, "Print N lines after match")
	before := flag.Int("B", 0, "Print N lines before match")
	countOnly := flag.Bool("c", false, "Count matches only")
	ignore := flag.Bool("i", false, "Ignore case")
	invert := flag.Bool("v", false, "Invert match")
	fixed := flag.Bool("F", false, "Fixed strings (no regex)")
	lineNum := flag.Bool("n", false, "Show line numbers")
	serversFlag := flag.String("servers", "localhost:50053", "Comma-separated list of server addresses")

	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: client [flags] pattern file")
		flag.PrintDefaults()
		os.Exit(1)
	}

	pattern := args[0]
	filePath := args[1]

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	serverAddrs := strings.Split(*serversFlag, ",")
	numServers := len(serverAddrs)
	if numServers == 0 {
		log.Fatal("no servers specified")
	}

	chunks := splitToChunks(lines, numServers)

	var wg sync.WaitGroup
	results := make(chan result, numServers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	for i, addr := range serverAddrs {
		wg.Add(1)
		currentOffset := 0
		for k := range i {
			currentOffset += len(chunks[k])
		}

		go func(index int, address string, chunk []string, offset int) {
			defer wg.Done()

			if len(chunk) == 0 {
				results <- result{rank: index, err: nil, resp: &pb.GrepResponse{}}
				return
			}

			conn, err := grpc.NewClient(
				address,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				log.Printf("failed to connect to %s: %v", address, err)
				results <- result{err: err}
				return
			}
			defer conn.Close()

			client := pb.NewGrepServiceClient(conn)

			resp, err := client.Grep(ctx, &pb.GrepRequest{
				Lines:      chunk,
				Pattern:    pattern,
				After:      int32(*after),
				Before:     int32(*before),
				CountOnly:  *countOnly,
				Ignore:     *ignore,
				Invert:     *invert,
				Fixed:      *fixed,
				LineNum:    *lineNum,
				LineOffset: int32(offset),
			})
			results <- result{rank: i, resp: resp, err: err}
		}(i, addr, chunks[i], currentOffset)
	}

	// Кворум (N/2 + 1)
	quorum := (numServers / 2) + 1
	success := 0
	failed := 0
	collected := make([]*pb.GrepResponse, numServers) // Fixed size for ordering

	// Сбор результатов по кворуму
	responsesReceived := 0
	for responsesReceived < numServers {
		res := <-results
		responsesReceived++

		if res.err != nil {
			log.Printf("server error: %v", res.err)
			failed++
		} else {
			success++
			collected[res.rank] = res.resp
		}

	}

	if success < quorum {
		log.Fatalf("Quorum not reached: success=%d, failed=%d, quorum=%d", success, failed, quorum)
	}

	if *countOnly {
		total := 0
		for _, resp := range collected {
			if resp == nil {
				continue
			}
			total += int(resp.Count)
		}
		fmt.Printf("Total Count (Quorum %d/%d): %d\n", success, numServers, total)
	} else {
		for _, resp := range collected {
			if resp == nil {
				continue
			}
			for _, line := range resp.Output {
				fmt.Println(line)
			}
		}
	}
}

func splitToChunks(lines []string, n int) [][]string {
	if n <= 0 {
		return nil
	}
	total := len(lines)
	if total == 0 {
		return make([][]string, n)
	}

	chunkSize := (total + n - 1) / n
	chunks := make([][]string, n)

	for i := 0; i < n; i++ {
		start := i * chunkSize
		if start >= total {
			break
		}
		end := min(start + chunkSize, total)
		chunks[i] = lines[start:end]
	}
	return chunks
}
