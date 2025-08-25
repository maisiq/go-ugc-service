package handler

import (
	"context"

	"errors"

	apperrors "github.com/maisiq/go-ugc-service/internal/errors"
	"github.com/maisiq/go-ugc-service/internal/mapper"
	"github.com/maisiq/go-ugc-service/internal/service"
	ugcv1pb "github.com/maisiq/go-ugc-service/pkg/pb/ugcservice/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UGCServiceServer struct {
	ugcv1pb.UnimplementedUGCServiceServer
	service *service.UGCService
}

func NewServer(service *service.UGCService) *UGCServiceServer {
	return &UGCServiceServer{
		service: service,
	}
}

func (s *UGCServiceServer) CreateReview(ctx context.Context, req *ugcv1pb.CreateReviewRequest) (*emptypb.Empty, error) {
	var empty emptypb.Empty

	err := s.service.CreateReview(ctx, req.Review.UserId, req.Review.MovieId, req.Review.Text)

	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrAlreadyExists):
			return nil, status.Errorf(codes.AlreadyExists, "this review already exists")
		default:
			return nil, status.Errorf(codes.Internal, "internal error")
		}
	}

	return &empty, nil
}

func (s *UGCServiceServer) GetReviews(ctx context.Context, req *ugcv1pb.GetReviewsRequest) (*ugcv1pb.GetReviewsResponse, error) {
	reviews, err := s.service.GetReviews(ctx, req.GetUserId(), req.GetMovieId())

	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			return nil, status.Errorf(codes.NotFound, "not found")
		default:
			return nil, status.Errorf(codes.Internal, "internal error")
		}
	}
	return mapper.FromReviewToPb(reviews), nil
}

func (s *UGCServiceServer) UpdateReview(ctx context.Context, req *ugcv1pb.UpdateReviewRequest) (*emptypb.Empty, error) {
	var empty emptypb.Empty

	err := s.service.UpdateReview(ctx, req.Review.GetUserId(), req.Review.GetMovieId(), req.Review.GetText())

	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			return &empty, status.Errorf(codes.NotFound, "could not find the review with this params")
		default:
			return &empty, status.Error(codes.Internal, "internal error")
		}
	}

	return &empty, nil
}
