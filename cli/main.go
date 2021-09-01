package main

import (
	"fmt"
	proxy "github.com/jleeh/websocket-proxy"
	"github.com/jleeh/websocket-proxy/config"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"os/signal"
)

func main() {
	log.Println("Starting the WebSocket authenticating proxy")

	c := config.New("config", configDefaults())
	u, err := url.Parse(c.Server)
	if err != nil {
		log.Fatalf("Error parsing server: %v", err)
	}
	log.SetLevel(c.LogLevel)

	wp, err := proxy.NewProxy(
		u,
		nil,
		c.AllowedOrigins,
		proxy.NewAuth(c.AuthType),
		proxy.NewKeyManager(c.KeyManagerType, c.KeyIdentifier),
	)
	if err != nil {
		log.Fatalf("Error creating new proxy: %v", err)
	}

	if conn, err := wp.Dial(); err != nil {
		log.Fatalf("Error dialing the server: %v", err)
	} else {
		log.WithField("url", c.Server).Println("Successfully dialed the server")
		_ = conn.Close()
	}

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", wp.Handler)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", c.Port), mux); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()
	log.WithField("port", c.Port).Println("Server started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	wp.Wait(sig)

	log.Println("Server stopped")
}

func configDefaults() map[string]interface{} {
	return map[string]interface{}{
		"port":             8080,
		"server":           "ws://localhost:3000",
		"auth_type":        "",
		"key_manager_type": "",
		"key_identifier":   "",
		"allowed_origins":  []string{},
		"log_level":        "5",
	}
}
