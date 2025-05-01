package files

import (
	"encoding/json"
	"io"
	"os"
)

// Event представляет структуру события, содержащую UUID, сокращённый URL и оригинальный URL.
type Event struct {
	ID          string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Producer отвечает за запись событий в файл в формате JSON.
type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

// NewProducer создаёт новый Producer, который будет записывать события в указанный файл.
func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// WriteEvent записывает событие в файл.
func (p *Producer) WriteEvent(event *Event) error {
	return p.encoder.Encode(&event)
}

// Consumer отвечает за чтение событий из файла в формате JSON.
type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

// NewConsumer создаёт новый Consumer, который будет читать события из указанного файла.
func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

// ReadEvents читает все события из файла и возвращает их в виде среза.
func (c *Consumer) ReadEvents() ([]*Event, error) {
	var events []*Event

	for {
		event := Event{}
		if err := c.decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		events = append(events, &event)
	}

	return events, nil
}
