package sshkey

// Package sshkey generates and loads OpenSSH key pairs used to access
// provisioned infrastructure.

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// KeyPair points at an on-disk private key and holds its OpenSSH public key.
type KeyPair struct {
	Name           string
	PrivateKeyPath string
	PublicKeyPath  string
	PublicKey      string
}

// Ensure returns the key pair named name inside dir, generating it if missing.
// Generation is idempotent: an existing pair is loaded and reused.
func Ensure(dir, name string) (KeyPair, error) {
	if strings.TrimSpace(name) == "" {
		return KeyPair{}, errors.New("sshkey: name is required")
	}

	pair := KeyPair{
		Name:           name,
		PrivateKeyPath: filepath.Join(dir, name),
		PublicKeyPath:  filepath.Join(dir, name+".pub"),
	}

	if pub, err := os.ReadFile(pair.PublicKeyPath); err == nil {
		if _, statErr := os.Stat(pair.PrivateKeyPath); statErr == nil {
			pair.PublicKey = strings.TrimSpace(string(pub))
			return pair, nil
		}
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return KeyPair{}, fmt.Errorf("sshkey: create key dir: %w", err)
	}

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyPair{}, fmt.Errorf("sshkey: generate key: %w", err)
	}

	privateBlock, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return KeyPair{}, fmt.Errorf("sshkey: marshal private key: %w", err)
	}

	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return KeyPair{}, fmt.Errorf("sshkey: marshal public key: %w", err)
	}
	authorizedKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPublicKey)))

	if err := os.WriteFile(pair.PrivateKeyPath, pem.EncodeToMemory(privateBlock), 0o600); err != nil {
		return KeyPair{}, fmt.Errorf("sshkey: write private key: %w", err)
	}
	if err := os.WriteFile(pair.PublicKeyPath, []byte(authorizedKey+"\n"), 0o644); err != nil {
		return KeyPair{}, fmt.Errorf("sshkey: write public key: %w", err)
	}

	pair.PublicKey = authorizedKey
	return pair, nil
}

// Remove deletes the key pair files, ignoring missing files.
func Remove(dir, name string) error {
	for _, path := range []string{filepath.Join(dir, name), filepath.Join(dir, name+".pub")} {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}
