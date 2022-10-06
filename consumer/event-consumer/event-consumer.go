package event_consumer

import (
	"InstaBot/events"
	"log"
	"sync"
	"time"
)

type Consumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	batchSize int
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int) *Consumer {
	return &Consumer{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
	}
}

func (c *Consumer) Start() error {
	for {
		gotEvents, err := c.fetcher.Fetch(c.batchSize)
		if err != nil {
			log.Printf("ERROR Consumer: %s", err.Error())
			continue
		}
		if len(gotEvents) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		if err = c.handleEvents(gotEvents); err != nil {
			log.Print(err)
			continue
		}

	}
}

func (c *Consumer) handleEvents(events []events.Event) error {
	var wg sync.WaitGroup
	wg.Add(len(events))
	for _, event := range events {
		func(wg *sync.WaitGroup) {
			log.Printf("got new event: %s", event.Text)

			if err := c.processor.Process(event); err != nil {
				log.Printf("can't handle event: %s", err.Error())
			}
			wg.Done()
		}(&wg)
	}
	wg.Wait()
	return nil
}
