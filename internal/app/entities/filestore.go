package entities

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

type FileLinks struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStore struct {
	Path        string
	FileStorage *os.File
	mu          sync.RWMutex
	cache       map[string]string
}

func NewFileStore(path string) *FileStore {
	fs := &FileStore{
		Path:  path,
		cache: make(map[string]string),
	}
	
	fs.loadCache()
	return fs
}

func (f *FileStore) loadCache() {
	file, err := os.OpenFile(f.Path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var res FileLinks
		if err := json.Unmarshal([]byte(scanner.Text()), &res); err == nil {
			f.cache[res.ShortURL] = res.OriginalURL
		}
	}
}

func (f *FileStore) AddHash(hash, link, userID string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if _, exists := f.cache[hash]; exists {
		return hash, nil
	}

	f.cache[hash] = link

	file, err := os.OpenFile(f.Path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := json.Marshal(FileLinks{
		ShortURL:    hash,
		OriginalURL: link,
	})
	if err != nil {
		return "", err
	}

	writer := bufio.NewWriter(file)
	writer.Write(data)
	writer.WriteRune('\n')
	writer.Flush()

	return "", nil
}

func (f *FileStore) GetHash(hash string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	if link, exists := f.cache[hash]; exists {
		return link
	}
	
	return ""
}

func (f *FileStore) CheckValExists(link string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	for _, cachedLink := range f.cache {
		if cachedLink == link {
			return true
		}
	}
	
	return false
}
