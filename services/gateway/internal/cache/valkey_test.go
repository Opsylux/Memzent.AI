package cache

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

type mockValkeyServer struct {
	ln          net.Listener
	responses   map[string]string
	defaultResp string
}

func startMockValkeyServer(t *testing.T) *mockValkeyServer {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock valkey server: %v", err)
	}

	server := &mockValkeyServer{
		ln:          ln,
		responses:   make(map[string]string),
		defaultResp: "+OK\r\n",
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // listener closed
			}
			go server.handleConnection(conn)
		}
	}()

	return server
}

func (s *mockValkeyServer) Close() {
	if s.ln != nil {
		s.ln.Close()
	}
}

func (s *mockValkeyServer) AddResponse(commandPart, response string) {
	s.responses[commandPart] = response
}

func (s *mockValkeyServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		if !strings.HasPrefix(line, "*") {
			continue
		}

		var arrayLen int
		_, err = fmt.Sscanf(line, "*%d\r\n", &arrayLen)
		if err != nil {
			return
		}

		var args []string
		for i := 0; i < arrayLen; i++ {
			lenLine, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			var argLen int
			_, err = fmt.Sscanf(lenLine, "$%d\r\n", &argLen)
			if err != nil {
				return
			}

			argBytes := make([]byte, argLen+2)
			_, err = io.ReadFull(reader, argBytes)
			if err != nil {
				return
			}
			args = append(args, string(argBytes[:argLen]))
		}

		fullCmd := strings.Join(args, " ")
		upperCmd := strings.ToUpper(fullCmd)

		response := s.defaultResp
		matched := false
		for k, v := range s.responses {
			if strings.Contains(upperCmd, strings.ToUpper(k)) {
				response = v
				matched = true
				break
			}
		}

		if !matched {
			if strings.HasPrefix(upperCmd, "HELLO") {
				response = "%3\r\n$6\r\nserver\r\n$6\r\nvalkey\r\n$5\r\nproto\r\n:3\r\n$4\r\nmode\r\n$10\r\nstandalone\r\n"
			} else if strings.HasPrefix(upperCmd, "CLUSTER") {
				response = "-ERR This instance has cluster support disabled\r\n"
			} else if strings.HasPrefix(upperCmd, "PING") {
				response = "+PONG\r\n"
			} else if strings.HasPrefix(upperCmd, "SET") {
				response = "+OK\r\n"
			} else if strings.HasPrefix(upperCmd, "GET") {
				response = "$-1\r\n"
			}
		}

		_, _ = conn.Write([]byte(response))
	}
}

func TestMemzentCache_Ping_Success(t *testing.T) {
	server := startMockValkeyServer(t)
	defer server.Close()

	ctx := context.Background()
	cache, err := NewMemzentCache(ctx, server.ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to create cache client: %v", err)
	}
	defer cache.Close()

	err = cache.Ping(ctx)
	if err != nil {
		t.Errorf("expected no error on Ping, got: %v", err)
	}
}

func TestMemzentCache_Ping_Error(t *testing.T) {
	server := startMockValkeyServer(t)
	defer server.Close()

	server.AddResponse("PING", "-ERR rate limit exceeded\r\n")

	ctx := context.Background()
	cache, err := NewMemzentCache(ctx, server.ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to create cache client: %v", err)
	}
	defer cache.Close()

	err = cache.Ping(ctx)
	if err == nil {
		t.Errorf("expected error on Ping, got nil")
	} else if !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMemzentCache_GetSemanticResult_Hit(t *testing.T) {
	server := startMockValkeyServer(t)
	defer server.Close()

	server.AddResponse("GET testkey", "$12\r\ncached_value\r\n")

	ctx := context.Background()
	cache, err := NewMemzentCache(ctx, server.ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to create cache client: %v", err)
	}
	defer cache.Close()

	val, err := cache.GetSemanticResult(ctx, "testkey")
	if err != nil {
		t.Errorf("expected no error on Get, got: %v", err)
	}
	if val != "cached_value" {
		t.Errorf("expected cached_value, got: %q", val)
	}
}

func TestMemzentCache_GetSemanticResult_Miss(t *testing.T) {
	server := startMockValkeyServer(t)
	defer server.Close()

	// Default response for GET is nil ($-1\r\n)
	ctx := context.Background()
	cache, err := NewMemzentCache(ctx, server.ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to create cache client: %v", err)
	}
	defer cache.Close()

	val, err := cache.GetSemanticResult(ctx, "missingkey")
	if err != nil {
		t.Errorf("expected no error on miss, got: %v", err)
	}
	if val != "" {
		t.Errorf("expected empty string on miss, got: %q", val)
	}
}

func TestMemzentCache_GetSemanticResult_Error(t *testing.T) {
	server := startMockValkeyServer(t)
	defer server.Close()

	server.AddResponse("GET errorkey", "-ERR database corruption\r\n")

	ctx := context.Background()
	cache, err := NewMemzentCache(ctx, server.ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to create cache client: %v", err)
	}
	defer cache.Close()

	_, err = cache.GetSemanticResult(ctx, "errorkey")
	if err == nil {
		t.Errorf("expected error on GET error key, got nil")
	} else if !strings.Contains(err.Error(), "database corruption") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMemzentCache_SetResult(t *testing.T) {
	server := startMockValkeyServer(t)
	defer server.Close()

	ctx := context.Background()
	cache, err := NewMemzentCache(ctx, server.ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to create cache client: %v", err)
	}
	defer cache.Close()

	err = cache.SetResult(ctx, "newkey", "newval", 10*time.Second)
	if err != nil {
		t.Errorf("expected no error on SetResult, got: %v", err)
	}
}

func TestMemzentCache_ConnectError(t *testing.T) {
	ctx := context.Background()
	// Use an invalid port/address where nothing is listening
	_, err := NewMemzentCache(ctx, "127.0.0.1:99999")
	if err == nil {
		t.Errorf("expected connection error, got nil")
	}
}
