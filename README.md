# WebSocket Proxy
Lightweight WebSocket proxy with API key management built on top of 
[gorilla/websockets](https://github.com/gorilla/websocket). This can either be imported for use in your 
own implementations with the http.Handler interface, or installed directly to be used as a CLI.

Supported authentication methods:

- Headers (X-API-KEY)
- Param (GET Param `apikey`)

Supported key management tools:

- Local file
- AWS Secrets Manager

## Install & Build

```
go get github.com/jleeh/websocket-proxy && go build -o $GOPATH/bin/websocket-proxy
```

## Running the Proxy
You can run the WebSocket proxy by just running the built binary:
```
./websocket-proxy
```

## Configure

Configuration for the CLI is via [Viper](https://github.com/spf13/viper), so configuration variables can be passed in
via environment variables or file.

Example configuration:
```yaml
port: 8080
server: ws://localhost:3000
auth_type: header
key_manager_type: file
key_identifier: /home/me/keys.json
allowed_origins:
  - localhost
  - google.com
```

Example .env file:
```dotenv
PORT=80
SERVER=ws://localhost:3000
AUTH_TYPE=header
KEY_MANAGER_TYPE=file
KEY_IDENTIFIER=/home/me/keys.json
ALLOWED_ORIGINS=localhost,google.com
```

### Using AWS Secrets Manager
Change the `key_manager_type` configuration variable to `aws_sm` and then change the `key_identifier` to the name 
of the secret in AWS.

## Custom Example

```
go get github.com/jleeh/websocket-proxy/proxy
```

```go
package main

import (
	"github.com/jleeh/websocket-proxy/proxy"
	"log"
	"net/http"
	"net/url"
)

func main() {
	u, _ := url.Parse("http://localhost:3000")
	wp, _ := proxy.NewProxy(u, nil, nil, nil, nil)
	
	http.HandleFunc("/", wp.Handler)
	log.Fatal(http.ListenAndServe(":80", nil))
}
```

### Extending Authentication

Interface:
```go
type Auth interface {
	Authenticate(*http.Request, KeyManager) bool
}
```

Implementation example (header):
```go
// Header is the Auth implementation that requires `X-API-KEY` to be set
// in the request headers
type Header struct{}

// Authenticate takes the `X-API-KEY` in the request headers and then authenticates
// it with the KeyManager
func (p *Header) Authenticate(r *http.Request, km KeyManager) bool {
	key := r.Header.Get("X-API-KEY")
	return km.ValidateKey(key)
}
```

### Extending KeyManager

Interface:
```go
type KeyManager interface {
	ValidateKey(string) bool
	FetchKeys() error
	setIdentifier(string)
}
```

Implementation example (file):
```go
// File manages keys on the local disk
type File struct {
	id string
	keys []string
}

// ValidateKey returns a boolean to whether a key given is present in the file
func (f *File) ValidateKey(key string) bool {
	for _, k := range f.keys {
		if k == key {
			return true
		}
	}
	return false
}

// FetchKeys sets the keys from the file on local disk
func (f *File) FetchKeys() error {
	if file, err := os.Open(f.id); err != nil {
		return err
	} else if b, err := ioutil.ReadAll(file); err != nil {
		return err
	} else if err := json.Unmarshal(b, &f.keys); err != nil {
		return err
	}
	return nil
}

func (f *File) setIdentifier(id string) {
	f.id = id
}
```

You can create your own implementation of a KeyManager type to then pass into the `proxy.NewProxy(...)` method.