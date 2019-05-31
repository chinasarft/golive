package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"time"

	"github.com/chinasarft/golive/container/flv"
)

func (c *ConnClient) readTag(flvData []byte, tagType uint8, ctx context.Context) error {

	defer func() {
		log.Println("finished")
		c.group.Done()
	}()
	r := bytes.NewReader(flvData[9+4:])

	ctx, _ = context.WithCancel(ctx)
	start := time.Now().UnixNano()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			tag, err := flv.ParseTag(r)
			if err == io.EOF {
				r.Reset(flvData[9+4:])
				continue
			} else if err != nil {
				log.Println(err)
				return err
			}
			if tag.TagType == tagType {
				continue
			}
			if flv.FlvTagVideo == tag.TagType {
				log.Println("send video:", tag.Timestamp, tag.DataSize24)
				if err = c.rtmpHandler.SendVideoMessage(tag.Data, tag.Timestamp); err != nil {
					log.Println(err)
					return err
				}
			} else {
				log.Println("send audio:", tag.Timestamp, tag.DataSize24)
				if err = c.rtmpHandler.SendAudioMessage(tag.Data, tag.Timestamp); err != nil {
					log.Println(err)
					return err
				}
			}

			expectTime := start + int64(tag.Timestamp)*1000000
			now := time.Now().UnixNano()

			sleepTime := (expectTime-now)/1000000 - 1
			if sleepTime > 0 {
				time.Sleep(time.Millisecond * time.Duration(sleepTime))
			}
		}
	}
}

func (c *ConnClient) readAudioTag(flvData []byte, ctx context.Context) error {
	return c.readTag(flvData, flv.FlvTagAudio, ctx)
}

func (c *ConnClient) readVideoTag(flvData []byte, ctx context.Context) error {
	return c.readTag(flvData, flv.FlvTagVideo, ctx)
}
