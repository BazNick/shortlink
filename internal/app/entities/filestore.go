package entities

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
)

type FileLinks struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStore struct {
	Path        string
	FileStorage *os.File
}

func NewFileStore(path string) *FileStore {
	return &FileStore{
		Path: path,
	}
}

func (f *FileStore) AddHash(hash, link string) (string, error) {
	file, err := os.OpenFile(f.Path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return "", err
	}

	bufWriter := bufio.NewWriter(file)

	data, err := json.Marshal(FileLinks{
		ShortURL:    hash,
		OriginalURL: link,
	})
	if err != nil {
		file.Close()
		return "", err
	}

	data = append(data, '\n')

	if _, err = bufWriter.Write(data); err != nil {
		file.Close()
		return "", err
	}

	if err := bufWriter.Flush(); err != nil {
		file.Close()
		return "", err
	}

	if err := file.Sync(); err != nil {
		file.Close()
		return "", err
	}

	return "", nil
}

func (f *FileStore) GetHash(hash string) string {
	reader, errOpenFile := os.OpenFile(f.Path, os.O_RDONLY|os.O_CREATE, 0666)
	if errOpenFile != nil {
		log.Fatalf("Ошибка при открытии файла %s: %v", f.Path, errOpenFile)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		var res FileLinks
		err := json.Unmarshal(scanner.Bytes(), &res)
		if err != nil {
			log.Fatalf("Ошибка при открытии файла %s: %v", f.Path, err)
		}
		if hash == res.ShortURL {
			return res.OriginalURL
		}
	}
	return ""
}

func (f *FileStore) CheckValExists(link string) bool {
	reader, errOpenFile := os.OpenFile(f.Path, os.O_RDONLY|os.O_CREATE, 0666)
	if errOpenFile != nil {
		log.Fatalf("Ошибка при открытии файла %s: %v", f.Path, errOpenFile)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		var res FileLinks
		err := json.Unmarshal(scanner.Bytes(), &res)
		if err != nil {
			log.Fatalf("Ошибка при открытии файла %s: %v", f.Path, err)
		}
		if link == res.OriginalURL {
			return true
		}
	}
	return false
}
