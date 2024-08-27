package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"

	model "github.com/IgorGreusunset/shortener/internal/app"
)

type Storage struct {
	db   map[string]model.URL
	file *os.File
	mu   sync.RWMutex
	scan *bufio.Scanner
}

// Фабричный метод создания нового экземпляра хранилища
func NewStorage(db map[string]model.URL) *Storage {
	return &Storage{db: db}
}

func (s *Storage) SetFile(f *os.File) {
	s.file = f
}

type Repository interface {
	Create(ctx context.Context, record *model.URL) error
	GetByID(id string) (model.URL, bool)
	Ping() error
	CreateBatch(ctx context.Context, urls []model.URL) error
	UsersURLs(userID string) ([]model.URL, error)
	Delete(ctx context.Context, shorts []string, userID string) error
}

// Метод для создания новой записи в хранилище
func (s *Storage) Create(ctx context.Context, record *model.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[record.ID] = *record
	record.UUID = len(s.db)

	if s.file != nil {
		name := s.file.Name()
		if err := saveToFile(*record, name); err != nil {
			log.Printf("Error write to file: %v", err)
			return err
		}
	}
	return nil
}

// Метода для получения записи из хранилища
func (s *Storage) GetByID(id string) (model.URL, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.db[id]
	return url, ok
}

func (s *Storage) Ping() error {
	return nil
}

func (s *Storage) FillFromFile(file *os.File) error {
	url := &model.URL{}
	s.scan = bufio.NewScanner(file)

	for s.scan.Scan() {
		err := json.Unmarshal(s.scan.Bytes(), url)
		if err != nil {
			return err
		}
		s.db[url.ID] = *url
	}

	return nil
}

func saveToFile(url model.URL, file string) error {
	fil, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer fil.Close()

	data, err := json.Marshal(&url)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = fil.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) CreateBatch(ctx context.Context, urls []model.URL) error {
	for _, u := range urls {
		if err := s.Create(ctx, &u); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) UsersURLs(userID string) ([]model.URL, error) {
	result := make([]model.URL, 0)
	s.mu.RLock()
	for _, u := range s.db {
		if u.UserID == userID {
			result = append(result, u)
		}
	}
	s.mu.RUnlock()
	return result, nil
}

func (s *Storage) Delete(ctx context.Context, shorts []string, userID string) error {

	for _, short := range shorts {
		u, ok := s.db[short]
		if !ok {
			return errors.New("not found")
		}
		if u.UserID == userID {
			u.DeletedFlag = true
			s.db[short] = u
		}
	}

	return nil
}
