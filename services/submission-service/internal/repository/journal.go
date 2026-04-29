package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/deepwrite/submission-service/internal/models"
)

type JournalRepository struct {
	db *sql.DB
}

func NewJournalRepository(db *sql.DB) *JournalRepository {
	return &JournalRepository{db: db}
}

func (r *JournalRepository) List(query, category, quartile string, page, limit int) ([]models.Journal, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if query != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR keywords @> ARRAY[$%d] OR publisher ILIKE $%d)", argIdx, argIdx+1, argIdx+2))
		args = append(args, "%"+query+"%", query, "%"+query+"%")
		argIdx += 3
	}
	if category != "" {
		where = append(where, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}
	if quartile != "" {
		where = append(where, fmt.Sprintf("quartile = $%d", argIdx))
		args = append(args, quartile)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := "SELECT COUNT(*) FROM journals WHERE " + whereClause
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf(
		"SELECT id, name, publisher, issn, category, subcategory, impact_factor, quartile, open_access, website_url, review_time_days, acceptance_rate, keywords, created_at, updated_at FROM journals WHERE %s ORDER BY impact_factor DESC LIMIT $%d OFFSET $%d",
		whereClause, argIdx, argIdx+1,
	)
	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var journals []models.Journal
	for rows.Next() {
		var j models.Journal
		if err := rows.Scan(&j.ID, &j.Name, &j.Publisher, &j.ISSN, &j.Category, &j.Subcategory, &j.ImpactFactor, &j.Quartile, &j.OpenAccess, &j.WebsiteURL, &j.ReviewTimeDays, &j.AcceptanceRate, &j.Keywords, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, 0, err
		}
		journals = append(journals, j)
	}
	return journals, total, nil
}

func (r *JournalRepository) GetByID(id int) (*models.Journal, error) {
	var j models.Journal
	err := r.db.QueryRow(
		"SELECT id, name, publisher, issn, category, subcategory, impact_factor, quartile, open_access, website_url, review_time_days, acceptance_rate, keywords, created_at, updated_at FROM journals WHERE id = $1",
		id,
	).Scan(&j.ID, &j.Name, &j.Publisher, &j.ISSN, &j.Category, &j.Subcategory, &j.ImpactFactor, &j.Quartile, &j.OpenAccess, &j.WebsiteURL, &j.ReviewTimeDays, &j.AcceptanceRate, &j.Keywords, &j.CreatedAt, &j.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, models.ErrJournalNotFound
	}
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *JournalRepository) GetAll() ([]models.Journal, error) {
	rows, err := r.db.Query("SELECT id, name, publisher, issn, category, subcategory, impact_factor, quartile, open_access, website_url, review_time_days, acceptance_rate, keywords, created_at, updated_at FROM journals ORDER BY impact_factor DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var journals []models.Journal
	for rows.Next() {
		var j models.Journal
		if err := rows.Scan(&j.ID, &j.Name, &j.Publisher, &j.ISSN, &j.Category, &j.Subcategory, &j.ImpactFactor, &j.Quartile, &j.OpenAccess, &j.WebsiteURL, &j.ReviewTimeDays, &j.AcceptanceRate, &j.Keywords, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		journals = append(journals, j)
	}
	return journals, nil
}

func (r *JournalRepository) GetCategories() ([]string, error) {
	rows, err := r.db.Query("SELECT DISTINCT category FROM journals ORDER BY category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}
