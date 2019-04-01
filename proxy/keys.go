package proxy

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

const (
	KeyManagerSecretsManager = "aws_sm"
)

// Keys type is the representation of how keys are stored in the manager
type Keys []string

// KeyManager is a generic interface for implementing a specific backend
// for managing API keys
type KeyManager interface {
	ValidateKey(string) bool
	FetchKeys() error
}

// NewKeyManager returns a pointer of an KeyManager implementation based
// on the type of KeyManager that was provided
func NewKeyManager(keyManagerType string) KeyManager {
	var k KeyManager
	switch keyManagerType {
	case KeyManagerSecretsManager:
		k = &SecretsManager{}
	}
	return k
}

// SecretsManager is the KeyManager implementation for AWS Secrets Manager
type SecretsManager struct {
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
		SecretId: aws.String("wap-keys"),
	}); err != nil {
		return err
	} else {
		var keys []string
		b := []byte(*svo.SecretString)
		if err := json.Unmarshal(b, &keys); err != nil {
			return err
		}
		sm.keys = keys
	}
	return nil
}
