package renderer

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// Server-sent events
// https://html.spec.whatwg.org/multipage/server-sent-events.html

type SSEvent struct {
	Event string
	Data  io.Reader
	ID    string
	Retry uint
}

func SSEventEncode(writer io.Writer, event SSEvent) error {
	buf := new(bytes.Buffer)
	if len(event.Event) > 0 {
		buf.WriteString(fmt.Sprintf("event: %s\n", event.Event))
	}
	if len(event.ID) > 0 {
		buf.WriteString(fmt.Sprintf("id: %s\n", event.ID))
	}
	if event.Retry > 0 {
		buf.WriteString(fmt.Sprintf("retry: %d\n", event.Retry))
	}

	buf.WriteString("data: ")
	if _, err := io.Copy(buf, event.Data); err != nil {
		return err
	}
	buf.WriteString("\n\n")

	_, err := writer.Write(buf.Bytes())
	return err
}

func SSEventDecode(reader io.Reader) ([]SSEvent, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(raw, []byte{'\n'})

	buf := new(bytes.Buffer)
	var event SSEvent
	var events []SSEvent

	for _, line := range lines {
		if len(line) == 0 {
			// Trim any leading and trailing space characters from line.
			data := bytes.TrimSpace(buf.Bytes())
			if len(data) == 0 && event.Event == "" {
				continue
			}
			if len(data) > 0 {
				event.Data = bytes.NewReader(data)
			}
			if event.Event == "" {
				event.Event = "message"
			}

			events = append(events, event)
			event = SSEvent{}
			buf.Reset()
			continue
		}

		if bytes.HasPrefix(line, []byte{':'}) {
			continue
		}

		var field, value []byte
		index := bytes.IndexRune(line, ':')
		if index != -1 {
			field = bytes.TrimSpace(line[:index])
			value = bytes.TrimSpace(line[index+1:])
		} else {
			field = line
			value = []byte{}
		}

		switch string(field) {
		case "event":
			event.Event = string(value)
		case "id":
			event.ID = string(value)
		case "retry":
			retry, err := strconv.Atoi(string(value))
			if err == nil {
				event.Retry = uint(retry)
			}
		case "data":
			buf.Write(value)
			buf.WriteString("\n")
		default:
			continue
		}
	}

	return events, nil
}
