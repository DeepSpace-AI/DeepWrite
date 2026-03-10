package collab

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	MessageSync      = 0
	MessageAwareness = 1
)

const (
	SyncStep1  = 0
	SyncStep2  = 1
	SyncUpdate = 2
)

func EncodeVarUint(value uint64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, value)
	return buf[:n]
}

func DecodeVarUint(data []byte, index *int) (uint64, error) {
	if *index >= len(data) {
		return 0, fmt.Errorf("varuint out of range")
	}

	value, n := binary.Uvarint(data[*index:])
	if n <= 0 {
		return 0, fmt.Errorf("invalid varuint")
	}

	*index += n
	return value, nil
}

func BuildSyncFrame(syncType uint64, payload []byte) []byte {
	var buf bytes.Buffer
	buf.Write(EncodeVarUint(MessageSync))
	buf.Write(EncodeVarUint(syncType))
	buf.Write(payload)
	return buf.Bytes()
}

func BuildAwarenessFrame(payload []byte) []byte {
	var buf bytes.Buffer
	buf.Write(EncodeVarUint(MessageAwareness))
	buf.Write(payload)
	return buf.Bytes()
}

func ParseIncomingFrame(frame []byte) (messageType uint64, subtype uint64, payload []byte, err error) {
	idx := 0
	messageType, err = DecodeVarUint(frame, &idx)
	if err != nil {
		return 0, 0, nil, err
	}

	switch messageType {
	case MessageSync:
		subtype, err = DecodeVarUint(frame, &idx)
		if err != nil {
			return 0, 0, nil, err
		}
		return messageType, subtype, frame[idx:], nil
	case MessageAwareness:
		return messageType, 0, frame[idx:], nil
	default:
		return messageType, 0, frame[idx:], nil
	}
}
