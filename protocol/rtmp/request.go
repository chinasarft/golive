package rtmp

import (
	"fmt"
	"io"
	"strings"

	"github.com/chinasarft/golive/utils/amf"
)

type RtmpUrl struct {
	appName    string
	streamName string
	tcurl      string
	originUrl  string
}

func parseRtmpUrl(addr string) (*RtmpUrl, error) {
	if strings.Index(addr, "rtmp") != 0 {
		return nil, fmt.Errorf("not correct rtmp url")
	}

	parts := strings.Split(addr, "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("wrong rtmp url:%s", addr)
	}
	return &RtmpUrl{
		appName:    parts[3],
		streamName: parts[4],
		tcurl:      strings.Join(parts[0:4], "/"),
		originUrl:  addr,
	}, nil
}

func (ru *RtmpUrl) getConnectCmdObj() amf.Object {

	cmdObj := make(amf.Object)
	cmdObj["app"] = ru.appName
	cmdObj["type"] = "nonprivate"
	cmdObj["flashVer"] = "FMS.3.1"
	cmdObj["tcUrl"] = ru.tcurl

	return cmdObj
}

/*
+----------------+---------+---------------------------------------+
|  Field Name    |  Type   |           Description                 |
+--------------- +---------+---------------------------------------+
| Command Name   | String  | Name of the command. Set to "connect".|
+----------------+---------+---------------------------------------+
| Transaction ID | Number  | Always set to 1.                      |
+----------------+---------+---------------------------------------+
| Command Object | Object  | Command information object which has  |
|                |         | the name-value pairs.                 |
+----------------+---------+---------------------------------------+
| Optional User  | Object  | Any optional information              |
| Arguments      |         |                                       |
+----------------+---------+---------------------------------------+
*/
func sendConnectMessage(w io.Writer, chunkPacker *ChunkPacker, cmdObj amf.Object) error {

	msg, err := NewConnectMessage(cmdObj)
	if err != nil {
		return err
	}

	return chunkPacker.WriteMessage(w, msg)
}

/*
+--------------+----------+----------------------------------------+
| Field Name   |   Type   |             Description                |
+--------------+----------+----------------------------------------+
| Command Name |  String  | Name of the command. Set to            |
|              |          | "createStream".                        |
+--------------+----------+----------------------------------------+
| Transaction  |  Number  | Transaction ID of the command.         |
| ID           |          |                                        |
+--------------+----------+----------------------------------------+
| Command      |  Object  | If there exists any command info this  |
| Object       |          | is set, else this is set to null type. |
+--------------+----------+----------------------------------------+
*/
func sendCreateStreamMessage(w io.Writer, chunkPacker *ChunkPacker, transactionId uint32, cmdObj amf.Object) error {

	msg, err := NewCreateStreamMessage(transactionId, cmdObj)
	if err != nil {
		return err
	}

	return chunkPacker.WriteMessage(w, msg)
}

func sendWidowAckMessage(w io.Writer, chunkPacker *ChunkPacker, windowSize uint32) error {
	msg := NewAckMessage(2500000)
	return chunkPacker.WriteMessage(w, msg)
}

func sendSetChunkMessage(w io.Writer, chunkPacker *ChunkPacker, chunkSize uint32) error {
	msg := NewSetChunkSizeMessage(chunkSize)
	return chunkPacker.WriteMessage(w, msg)
}

func sendGetStreamLengthMessage(w io.Writer, chunkPacker *ChunkPacker, transactionId uint32, streamName string) error {

	msg, err := NewGetStreamLengthMessage(transactionId, streamName)
	if err != nil {
		return err
	}

	return chunkPacker.WriteMessage(w, msg)
}

