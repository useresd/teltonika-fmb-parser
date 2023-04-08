package main

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"time"
)

type AVLDataArray struct {
	CodecID  uint8
	NoOfData uint8
	AVLData  []AVLData
}

type AVLData struct {
	Timestamp  time.Time
	Priority   uint8
	GPSElement GPSElement
	IOElement  IOElement
}

type IOElement struct {
	EventIOID   uint8
	NoOfTotalIO uint8
	OneByteIO   []map[uint8]uint8
	TwoByteIO   []map[uint8]uint16
	FourByteIO  []map[uint8]uint32
	EightByteIO []map[uint8]uint64
}

type GPSElement struct {
	Longitude  uint32
	Latitude   uint32
	Altitude   uint16
	Angle      uint16
	Satellites uint8
	Speed      uint16
}

func parseGPSElement(data []byte) GPSElement {

	var gps GPSElement

	gps.Longitude = binary.BigEndian.Uint32(data[0:4])
	gps.Latitude = binary.BigEndian.Uint32(data[4:8])
	gps.Altitude = binary.BigEndian.Uint16(data[8:10])
	gps.Angle = binary.BigEndian.Uint16(data[10:12])
	gps.Satellites = data[12]
	gps.Speed = binary.BigEndian.Uint16(data[13:15])

	return gps
}

func parseTimestamp(data []byte) time.Time {
	unixTimestamp := binary.BigEndian.Uint64(data)
	return time.UnixMilli(int64(unixTimestamp))
}

func parseOneByteIO(data []byte) []map[uint8]uint8 {

	oneByteIO := []map[uint8]uint8{}

	for i := 0; i < len(data); i += 2 {
		oneByteIO = append(oneByteIO, map[uint8]uint8{data[i]: data[i+1]})
	}

	return oneByteIO
}

func parseTwoByteIO(data []byte) []map[uint8]uint16 {

	twoByteIO := []map[uint8]uint16{}

	for i := 0; i < len(data); i += 3 {
		twoByteIO = append(twoByteIO, map[uint8]uint16{data[i]: binary.BigEndian.Uint16(data[i+1 : i+3])})
	}

	return twoByteIO
}

func parseFourByteIO(data []byte) []map[uint8]uint32 {

	fourByteIO := []map[uint8]uint32{}

	for i := 0; i < len(data); i += 5 {
		fourByteIO = append(fourByteIO, map[uint8]uint32{data[i]: binary.BigEndian.Uint32(data[i+1 : i+5])})
	}

	return fourByteIO
}

func parseEightByteIO(data []byte) []map[uint8]uint64 {

	eightByteIO := []map[uint8]uint64{}

	for i := 0; i < len(data); i += 9 {
		eightByteIO = append(eightByteIO, map[uint8]uint64{data[i]: binary.BigEndian.Uint64(data[i+1 : i+9])})
	}

	return eightByteIO
}

