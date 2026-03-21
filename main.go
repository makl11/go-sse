package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"math"
	"net/http"
	"os"
	"time"
)

//go:embed client/*
var clientData embed.FS

func getHtmlRoot(clientFs fs.FS) fs.FS {
	htmlRoot, err := fs.Sub(clientFs, "client")
	if err != nil {
		panic(err)
	}
	return htmlRoot
}

func main() {
	host := flag.String("host", "localhost", "Host the server listens on")
	port := flag.Int("port", 6969, "Port the server listens on")
	flag.Parse()

	http.HandleFunc("/sse", handleSse)
	http.Handle("/", http.FileServerFS(getHtmlRoot(clientData)))

	addr := fmt.Sprintf("%s:%d", *host, *port)
	slog.Info("Listening", "host", *host, "port", *port, "addr", fmt.Sprintf("http://%s", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("Failed to start server", "host", *host, "port", *port, "err", err)
		os.Exit(1)
	}
}

func handleSse(w http.ResponseWriter, r *http.Request) {
	rc := http.NewResponseController(w)
	w.Header().Add("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)

	time.Sleep(1 * time.Second)
	counter := 0
	for {
		counter++
		data := fmt.Sprintf("data: This is event number %d\r\n\r\n", counter)
		_, err := w.Write([]byte(data))
		if err != nil {
			slog.Error("Failed to write event", "err", err)
			return
		}
		err = rc.Flush()
		if err != nil {
			slog.Error("Failed to flush event", "err", err)
			return
		}
		if counter >= math.MaxInt {
			counter = 0
		}

		time.Sleep(1 * time.Second)
	}
}
