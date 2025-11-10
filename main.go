package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
)

func main() {
	fmt.Println("Listening on port :6379")

	// Create a new server
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Load dbfile
	aof, err := NewAof("db.aof")
	if err != nil {
		fmt.Println(err)
	}
	defer aof.Close()

	aof.Read(func(value Value) {
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			return
		}

		handler(args)
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var wg sync.WaitGroup
	var conns sync.Map

	go func() {
		<-ctx.Done()
		fmt.Println("Shutting down...")
		l.Close()

		conns.Range(func(key, value any) bool {
			conn := value.(net.Conn)
			fmt.Println("Closing connection: ", conn.RemoteAddr().String())
			conn.Close()

			return true
		})
	}()

	// Listen for connections
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				wg.Wait()
				fmt.Println("Shutdown complete.")
				return
			default:
				fmt.Println("Accept error: ", err)
				continue
			}
		}
		fmt.Println("Client connected: ", conn.RemoteAddr().String())
		conns.Store(conn.RemoteAddr().String(), conn)
		wg.Add(1)

		go func(conn net.Conn) {
			defer wg.Done()
			defer conns.Delete(conn.RemoteAddr().String())
			defer func() {
				_ = conn.Close()
				fmt.Println("Connection closed: ", conn.RemoteAddr().String())
			}()
			defer fmt.Println("Client disconnected: ", conn.RemoteAddr().String())

			resp := NewResp(conn)
			writer := NewWriter(conn)

			for {
				value, err := resp.Read()

				if err != nil {
					if errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection") {
						// graceful shutdown in progress, just exit silently
						return
					}
					if err != io.EOF {
						fmt.Println("Read error: ", err)
					}
					return
				}

				fmt.Println("Command received: ", value.String())

				if value.typ != "array" {
					fmt.Println("Invalid request, expected array")
					continue
				}

				if len(value.array) == 0 {
					fmt.Println("Invalid request, expected array length > 0")
					continue
				}

				command := strings.ToUpper(value.array[0].bulk)
				args := value.array[1:]

				handler, ok := Handlers[command]
				if !ok {
					fmt.Println("Invalid command: ", command)
					_ = writer.EmptyWrite()
					continue
				}

				if command == "SET" || command == "HSET" {
					_ = aof.Write(value)
				}

				result := handler(args)
				_ = writer.Write(result)
			}
		}(conn)
	}
}
