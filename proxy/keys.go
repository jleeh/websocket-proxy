package proxy

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"io/ioutil"
	"os"
)

const (
	KeyManagerFile           = "file"
	KeyManagerSecretsManager = "aws_sm"
)

// Keys type is the representation of how keys are stored in the manager
type Keys []string

// KeyManager is a generic interface for implementing a specific backend
// for managing API keys
type KeyManager interface {
	ValidateKey(string) bool
	FetchKeys() error
	setIdentifier(string)
}

// NewKeyManager returns a pointer of an KeyManager implementation based
// on the type of KeyManager that was provided
func NewKeyManager(keyManagerType string, id string) KeyManager {
	var k KeyManager
	switch keyManagerType {
	case KeyManagerFile:
		k = &File{}
		break
	case KeyManagerSecretsManager:
		k = &SecretsManager{}
	}
	if k != nil {
		k.setIdentifier(id)
	}
	return k
}

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

// SecretsManager is the KeyManager implementation for AWS Secrets Manager
type SecretsManager struct {
	id string
	keys []string
}

// ValidateKey returns a boolean to whether a key given is present in the manager
func (sm *SecretsManager) ValidateKey(key string) bool {
	for _, k := range sm.keys {
		if k == key {
			return true
		}
	}
	return false
}

// FetchKeys sets the keys from the manager backend into memory for use on client auth
func (sm *SecretsManager) FetchKeys() error {
	svc := secretsmanager.New(session.Must(session.NewSession(aws.NewConfig())))
	if svo, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(sm.id),
	}); err != nil {
		return err
	} else {
		b := []byte(*svo.SecretString)
		if err := json.Unmarshal(b, &sm.keys); err != nil {
			return err
		}
	}
	return nil
}

func (sm *SecretsManager) setIdentifier(id string) {
	sm.id = id
}