/*
+--------------+----------+-----------------------------------------+
| Field Name   |   Type   |             Description                 |
+--------------+----------+-----------------------------------------+
| Command Name |  String  | Name of the command. Set to "play".     |
+--------------+----------+-----------------------------------------+
| Transaction  |  Number  | Transaction ID set to 0.                |
| ID           |          |                                         |
+--------------+----------+-----------------------------------------+
| Command      |   Null   | Command information does not exist.     |
| Object       |          | Set to null type.                       |
+--------------+----------+-----------------------------------------+
| Stream Name  |  String  | Name of the stream to play.             |
|              |          | To play video (FLV) files, specify the  |
|              |          | name of the stream without a file       |
|              |          | extension (for example, "sample"). To   |
|              |          | play back MP3 or ID3 tags, you must     |
|              |          | precede the stream name with mp3:       |
|              |          | (for example, "mp3:sample". To play     |
|              |          | H.264/AAC files, you must precede the   |
|              |          | stream name with mp4: and specify the   |
|              |          | file extension. For example, to play the|
|              |          | file sample.m4v,specify "mp4:sample.m4v"|
|              |          |                                         |
+--------------+----------+-----------------------------------------+
|     Start    |  Number  |   An optional parameter that specifies  |
|              |          | the start time in seconds. The default  |
|              |          | value is -2, which means the subscriber |
|              |          | first tries to play the live stream     |
|              |          | specified in the Stream Name field. If a|
|              |          | live stream of that name is not found,it|
|              |          | plays the recorded stream of the same   |
|              |          | name. If there is no recorded stream    |
|              |          | with that name, the subscriber waits for|
|              |          | a new live stream with that name and    |
|              |          | plays it when available. If you pass -1 |
|              |          | in the Start field, only the live stream|
|              |          | specified in the Stream Name field is   |
|              |          | played. If you pass 0 or a positive     |
|              |          | number in the Start field, a recorded   |
|              |          | stream specified in the Stream Name     |
|              |          | field is played beginning from the time |
|              |          | specified in the Start field. If no     |
|              |          | recorded stream is found, the next item |
|              |          | in the playlist is played.              |
|              |          |                                         |
+--------------+----------+-----------------------------------------+
|   Duration   |  Number  | An optional parameter that specifies the|
|              |          | duration of playback in seconds. The    |
|              |          | default value is -1. The -1 value means |
|              |          | a live stream is played until it is no  |
|              |          | longer available or a recorded stream is|
|              |          | played until it ends. If you pass 0, it |
|              |          | plays the single frame since the time   |
|              |          | specified in the Start field from the   |
|              |          | beginning of a recorded stream. It is   |
|              |          | assumed that the value specified in     |
|              |          | the Start field is equal to or greater  |
|              |          | than 0. If you pass a positive number,  |
|              |          | it plays a live stream for              |
|              |          | the time period specified in the        |
|              |          | Duration field. After that it becomes   |
|              |          | available or plays a recorded stream    |
|              |          | for the time specified in the Duration  |
|              |          | field. (If a stream ends before the     |
|              |          | time specified in the Duration field,   |
|              |          | playback ends when the stream ends.)    |
|              |          | If you pass a negative number other     |
|              |          | than -1 in the Duration field, it       |
|              |          | interprets the value as if it were -1.  |
|              |          |                                         |
+--------------+----------+-----------------------------------------+
| Reset        | Boolean  | An optional Boolean value or number     |
|              |          | that specifies whether to flush any     |
|              |          | previous playlist.                      |
+--------------+----------+-----------------------------------------+
*/
func sendPlayMessage(w io.Writer, chunkPacker *ChunkPacker, transactionId uint32,
	duration int, streamName string) error {

	msg, err := NewPlayMessage(transactionId, streamName, duration)
	if err != nil {
		return err
	}
	return chunkPacker.WriteMessage(w, msg)
}

/*
+---------------+--------------------------------------------------+
|  SetBuffer    | The client sends this event to inform the server |
|  Length (=3)  | of the buffer size (in milliseconds) that is     |
|               |used to buffer any data coming over a stream.     |
|               |This event is sent before the server starts       |
|               |processing the stream. The first 4 bytes of the   |
|               |event data represent the stream ID and the next   |
|               |4 bytes represent the buffer length, in           |
|               |milliseconds.                                     |
+---------------+--------------------------------------------------+
*/

func sendSetBufferLengthMessage(w io.Writer, chunkPacker *ChunkPacker,
	transactionId uint32, bufferInMiliSecond int) error {

	msg := NewSetBufferLengthMessage(transactionId, bufferInMiliSecond)
	return chunkPacker.WriteMessage(w, msg)
}

func sendPublishMessage(w io.Writer, chunkPacker *ChunkPacker,
	transactionId uint32, appName, streamName string) error {

	msg, err := NewPublishMessage(transactionId, appName, streamName)
	if err != nil {
		return err
	}
	return chunkPacker.WriteMessage(w, msg)
}

func sendReleaseStreamMessage(w io.Writer, chunkPacker *ChunkPacker,
	transactionId uint32, streamName string) error {

	msg, err := NewReleaseStreamMessage(transactionId, streamName)
	if err != nil {
		return err
	}
	return chunkPacker.WriteMessage(w, msg)
}

func sendFCPublishMessage(w io.Writer, chunkPacker *ChunkPacker,
	transactionId uint32, streamName string) error {

	msg, err := NewFCPublishMessage(transactionId, streamName)
	if err != nil {
		return err
	}
	return chunkPacker.WriteMessage(w, msg)
}
