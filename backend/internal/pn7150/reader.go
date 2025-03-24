package pn7150

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
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
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

func tmp_main() {
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
	initTagCh()
	go tagPoll(reader.TagChannelPoll)

	return nil
}

func NewTagReader(deviceConnection string, resetPin int) *TagReader {
  return &TagReader{
    TagChannel: make(chan string, 10),
    TagChannelPoll: make(chan bool),
  }
}

// Reset performs a hardware reset by pulling the ResetPin low and then releasing.
func (reader *TagReader) Reset() {
  // TODO: needed?
	log.Println("Resetting the reader..")
}

func (reader *TagReader) Cleanup() error {
	return nil
}

func (reader *TagReader) GetTagChannel() <-chan string {
	return reader.TagChannel
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
		case msg := <-tagCh:
			fmt.Println("> Text:", msg.uid)
			fmt.Println("> UID:", msg.text)
			fmt.Println("> err:", msg.err)

      reader.TagChannel <- msg.text
    }
    time.Sleep(time.Second * 1)
  }
}

