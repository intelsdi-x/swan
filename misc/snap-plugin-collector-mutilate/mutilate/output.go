package mutilate

import (
	"bytes"
	"encoding/csv"
	"golang.org/x/exp/inotify"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	NEW_LINE byte = 10
)

type MutilateRow struct {
	time    time.Time
	latency float64
}

func parse_mutilate_output(event inotify.Event, baseTime time.Time) ([]MutilateRow, error) {
	csvFile, readError := os.Open(event.Name)
	defer csvFile.Close()
	if readError != nil {
		var output []MutilateRow
		return output, readError
	}
	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = ' '
	startTime := get_first_row_time(csvFile, baseTime)
	output := get_structs_from_file(csvReader, startTime)

	return output, nil
}

func get_first_row_time(file *os.File, baseTime time.Time) time.Time {
	file.Seek(0, os.SEEK_END)
	defer file.Seek(0, os.SEEK_SET)

	lastLineBytes := find_last_row(file)
	reader := csv.NewReader(bytes.NewReader(lastLineBytes))
	reader.Comma = ' '
	row, _ := reader.Read()
	row[0] = strings.Trim(row[0], "")
	seconds, _ := strconv.ParseFloat(row[0], 64)
	lastRowUnix := baseTime.Unix() - int64(seconds)

	return time.Unix(int64(lastRowUnix), 0)
}

func get_structs_from_file(csvReader *csv.Reader, startTime time.Time) []MutilateRow {
	var output []MutilateRow
	for true {
		row, error := csvReader.Read()
		if error == io.EOF {
			break
		}
		floatOffset, _ := strconv.ParseFloat(row[0], 64)
		offset := time.Duration(time.Duration(floatOffset) * time.Second)
		eventTime := startTime.Add(offset)
		latency, _ := strconv.ParseFloat(row[1], 64)
		output = append(output, MutilateRow{eventTime, latency})
	}

	return output
}

func find_last_row(file *os.File) []byte {
	var i int64
	var line []byte
	singleByte := make([]byte, 1)
	for true {
		file.Seek(i, os.SEEK_END)
		file.Read(singleByte)
		if singleByte[0] == NEW_LINE {
			break
		}
		line = append(singleByte, line...)
		i--
	}

	return line
}
