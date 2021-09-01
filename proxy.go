package websocket_proxy

import (
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"net/http"
	"net/url"
	"os"
	"time"
)

const Timeout = time.Hour * 24 * 7

// WebsocketProxy is the generic interface for a proxy implementation
type WebsocketProxy interface {
	Dial() (*websocket.Conn, error)
	Handler(w http.ResponseWriter, r *http.Request)
	Close()
	Wait(<-chan os.Signal)
}

type connection struct {
	client *websocket.Conn
	server *websocket.Conn
}

type websocketProxy struct {
	URL      *url.URL
	Header   http.Header
	Upgrader *websocket.Upgrader

	auth        Auth
	keyManager  KeyManager
	connections []*connection
}

// NewSimpleProxy returns a configured proxy instance from just a url
func NewSimpleProxy(url *url.URL) (WebsocketProxy, error) {
	return NewProxy(url, nil, nil, nil, nil)
}

// NewProxy returns a configured WebsocketProxy instance and fetches keys if required
func NewProxy(
	url *url.URL,
	header http.Header,
	origins []string,
	auth Auth,
	keyManager KeyManager,
) (WebsocketProxy, error) {
	wsp := websocketProxy{
		URL:    url,
		Header: header,
		Upgrader: &websocket.Upgrader{
			CheckOrigin: checkOrigin(origins),
		},
		auth:       auth,
		keyManager: keyManager,
	}
	if wsp.auth != nil && wsp.keyManager != nil {
		if err := wsp.keyManager.FetchKeys(); err != nil {
			return wsp, fmt.Errorf("error fetching keys: %v", err)
		}
	}
	return wsp, nil
}

// Dial connects to the Websocket backend and returns an error if failing
func (wp websocketProxy) Dial() (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(wp.URL.String(), wp.Header)
	if err != nil {
		return nil, err
	}
	err = c.SetReadDeadline(time.Now().Add(Timeout))
	if err != nil {
		return nil, err
	}
	return c, c.SetWriteDeadline(time.Now().Add(Timeout))
}

// Handler for an in-built http server. It authenticates the user if required,
// dials the backend and then stores the connection between client & server,
// relaying any messages sent by either client and server
func (wp websocketProxy) Handler(w http.ResponseWriter, r *http.Request) {
	if wp.auth != nil {
		if ok := wp.auth.Authenticate(r, wp.keyManager); !ok {
			http.Error(w, "Invalid key", 401)
			return
		}
	}
	cc, err := wp.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	sc, err := wp.Dial()
	if err != nil {
		return
	}
	conn := &connection{cc, sc}
	defer wp.close(conn)
	log.WithField("IP", conn.client.RemoteAddr()).Info("New connection")

	g := errgroup.Group{}
	g.Go(func() error {
		return wp.read(w, conn)
	})
	g.Go(func() error {
		return wp.write(w, conn)
	})
	if err := g.Wait(); err != nil {
		log.Errorf("Error during handling websocket connection: %v", err)
	}
}

// Close will disconnect all the active connections between client and server
func (wp websocketProxy) Close() {
	for _, c := range wp.connections {
		wp.close(c)
	}
}

// Wait listens to any interrupt signals and then closes all connections if one is received
func (wp websocketProxy) Wait(interrupt <-chan os.Signal) {
	<-interrupt
	wp.Close()
}

func (wp websocketProxy) write(_ http.ResponseWriter, conn *connection) error {
	for {
		t, msg, err := conn.client.ReadMessage()
		if err != nil {
			return err
		}
		err = conn.server.WriteMessage(t, msg)
		if err != nil {
			return err
		}
		log.WithField("IP", conn.client.RemoteAddr()).
			WithField("Message", string(msg)).
			Debug("Written message to server")
	}
}

func (wp websocketProxy) read(_ http.ResponseWriter, conn *connection) error {
	for {
		t, msg, err := conn.server.ReadMessage()
		if err != nil {
			return err
		}
		err = conn.client.WriteMessage(t, msg)
		if err != nil {
			return err
		}
		log.WithField("IP", conn.client.RemoteAddr()).
			WithField("Message", string(msg)).
			Debug("Read message from server")
	}
}

func (wp websocketProxy) close(conn *connection) {
	_ = conn.client.Close()
	_ = conn.server.Close()
	log.WithField("IP", conn.client.RemoteAddr()).Info("Closed connection")
}

func checkOrigin(origins []string) func(*http.Request) bool {
	return func(r *http.Request) bool {
		if len(origins) == 0 {
			return true
		}
		co := r.Header.Get("Origin")
		for _, o := range origins {
			if o == co {
				return true
			}
		}
		return false
	}
}
