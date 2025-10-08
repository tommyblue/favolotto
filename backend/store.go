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
	"time"
)

type Store struct {
	inNfc     <-chan string
	inFname   chan<- string
	storePath string
	data      []Metadata
	lastNfc   string

	mu sync.Mutex
}

type Metadata struct {
	NfcTag string `json:"nfc_tag"`
	Name   string `json:"name"`
}

func NewStore(storePath string, inNfc <-chan string, inFname chan<- string) (*Store, error) {
	store := &Store{
		inNfc:     inNfc,
		inFname:   inFname,
		storePath: storePath,
	}

	_, err := os.ReadFile(filepath.Join(storePath, "metadata.json"))
	if err != nil {
		log.Printf("metadata file not found, creating new one")
		store.storeMetadata([]Metadata{})
	}

	store.refreshMetadata()

	return store, nil
}

func (s *Store) Run(ctx context.Context) {
	resetTime := 5 * time.Second
	resetTicker := time.NewTicker(resetTime)
	defer resetTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("store context done")
			return
		case <-resetTicker.C:
			if s.lastNfc != "" {
				log.Printf("resetting last NFC tag")
				s.lastNfc = ""
				// stop audio playback
				s.inFname <- ""
			}
		case nfc := <-s.inNfc:
			resetTicker.Reset(resetTime)
			if s.lastNfc == nfc {
				continue
			}
			s.lastNfc = nfc
			for _, m := range s.data {
				if m.NfcTag == nfc {
					log.Printf("NFC tag %s found, playing %s", nfc, m.Name)
					s.inFname <- filepath.Join(s.storePath, m.Name)
					break
				}
			}
		}
	}
}

func (s *Store) LastNfc() string {
	return s.lastNfc
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

	// check if nfcTag already exists and replace it
	found := false
	meta := s.data
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

	if err := s.storeMetadata(meta); err != nil {
		return fmt.Errorf("error storing metadata: %w", err)
	}

	return nil
}

func (s *Store) deleteSong(nfcTag string) error {
	// check if nfcTag exists
	found := false
	meta := s.data
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

	if err := s.storeMetadata(meta); err != nil {
		return fmt.Errorf("error storing metadata: %w", err)
	}

	return nil
}

func (s *Store) getMetadata() []Metadata {
	return s.data
}

func (s *Store) storeMetadata(data []Metadata) error {
	s.mu.Lock()
	f, err := os.CreateTemp(os.TempDir(), "metadata.json")
	if err != nil {
		log.Printf("error creating metadata file: %v", err)
		return fmt.Errorf("error creating metadata file: %w", err)
	}

	if err := json.NewEncoder(f).Encode(data); err != nil {
		log.Printf("error encoding metadata file: %v", err)
		return fmt.Errorf("error encoding metadata file: %w", err)
	}
	f.Close()
	if err := os.Rename(f.Name(), filepath.Join(s.storePath, "metadata.json")); err != nil {
		log.Printf("error renaming metadata file: %v", err)
		return fmt.Errorf("error renaming metadata file: %w", err)
	}
	s.mu.Unlock()

	s.refreshMetadata()

	return nil
}

func (s *Store) refreshMetadata() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var meta []Metadata
	f, err := os.ReadFile(filepath.Join(s.storePath, "metadata.json"))
	if err != nil {
		log.Fatalf("Error reading metadata file: %v", err)
	}
	if err := json.Unmarshal(f, &meta); err != nil {
		log.Fatalf("Error unmarshalling metadata file: %v", err)
	}

	s.data = meta
}
