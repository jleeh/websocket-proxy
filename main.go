package main

import (
	"fmt"
	"github.com/jleeh/ws-auth-proxy/config"
	"github.com/jleeh/ws-auth-proxy/proxy"
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
	wp, err := proxy.NewProxy(u, nil, proxy.NewAuth(c.AuthType), proxy.NewKeyManager(c.KeyManagerType))
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
		http.HandleFunc("/", wp.Handler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil))
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
	}
}
