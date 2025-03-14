package favolotto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	inNfc     <-chan string
	inFname   chan<- string
	storePath string

	mu sync.Mutex
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
	meta := s.loadMetadata()

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

func (s *Store) putSong(nfcTag, fname string, file multipart.File) error {
	filePath := filepath.Join(s.storePath, fname)
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	// update metadata
	meta := s.loadMetadata()
	// check if nfcTag already exists and replace it
	found := false
	for idx, m := range meta {
		if m.NfcTag == nfcTag {
			// remove old file
			if err := os.Remove(filepath.Join(s.storePath, meta[idx].Name)); err != nil {
				return fmt.Errorf("error removing old file: %w", err)
			}
			// update metadata
			meta[idx].Name = fname
			found = true
			break
		}
	}
	if !found {
		meta = append(meta, Metadata{NfcTag: nfcTag, Name: fname})
	}

	s.mu.Lock()
	f, err := os.Create(filepath.Join(s.storePath, "metadata.json"))
	if err != nil {
		return fmt.Errorf("error creating metadata file: %w", err)
	}

	if err := json.NewEncoder(f).Encode(meta); err != nil {
		return fmt.Errorf("error encoding metadata file: %w", err)
	}
	f.Close()
	s.mu.Unlock()

	return nil
}

func (s *Store) deleteSong(nfcTag string) error {
	meta := s.loadMetadata()
	// check if nfcTag exists
	found := false
	for idx, m := range meta {
		if m.NfcTag == nfcTag {
			// remove file
			if err := os.Remove(filepath.Join(s.storePath, meta[idx].Name)); err != nil {
				return fmt.Errorf("error removing file: %w", err)
			}
			// update metadata
			meta = append(meta[:idx], meta[idx+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("nfcTag not found")
	}

	s.mu.Lock()
	f, err := os.Create(filepath.Join(s.storePath, "metadata.json"))
	if err != nil {
		return fmt.Errorf("error creating metadata file: %w", err)
	}

	if err := json.NewEncoder(f).Encode(meta); err != nil {
		return fmt.Errorf("error encoding metadata file: %w", err)
	}
	f.Close()
	s.mu.Unlock()

	return nil
}

func (s *Store) loadMetadata() []Metadata {
	var meta []Metadata
	f, err := os.ReadFile(filepath.Join(s.storePath, "metadata.json"))
	if err != nil {
		log.Fatalf("Error reading metadata file: %v", err)
	}
	if err := json.Unmarshal(f, &meta); err != nil {
		log.Fatalf("Error unmarshalling metadata file: %v", err)
	}

	return meta
}
