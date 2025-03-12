package favolotto

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Store struct {
	inNfc     <-chan string
	inFname   chan<- string
	storePath string
}

type Metadata struct {
	NfcTag string `json:"nfc_tag"`
	Name   string `json:"name"`
}

func NewStore(storePath string, inNfc <-chan string, inFname chan<- string) *Store {
	return &Store{
		inNfc:     inNfc,
		inFname:   inFname,
		storePath: storePath,
	}
}

func (s *Store) Run(ctx context.Context) {
	var meta []Metadata
	f, err := os.ReadFile(filepath.Join(s.storePath, "metadata.json"))
	if err != nil {
		log.Fatalf("Error reading metadata file: %v", err)
	}
	if err := json.Unmarshal(f, &meta); err != nil {
		log.Fatalf("Error unmarshalling metadata file: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case nfc := <-s.inNfc:
			for _, m := range meta {
				if m.NfcTag == nfc {
					s.inFname <- filepath.Join(s.storePath, m.Name)
					break
				}
			}
		}
	}
}
