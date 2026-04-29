package service

import (
	"github.com/deepwrite/submission-service/internal/models"
	"github.com/deepwrite/submission-service/internal/repository"
)

type JournalService struct {
	repo *repository.JournalRepository
}

func NewJournalService(repo *repository.JournalRepository) *JournalService {
	return &JournalService{repo: repo}
}

func (s *JournalService) List(query, category, quartile string, page, limit int) ([]models.Journal, int, error) {
	return s.repo.List(query, category, quartile, page, limit)
}

func (s *JournalService) GetByID(id int) (*models.Journal, error) {
	return s.repo.GetByID(id)
}

func (s *JournalService) GetAll() ([]models.Journal, error) {
	return s.repo.GetAll()
}

func (s *JournalService) GetCategories() ([]string, error) {
	return s.repo.GetCategories()
}

func (s *JournalService) Recommend(title, abstract string, keywords []string) ([]models.Journal, error) {
	allJournals, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	scored := []struct {
		journal models.Journal
		score   float64
	}{}

	for _, j := range allJournals {
		score := 0.0

		for _, kw := range keywords {
			for _, jkw := range j.Keywords {
				if kw == jkw {
					score += 10.0
				}
				if contains(kw, jkw) || contains(jkw, kw) {
					score += 5.0
				}
			}
		}

		if contains(title, j.Category) || contains(j.Category, title) {
			score += 8.0
		}
		if contains(title, j.Subcategory) || contains(j.Subcategory, title) {
			score += 6.0
		}

		score += float64(j.ImpactFactor) * 0.1

		if score > 0 {
			scored = append(scored, struct {
				journal models.Journal
				score   float64
			}{j, score})
		}
	}

	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	result := []models.Journal{}
	limit := 10
	if len(scored) < limit {
		limit = len(scored)
	}
	for i := 0; i < limit; i++ {
		result = append(result, scored[i].journal)
	}

	return result, nil
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
