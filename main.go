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

	"github.com/makl11/go-sse/channel_mux"
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

type Event string

func (event *Event) Marshall() ([]byte, error) {
	return []byte(*event), nil
}

func createProducer() chan Event {
	outChan := make(chan Event)
	counter := 0
	go func() {
		for {
			counter++
			data := fmt.Sprintf("data: This is event number %d\r\n\r\n", counter)
			outChan <- Event(data)
			if counter >= math.MaxInt {
				counter = 0
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return outChan
}

func main() {
	host := flag.String("host", "localhost", "Host the server listens on")
	port := flag.Int("port", 6969, "Port the server listens on")
	debug := flag.Bool("debug", false, "enable debug output")
	flag.Parse()

	if *debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	producer := channel_mux.NewChannelMux(createProducer())
	http.HandleFunc("/sse", makeHandleSse(producer))
	http.Handle("/", http.FileServerFS(getHtmlRoot(clientData)))

	addr := fmt.Sprintf("%s:%d", *host, *port)
	slog.Info("Listening", "host", *host, "port", *port, "addr", fmt.Sprintf("http://%s", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("Failed to start server", "host", *host, "port", *port, "err", err)
		os.Exit(1)
	}
}
func makeHandleSse(channelMux channel_mux.ChannelMux[Event]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		eventChannel := channelMux.NewOut()
		defer close(eventChannel)
		defer channelMux.RemoveOut(eventChannel)
		rc := http.NewResponseController(w)
		w.Header().Add("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		for event := range eventChannel {
			select {
			case <-r.Context().Done():
				return
			default:
				data, err := event.Marshall()
				if err != nil {
					slog.Error("Failed to marshall event", "event", event, "err", err)
					return
				}
				_, err = w.Write(data)
				if err != nil {
					slog.Error("Failed to write event", "event", event, "err", err)
					return
				}
				err = rc.Flush()
				if err != nil {
					slog.Error("Failed to flush event", "event", event, "err", err)
					return
				}
			}
		}
	}
}
