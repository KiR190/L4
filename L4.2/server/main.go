package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"regexp"
	"runtime"
	"strings"
	"sync"

	pb "grpc-grep/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedGrepServiceServer
}

type Options struct {
	after      int
	before     int
	countOnly  bool
	ignore     bool
	invert     bool
	fixed      bool
	lineNum    bool
	lineOffset int
}

// compilePattern подготавливает функцию проверки строки
func compilePattern(pattern string, opts Options) (func(string) bool, error) {
	if opts.ignore {
		pattern = "(?i)" + pattern
	}
	if opts.fixed {
		if opts.ignore {
			pattern = strings.ToLower(pattern)
			return func(s string) bool {
				return strings.Contains(strings.ToLower(s), pattern)
			}, nil
		}
		return func(s string) bool {
			return strings.Contains(s, pattern)
		}, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return func(s string) bool {
		return re.MatchString(s)
	}, nil
}

func GrepLines(
	lines []string,
	pattern string,
	opts Options,
) ([]string, int, error) {
	matchFunc, err := compilePattern(pattern, opts)
	if err != nil {
		return nil, 0, err
	}

	numLines := len(lines)
	matched := make([]bool, numLines)

	// Параллельная обработка строк
	numWorkers := runtime.NumCPU()
	if numWorkers > numLines {
		numWorkers = numLines
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	jobs := make(chan int, numLines)
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				ok := matchFunc(lines[i])
				if opts.invert {
					ok = !ok
				}
				matched[i] = ok
			}
		}()
	}

	for i := 0; i < numLines; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	if opts.countOnly {
		count := 0
		for _, m := range matched {
			if m {
				count++
			}
		}
		return nil, count, nil
	}

	var out []string
	printed := make(map[int]bool)

	for i := range lines {
		if !matched[i] {
			continue
		}

		start := i - opts.before
		if start < 0 {
			start = 0
		}
		end := i + opts.after
		if end >= len(lines) {
			end = len(lines) - 1
		}

		for j := start; j <= end; j++ {
			if printed[j] {
				continue
			}
			printed[j] = true

			if opts.lineNum {
				out = append(out, fmt.Sprintf("%d:%s", opts.lineOffset+j+1, lines[j]))
			} else {
				out = append(out, lines[j])
			}
		}
	}

	return out, 0, nil
}

func (s *server) Grep(
	ctx context.Context,
	req *pb.GrepRequest,
) (*pb.GrepResponse, error) {
	opts := Options{
		after:      int(req.After),
		before:     int(req.Before),
		countOnly:  req.CountOnly,
		ignore:     req.Ignore,
		invert:     req.Invert,
		fixed:      req.Fixed,
		lineNum:    req.LineNum,
		lineOffset: int(req.LineOffset),
	}

	out, count, err := GrepLines(
		req.Lines,
		req.Pattern,
		opts,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.GrepResponse{
		Output: out,
		Count:  int32(count),
	}, nil
}

func main() {
	port := flag.Int("port", 50053, "gRPC server port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGrepServiceServer(grpcServer, &server{})

	log.Printf("gRPC server listening on :%d", *port)
	log.Fatal(grpcServer.Serve(lis))
}
