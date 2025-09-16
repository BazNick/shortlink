package grpc

import (
	"context"
	"net"

	"github.com/BazNick/shortlink/internal/app/entities"
	"github.com/BazNick/shortlink/internal/app/service"
	"github.com/BazNick/shortlink/internal/app/storage"
	"github.com/BazNick/shortlink/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server представляет gRPC сервер для сервиса сокращения ссылок
type Server struct {
	proto.UnimplementedShortLinkServiceServer
	urlService *service.URLService
	grpcServer *grpc.Server
}

// NewServer создаёт новый gRPC сервер
func NewServer(storage storage.Storage, workerManager *entities.DeleteWorkerManager) *Server {
	urlService := service.NewURLService(storage, workerManager)

	grpcServer := grpc.NewServer()
	server := &Server{
		urlService: urlService,
		grpcServer: grpcServer,
	}

	proto.RegisterShortLinkServiceServer(grpcServer, server)
	return server
}

// Start запускает gRPC сервер на указанном адресе
func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	return s.grpcServer.Serve(listener)
}

// Stop корректно останавливает gRPC сервер
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}

// CreateShortURL реализует gRPC метод CreateShortURL
func (s *Server) CreateShortURL(ctx context.Context, req *proto.CreateShortURLRequest) (*proto.CreateShortURLResponse, error) {
	serviceReq := service.CreateShortURLRequest{
		URL:    req.Url,
		UserID: req.UserId,
	}

	resp, err := s.urlService.CreateShortURL(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create short URL: %v", err)
	}

	return &proto.CreateShortURLResponse{
		ShortUrl:      resp.ShortURL,
		AlreadyExists: resp.AlreadyExists,
	}, nil
}

// GetOriginalURL реализует gRPC метод GetOriginalURL
func (s *Server) GetOriginalURL(ctx context.Context, req *proto.GetOriginalURLRequest) (*proto.GetOriginalURLResponse, error) {
	serviceReq := service.GetOriginalURLRequest{
		ShortURL: req.ShortUrl,
	}

	resp, err := s.urlService.GetOriginalURL(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get original URL: %v", err)
	}

	return &proto.GetOriginalURLResponse{
		OriginalUrl: resp.OriginalURL,
		Found:       resp.Found,
	}, nil
}

// CreateShortURLBatch реализует gRPC метод CreateShortURLBatch
func (s *Server) CreateShortURLBatch(ctx context.Context, req *proto.CreateShortURLBatchRequest) (*proto.CreateShortURLBatchResponse, error) {
	serviceURLs := make([]service.BatchURLItem, len(req.Urls))
	for i, url := range req.Urls {
		serviceURLs[i] = service.BatchURLItem{
			CorrelationID: url.CorrelationId,
			OriginalURL:   url.OriginalUrl,
		}
	}

	serviceReq := service.CreateShortURLBatchRequest{
		URLs:    serviceURLs,
		UserID:  req.UserId,
		BaseURL: "", // gRPC не имеет концепции базового URL, поэтому возвращаем только хэш
	}

	resp, err := s.urlService.CreateShortURLBatch(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create short URL batch: %v", err)
	}

	protoResults := make([]*proto.BatchURLResult, len(resp.Results))
	for i, result := range resp.Results {
		protoResults[i] = &proto.BatchURLResult{
			CorrelationId: result.CorrelationID,
			ShortUrl:      result.ShortURL,
			AlreadyExists: result.AlreadyExists,
		}
	}

	return &proto.CreateShortURLBatchResponse{
		Results: protoResults,
	}, nil
}

// GetUserURLs реализует gRPC метод GetUserURLs
func (s *Server) GetUserURLs(ctx context.Context, req *proto.GetUserURLsRequest) (*proto.GetUserURLsResponse, error) {
	serviceReq := service.GetUserURLsRequest{
		UserID:  req.UserId,
		BaseURL: "", // gRPC не имеет концепции базового URL
	}

	resp, err := s.urlService.GetUserURLs(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user URLs: %v", err)
	}

	protoURLs := make([]*proto.UserURLItem, len(resp.URLs))
	for i, url := range resp.URLs {
		protoURLs[i] = &proto.UserURLItem{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		}
	}

	return &proto.GetUserURLsResponse{
		Urls: protoURLs,
	}, nil
}

// DeleteUserURLs реализует gRPC метод DeleteUserURLs
func (s *Server) DeleteUserURLs(ctx context.Context, req *proto.DeleteUserURLsRequest) (*proto.DeleteUserURLsResponse, error) {
	serviceReq := service.DeleteUserURLsRequest{
		ShortURLs: req.ShortUrls,
		UserID:    req.UserId,
	}

	resp, err := s.urlService.DeleteUserURLs(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user URLs: %v", err)
	}

	return &proto.DeleteUserURLsResponse{
		Success: resp.Success,
	}, nil
}

// Ping реализует gRPC метод Ping
func (s *Server) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	serviceReq := service.PingRequest{}

	resp, err := s.urlService.Ping(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to ping: %v", err)
	}

	return &proto.PingResponse{
		Healthy: resp.Healthy,
		Message: resp.Message,
	}, nil
}

// GetStats реализует gRPC метод GetStats
func (s *Server) GetStats(ctx context.Context, req *proto.GetStatsRequest) (*proto.GetStatsResponse, error) {
	serviceReq := service.GetStatsRequest{}

	resp, err := s.urlService.GetStats(ctx, serviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
	}

	return &proto.GetStatsResponse{
		Urls:  int32(resp.URLs),
		Users: int32(resp.Users),
	}, nil
}
