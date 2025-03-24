package main

/*
#include "nfc_lib.h"
#include <stdint.h>
*/
import "C"

import (
	"encoding/hex"
	"fmt"
	"unsafe"
	"sync"
)

type Tag struct {
	text string
	uid string
	err string
}

var (
	tagCh chan Tag
	tagChRemove chan bool
	once      sync.Once
)


func initTagCh() {
	once.Do(func() {
		tagCh = make(chan Tag)
		tagChRemove = make(chan bool)
	})
}

//export onTagRemove
func onTagRemove() {
	initTagCh()
	tagChRemove <- true
}

//export exportTag
func exportTag(p *C.Tag) {
	initTagCh()
	uid := C.GoBytes(unsafe.Pointer(&p.uid), p.uid_length)
	text := C.GoBytes(unsafe.Pointer(&p.text), p.text_length)
	err := ""
	switch int(p.error) {
	case 1:
		err = "Not a NDEF tag"
	case 2:
		err = "Not a NDEF text record"
	case 3:
		err = "Read NDEF text error"
	}

	tagContext := Tag {
		uid: hex.EncodeToString(uid),
		text: string(text),
		err: err,
	}

	tagCh <- tagContext
}

func tagPoll(tag_poll chan bool) {
	C.read_tag()
	fmt.Println("Tag poll ends")
	tag_poll <- true
}

func main() {
	initTagCh()

	tag_poll := make(chan bool)

	go tagPoll(tag_poll)

	for {
		select {
		case msg := <-tagCh:
			fmt.Println("> Text:", msg.uid)
			fmt.Println("> UID:", msg.text)
			fmt.Println("> err:", msg.err)
		case <-tagChRemove:
			fmt.Println("> Removed")
		case <-tag_poll:
			fmt.Println("Tag polls ends")
			return
		}
	}

}

type Reader interface {
	ListenForTags(ctx context.Context) error
	Cleanup() error
	Reset()
	GetTagChannel() <-chan string
}

type TagReader struct {
	TagChannel       chan string
  TagChannelPoll   chan bool
}

func (reader *TagReader) init() error {
	tag_poll := make(chan bool)


	return nil
}

func NewTagReader(deviceConnection string, resetPin int) *TagReader {
	return &TagReader{DeviceConnection: deviceConnection, TagChannel: make(chan string, 10), ResetPin: resetPin}
}

// Reset performs a hardware reset by pulling the ResetPin low and then releasing.
func (reader *TagReader) Reset() {
  // TODO: needed?
	log.Println("Resetting the reader..")
}

func (reader *TagReader) Cleanup() error {
	return reader.reader.Close()
}

func (reader *TagReader) GetTagChannel() <-chan string {
	return reader.TagChannel
}

func (reader *TagReader) getIdFromTarget(target nfc.Target) (*string, error) {
	var UID string
	// Transform the target to a specific tag Type and send the UID to the channel
	switch target.Modulation() {
	case nfc.Modulation{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106}:
		var card = target.(*nfc.ISO14443aTarget)
		var UIDLen = card.UIDLen
		var ID = card.UID
		UID = hex.EncodeToString(ID[:UIDLen])
	default:
		return nil, errors.New("Unknown modulation")
	}

	return &UID, nil
}

func (reader *TagReader) ListenForTags(ctx context.Context) error {
	//Initialize the reader
	err := reader.init()
	if err != nil {
		return errors.Wrap(err, "Cannot initialize the reader")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Poll for 300ms
			tagCount, target, err := reader.reader.InitiatorPollTarget(modulations, 1, 300*time.Millisecond)
			if err != nil {
				log.Println("Error polling the reader", err)
				continue
			}

			// Check if a tag was detected
			if tagCount > 0 {
				// Get the UID of the tag based on the modulation type
				id, err := reader.getIdFromTarget(target)
				if err != nil {
					log.Println("Error getting ID from target", err)
					continue
				}

				// Send the UID of the tag to main goroutine
				reader.TagChannel <- *id

				// Read data from the tag
				// err = reader.readDataFromTag(target)
				// if err != nil {
				// 	log.Println("Error reading data from the tag", err)
				// 	continue
				// }
			}

			time.Sleep(time.Second * 1)
		}
	}
}

func (reader *TagReader) readDataFromTag(target nfc.Target) error {
	// // Read data from the tag
	// _, err := reader.reader.InitiatorTransceiveBits(target, []byte{0x30, 0x00}, 7)
	// if err != nil {
	// 	return errors.Wrap(err, "Error reading data from the tag")
	// }

	// return nil
	// Leggere i primi 16 blocchi del tag
	for i := 0; i < 16; i++ {
		cmd := []byte{0x30, byte(i)} // 0x30 = comando di lettura, i = numero del blocco
		resp := make([]byte, 16)     // La risposta deve contenere 16 byte

		// _, err := reader.reader.InitiatorTransceiveBits(target, []byte{0x30, byte(i)}, 7)
		// if err != nil {
		// 	return errors.Wrap(err, "Error reading data from the tag")
		// }

		n, err := reader.reader.InitiatorTransceiveBytes(cmd, resp, -1)
		if err != nil {
			log.Printf("Errore nella lettura del blocco %d: %v\n", i, err)
			continue
		}

		// Stampa il blocco letto
		log.Printf("Block %02d: % X\n", i, resp[:n])
	}
	return nil
}
