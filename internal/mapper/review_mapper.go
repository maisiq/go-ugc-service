package mapper

import (
	"github.com/maisiq/go-ugc-service/internal/repository"
	ugcv1pb "github.com/maisiq/go-ugc-service/pkg/pb/ugcservice/v1"
)

func FromReviewToPb(reviews []repository.Review) *ugcv1pb.GetReviewsResponse {
	var reviewsPb []*ugcv1pb.Review

	for _, review := range reviews {
		reviewsPb = append(reviewsPb, &ugcv1pb.Review{
			MovieId: review.MovieID,
			UserId:  review.UserID,
			Text:    review.Text,
		})
	}
	response := ugcv1pb.GetReviewsResponse{
		Reviews: reviewsPb,
	}

	return &response
}