func Parse(body []byte) AVLDataArray {

	var avlDataArray AVLDataArray

	// codec ID
	codecID := body[0]
	avlDataArray.CodecID = codecID

	// no of data
	noOfData := body[1]
	avlDataArray.NoOfData = noOfData

	startIndex := 2

	for i := 0; i < int(noOfData); i++ {

		avlData := AVLData{}

		// timestamp
		timestampStartIndex := startIndex
		timestampEndIndex := timestampStartIndex + 8
		timestamp := body[timestampStartIndex:timestampEndIndex]
		avlData.Timestamp = parseTimestamp(timestamp)

		// priority
		priorityIndex := timestampEndIndex
		priority := body[timestampEndIndex]
		avlData.Priority = priority

		// gps
		gpsStartIndex := priorityIndex + 1
		gpsEndIndex := gpsStartIndex + 15
		gps := body[gpsStartIndex:gpsEndIndex]
		avlData.GPSElement = parseGPSElement(gps)

		IOElement := IOElement{}

		// IO
		eventIOIDIndex := gpsEndIndex
		eventIOID := body[eventIOIDIndex]
		IOElement.EventIOID = eventIOID

		// Total Number of IO
		noOfTotalIOIndex := eventIOIDIndex + 1
		noOfTotalIO := body[noOfTotalIOIndex]
		IOElement.NoOfTotalIO = noOfTotalIO

		// One Byte IO Number
		noOfOneByteIOIndex := noOfTotalIOIndex + 1
		noOfOneByteIO := body[noOfOneByteIOIndex]

		// One Byte IO Data
		oneByteIOStartIndex := noOfOneByteIOIndex + 1
		oneByteIOEndIndex := oneByteIOStartIndex + int(noOfOneByteIO)*2
		oneByteIO := body[oneByteIOStartIndex:oneByteIOEndIndex]
		IOElement.OneByteIO = parseOneByteIO(oneByteIO)

		// Two Byte IO Number
		noOfTwoByteIOIndex := oneByteIOEndIndex
		noOfTwoByteIO := body[noOfTwoByteIOIndex]

		// Two Byte IO Data
		twoByteIOStartIndex := noOfTwoByteIOIndex + 1
		twoByteIOEndIndex := twoByteIOStartIndex + int(noOfTwoByteIO)*3
		twoByteIO := body[twoByteIOStartIndex:twoByteIOEndIndex]
		IOElement.TwoByteIO = parseTwoByteIO(twoByteIO)

		// Four Byte IO
		noOfFourByteIOIndex := twoByteIOEndIndex
		noOfFourByteIO := body[noOfFourByteIOIndex]
		fourByteIOStartIndex := noOfFourByteIOIndex + 1
		fourByteIOEndIndex := fourByteIOStartIndex + int(noOfFourByteIO)*5
		fourByteIO := body[fourByteIOStartIndex:fourByteIOEndIndex]
		IOElement.FourByteIO = parseFourByteIO(fourByteIO)

		// Eight Byte IO
		noOfEightByteIOIndex := fourByteIOEndIndex
		noOfEightByteIO := body[noOfEightByteIOIndex]
		eightByteIOStartIndex := noOfEightByteIOIndex + 1
		eightByteIOEndIndex := eightByteIOStartIndex + int(noOfEightByteIO)*9
		eightByteIO := body[eightByteIOStartIndex:eightByteIOEndIndex]
		IOElement.EightByteIO = parseEightByteIO(eightByteIO)

		startIndex = noOfEightByteIOIndex + 1 + int(noOfEightByteIO)*9

		avlData.IOElement = IOElement
		avlDataArray.AVLData = append(avlDataArray.AVLData, avlData)
	}

	return avlDataArray
}

func main() {

	body := []byte{
		0x08, 0x04, 0x00, 0x00, 0x01, 0x13, 0xfc, 0x20, 0x8d, 0xff,
		0x00, 0x0f, 0x14, 0xf6, 0x50, 0x20, 0x9c, 0xca, 0x80, 0x00,
		0x6f, 0x00, 0xd6, 0x04, 0x00, 0x04, 0x00, 0x04, 0x03, 0x01,
		0x01, 0x15, 0x03, 0x16, 0x03, 0x00, 0x01, 0x46, 0x00, 0x00,
		0x01, 0x5d, 0x00, 0x00, 0x00, 0x01, 0x13, 0xfc, 0x17, 0x61,
		0x0b, 0x00, 0x0f, 0x14, 0xff, 0xe0, 0x20, 0x9c, 0xc5, 0x80,
		0x00, 0x6e, 0x00, 0xc0, 0x05, 0x00, 0x01, 0x00, 0x04, 0x03,
		0x01, 0x01, 0x15, 0x03, 0x16, 0x01, 0x00, 0x01, 0x46, 0x00,
		0x00, 0x01, 0x5e, 0x00, 0x00, 0x00, 0x01, 0x13, 0xfc, 0x28,
		0x49, 0x45, 0x00, 0x0f, 0x15, 0x0f, 0x00, 0x20, 0x9c, 0xd2,
		0x00, 0x00, 0x95, 0x01, 0x08, 0x04, 0x00, 0x00, 0x00, 0x04,
		0x03, 0x01, 0x01, 0x15, 0x00, 0x16, 0x03, 0x00, 0x01, 0x46,
		0x00, 0x00, 0x01, 0x5d, 0x00, 0x00, 0x00, 0x01, 0x13, 0xfc,
		0x26, 0x7c, 0x5b, 0x00, 0x0f, 0x15, 0x0a, 0x50, 0x20, 0x9c,
		0xcc, 0xc0, 0x00, 0x93, 0x00, 0x68, 0x04, 0x00, 0x00, 0x00,
		0x04, 0x03, 0x01, 0x01, 0x15, 0x00, 0x16, 0x03, 0x00, 0x01,
		0x46, 0x00, 0x00, 0x01, 0x5b, 0x00, 0x04,
	}

	parsed := Parse(body)

	json.NewEncoder(os.Stdout).Encode(parsed)
}
