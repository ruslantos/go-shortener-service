package file

import (
	"encoding/json"
	"io"
	"os"
)

type Event struct {
	ID          string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	//scanner *bufio.Scanner
	encoder *json.Encoder
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file: file,
		// создаём новый Writer
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) WriteEvent(event *Event) error {
	//data, err := json.Marshal(&event)
	//if err != nil {
	//	return err
	//}
	//
	//// записываем событие в буфер
	//if _, err := p.writer.Write(data); err != nil {
	//	return err
	//}
	//
	//// добавляем перенос строки
	//if err := p.writer.WriteByte('\n'); err != nil {
	//	return err
	//}
	//
	//// записываем буфер в файл
	//return p.writer.Flush()
	return p.encoder.Encode(&event)
}

type Consumer struct {
	file *os.File
	// добавляем reader в Consumer
	//reader *bufio.Reader
	decoder *json.Decoder
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file: file,
		// создаём новый Reader
		//reader: bufio.NewReader(file),
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) ReadEvent() (*Event, error) {
	// читаем данные до символа переноса строки
	//data, err := c.reader.ReadBytes('\n')
	//if err != nil {
	//	return nil, err
	//}

	// преобразуем данные из JSON-представления в структуру
	event := Event{}
	//err = json.Unmarshal(data, &event)
	//if err != nil {
	//	return nil, err
	//}

	if err := c.decoder.Decode(&event); err != nil {
		return nil, err
	}

	return &event, nil
}

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
