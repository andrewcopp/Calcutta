package dtos

import "github.com/andrewcopp/Calcutta/backend/internal/models"

type SchoolResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewSchoolResponse(s *models.School) *SchoolResponse {
	return &SchoolResponse{ID: s.ID, Name: s.Name}
}

func NewSchoolListResponse(schools []models.School) []*SchoolResponse {
	responses := make([]*SchoolResponse, 0, len(schools))
	for _, school := range schools {
		school := school
		responses = append(responses, NewSchoolResponse(&school))
	}
	return responses
}
