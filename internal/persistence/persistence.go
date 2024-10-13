

package persistence

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/umgbhalla/gokv/internal/store"
)


type Persistence struct {
	store    *store.Store
	filename string
	interval time.Duration
	stopChan chan struct{}
	wg       sync.WaitGroup
}


func New(s *store.Store, filename string, interval time.Duration) *Persistence {
	return &Persistence{
		store:    s,
		filename: filename,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}


func (p *Persistence) Start() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := p.Save(); err != nil {
					// TODO: Implement proper error handling
					println("Error saving data:", err.Error())
				}
			case <-p.stopChan:
				return
			}
		}
	}()
}

func (p *Persistence) Save() error {
	data := p.store.GetAll()
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(p.filename, jsonData, 0644)
}

func (p *Persistence) Load() error {
	jsonData, err := os.ReadFile(p.filename)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, it's not an error
			// It might be the first run
			return nil
		}
		return err
	}

	var data map[string]store.Value
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	p.store.SetAll(data)
	return nil
}

func (p *Persistence) Stop() {
	close(p.stopChan)
	p.wg.Wait()
}
