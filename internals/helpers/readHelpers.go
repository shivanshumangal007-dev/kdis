package helpers

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

func readCommand(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")

	if len(line) == 0 || line[0] != '*' {
		return nil, fmt.Errorf("expected array, got: %q", line)
	}
	count, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, err
	}

	args := make([]string, count)
	for i := 0; i < count; i++ {
		bulkheader, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		bulkheader = strings.TrimRight(bulkheader, "\r\n")
		length, err := strconv.Atoi(bulkheader[1:])
		if err != nil {
			return nil, err
		}

		buf := make([]byte, length+2)
		_, err = readFull(reader, buf)
		if err != nil {
			return nil, err
		}
		args[i] = string(buf[:length])

	}

	return args, nil
}

func readFull(reader *bufio.Reader, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := reader.Read(buf[total:])
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}
