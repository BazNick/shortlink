package service

import (
	"context"
	"fmt"

	"github.com/BazNick/shortlink/internal/app/apperr"
	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/functions"
	"github.com/BazNick/shortlink/internal/app/storage"
)

// UrlService предоставляет бизнес-логику для операций с URL-адресами
type URLService struct {
	storage       storage.Storage
	workerManager *entities.DeleteWorkerManager
}

// NewURLService создает новый экземпляр UrlService
func NewURLService(storage storage.Storage, workerManager *entities.DeleteWorkerManager) *URLService {
	return &URLService{
		storage:       storage,
		workerManager: workerManager,
	}
}

// CreateShortURLRequest представляет собой запрос на создание короткого URL-адреса
type CreateShortURLRequest struct {
	URL    string
	UserID string
}

// CreateShortURLResponse представляет собой ответ на создание короткого URL-адреса
type CreateShortURLResponse struct {
	ShortURL      string
	AlreadyExists bool
}

// CreateShortURL создает короткий URL-адрес из длинного URL-адреса
func (s *URLService) CreateShortURL(ctx context.Context, req CreateShortURLRequest) (*CreateShortURLResponse, error) {
	if _, ok := s.storage.(*entities.DB); !ok {
		alreadyExists := s.storage.CheckValExists(req.URL)
		if alreadyExists {
			existingShortURL := s.getExistingShortURL(req.URL)
			if existingShortURL != "" {
				return &CreateShortURLResponse{
					ShortURL:      existingShortURL,
					AlreadyExists: true,
				}, nil
			}
		}
	}

	randStr := functions.RandSeq(8)

	shortURL, err := s.storage.AddHash(randStr, req.URL, req.UserID)
	if err != nil {
		if err == apperr.ErrValAlreadyExists {
			return &CreateShortURLResponse{
				ShortURL:      shortURL,
				AlreadyExists: true,
			}, nil
		}
		return nil, fmt.Errorf("failed to add hash: %w", err)
	}

	return &CreateShortURLResponse{
		ShortURL:      randStr,
		AlreadyExists: false,
	}, nil
}

// GetOriginalURLRequest представляет собой запрос на получение исходного URL-адреса
type GetOriginalURLRequest struct {
	ShortURL string
}

// GetOriginalURLResponse представляет собой ответ на получение исходного URL-адреса
type GetOriginalURLResponse struct {
	OriginalURL string
	Found       bool
}

// GetOriginalURL извлекает исходный URL-адрес из короткого URL-адреса
func (s *URLService) GetOriginalURL(ctx context.Context, req GetOriginalURLRequest) (*GetOriginalURLResponse, error) {
	originalURL := s.storage.GetHash(req.ShortURL)
	if originalURL == "" {
		return &GetOriginalURLResponse{
			OriginalURL: "",
			Found:       false,
		}, nil
	}

	return &GetOriginalURLResponse{
		OriginalURL: originalURL,
		Found:       true,
	}, nil
}

// BatchURLItem представляет собой один URL-адрес в пакетном запросе
type BatchURLItem struct {
	CorrelationID string
	OriginalURL   string
}

// BatchURLResult представляет результат для одного URL-адреса в пакетном ответе
type BatchURLResult struct {
	CorrelationID string
	ShortURL      string
	AlreadyExists bool
}

// CreateShortURLBatchRequest представляет собой запрос на создание нескольких коротких URL-адресов
type CreateShortURLBatchRequest struct {
	URLs    []BatchURLItem
	UserID  string
	BaseURL string
}

// CreateShortURLBatchResponse представляет собой ответ на создание нескольких коротких URL-адресов
type CreateShortURLBatchResponse struct {
	Results []BatchURLResult
}

// CreateShortURLBatch создает несколько коротких URL-адресов в одном запросе
func (s *URLService) CreateShortURLBatch(ctx context.Context, req CreateShortURLBatchRequest) (*CreateShortURLBatchResponse, error) {
	results := make([]BatchURLResult, 0, len(req.URLs))

	for _, item := range req.URLs {
		createReq := CreateShortURLRequest{
			URL:    item.OriginalURL,
			UserID: req.UserID,
		}

		createResp, err := s.CreateShortURL(ctx, createReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create short URL for correlation ID %s: %w", item.CorrelationID, err)
		}

		shortURL := createResp.ShortURL
		if req.BaseURL != "" {
			shortURL = req.BaseURL + "/" + shortURL
		}

		results = append(results, BatchURLResult{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
			AlreadyExists: createResp.AlreadyExists,
		})
	}

	return &CreateShortURLBatchResponse{
		Results: results,
	}, nil
}

// UserURLItem представляет собой URL-адрес пользователя
type UserURLItem struct {
	ShortURL    string
	OriginalURL string
}

// GetUserURLsRequest представляет собой запрос на получение URL-адресов пользователей
type GetUserURLsRequest struct {
	UserID  string
	BaseURL string
}

// GetUserURLsResponse представляет собой ответ на получение URL-адресов пользователей
type GetUserURLsResponse struct {
	URLs []UserURLItem
}

// GetUserURLs извлекает все URL-адреса, созданные пользователем
func (s *URLService) GetUserURLs(ctx context.Context, req GetUserURLsRequest) (*GetUserURLsResponse, error) {
	return &GetUserURLsResponse{
		URLs: []UserURLItem{},
	}, nil
}

// DeleteUserURLsRequest представляет собой запрос на удаление URL-адресов пользователей
type DeleteUserURLsRequest struct {
	ShortURLs []string
	UserID    string
}

// DeleteUserURLsResponse представляет собой ответ на удаление URL-адресов пользователей
type DeleteUserURLsResponse struct {
	Success bool
}

// DeleteUserURLs удаляет несколько URL-адресов, созданных пользователем
func (s *URLService) DeleteUserURLs(ctx context.Context, req DeleteUserURLsRequest) (*DeleteUserURLsResponse, error) {
	if s.workerManager != nil {
		deleteReq := entities.DeleteRequest{
			UserID:    req.UserID,
			ShortURLs: req.ShortURLs,
		}
		s.workerManager.SendDeleteRequest(deleteReq)
	}

	return &DeleteUserURLsResponse{
		Success: true,
	}, nil
}

// PingRequest представляет собой запрос на проверку связи
type PingRequest struct{}

// PingResponse представляет собой ответ на пинг-запрос
type PingResponse struct {
	Healthy bool
	Message string
}

// Ping проверяет работоспособность БД
func (s *URLService) Ping(ctx context.Context, req PingRequest) (*PingResponse, error) {
	return &PingResponse{
		Healthy: true,
		Message: "Service is healthy",
	}, nil
}

// GetStatsRequest представляет собой запрос статистики
type GetStatsRequest struct{}

// GetStatsResponse представляет собой ответ по статистике
type GetStatsResponse struct {
	URLs  int
	Users int
}

// GetStats возвращает статистику сервиса
func (s *URLService) GetStats(ctx context.Context, req GetStatsRequest) (*GetStatsResponse, error) {
	urls, users, err := s.storage.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &GetStatsResponse{
		URLs:  urls,
		Users: users,
	}, nil
}

// getExistingShortURL извлекает существующий короткий URL-адрес для данного исходного URL-адреса
func (s *URLService) getExistingShortURL(originalURL string) string {
	if hashDict, ok := s.storage.(*entities.HashDict); ok {
		return hashDict.RevDict[originalURL]
	}
	return ""
}
