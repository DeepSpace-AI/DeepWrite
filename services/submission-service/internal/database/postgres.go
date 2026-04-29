package database

import (
	"database/sql"
	"fmt"

	"github.com/deepwrite/submission-service/internal/config"
	"github.com/lib/pq"
)

func NewPostgres(cfg *config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func Migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS journals (
			id SERIAL PRIMARY KEY,
			name VARCHAR(500) NOT NULL,
			publisher VARCHAR(255),
			issn VARCHAR(50),
			category VARCHAR(255),
			subcategory VARCHAR(255),
			impact_factor DECIMAL(5,3),
			quartile VARCHAR(10),
			open_access BOOLEAN DEFAULT FALSE,
			website_url TEXT,
			review_time_days INT,
			acceptance_rate DECIMAL(5,2),
			keywords TEXT[],
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS submissions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL,
			document_id UUID,
			user_id UUID NOT NULL,
			journal_id INT REFERENCES journals(id),
			title TEXT NOT NULL,
			abstract TEXT,
			keywords TEXT[],
			status VARCHAR(50) DEFAULT 'draft',
			stage VARCHAR(50) DEFAULT 'preparation',
			manuscript_url TEXT,
			supplementary_files JSONB DEFAULT '[]',
			submission_date TIMESTAMP,
			last_status_change TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			notes TEXT,
			recommendation_score DECIMAL(5,2),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS submission_reviews (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
			reviewer_name VARCHAR(255),
			review_type VARCHAR(50) DEFAULT 'external',
			status VARCHAR(50) DEFAULT 'pending',
			content TEXT,
			rating INT CHECK (rating >= 1 AND rating <= 5),
			recommendation VARCHAR(50),
			received_at TIMESTAMP,
			responded_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS submission_history (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
			from_status VARCHAR(50),
			to_status VARCHAR(50),
			note TEXT,
			created_by UUID,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_project_id ON submissions(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_user_id ON submissions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_status ON submissions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_journals_category ON journals(category)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_submission_id ON submission_reviews(submission_id)`,
	}

	for i, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	// Seed sample journals if table is empty
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM journals").Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if err := seedJournals(db); err != nil {
			return fmt.Errorf("failed to seed journals: %w", err)
		}
	}

	return nil
}

func seedJournals(db *sql.DB) error {
	journals := []struct {
		name            string
		publisher       string
		issn            string
		category        string
		subcategory     string
		impactFactor    float64
		quartile        string
		openAccess      bool
		websiteURL      string
		reviewTimeDays  int
		acceptanceRate  float64
		keywords        []string
	}{
		{"Nature", "Nature Publishing Group", "0028-0836", "Multidisciplinary", "General Science", 49.962, "Q1", false, "https://www.nature.com", 180, 8.0, []string{"multidisciplinary", "general science", "high impact"}},
		{"Science", "American Association for the Advancement of Science", "0036-8075", "Multidisciplinary", "General Science", 47.728, "Q1", false, "https://www.science.org", 150, 7.0, []string{"multidisciplinary", "general science", "high impact"}},
		{"Cell", "Cell Press", "0092-8674", "Biology", "Molecular Biology", 45.5, "Q1", false, "https://www.cell.com", 120, 10.0, []string{"biology", "molecular biology", "cell biology"}},
		{"IEEE Transactions on Pattern Analysis and Machine Intelligence", "IEEE", "0162-8828", "Computer Science", "Artificial Intelligence", 24.314, "Q1", false, "https://ieeexplore.ieee.org", 90, 15.0, []string{"computer vision", "machine learning", "pattern recognition"}},
		{"Nature Machine Intelligence", "Nature Publishing Group", "2522-5839", "Computer Science", "Artificial Intelligence", 25.0, "Q1", false, "https://www.nature.com/natmachintell", 90, 12.0, []string{"machine learning", "artificial intelligence", "robotics"}},
		{"The Lancet", "Elsevier", "0140-6736", "Medicine", "General Medicine", 79.321, "Q1", false, "https://www.thelancet.com", 60, 5.0, []string{"medicine", "clinical research", "public health"}},
		{"Journal of the American Medical Association", "American Medical Association", "0098-7484", "Medicine", "General Medicine", 56.272, "Q1", false, "https://jamanetwork.com", 45, 6.0, []string{"medicine", "clinical trials", "healthcare"}},
		{"Physical Review Letters", "American Physical Society", "0031-9007", "Physics", "General Physics", 8.385, "Q1", false, "https://journals.aps.org/prl", 60, 25.0, []string{"physics", "condensed matter", "particle physics"}},
		{"ACS Nano", "American Chemical Society", "1936-0851", "Chemistry", "Nanoscience", 17.1, "Q1", false, "https://pubs.acs.org/journal/ancac3", 45, 15.0, []string{"nanotechnology", "materials science", "chemistry"}},
		{"Advanced Materials", "Wiley", "0935-9648", "Materials Science", "Multidisciplinary", 29.4, "Q1", false, "https://onlinelibrary.wiley.com/journal/15214095", 30, 20.0, []string{"materials science", "nanotechnology", "energy"}},
		{"PLOS ONE", "Public Library of Science", "1932-6203", "Multidisciplinary", "General Science", 3.7, "Q2", true, "https://journals.plos.org/plosone", 90, 60.0, []string{"multidisciplinary", "open access", "general science"}},
		{"Scientific Reports", "Nature Publishing Group", "2045-2322", "Multidisciplinary", "General Science", 4.6, "Q2", true, "https://www.nature.com/srep", 60, 50.0, []string{"multidisciplinary", "open access", "general science"}},
		{"IEEE Access", "IEEE", "2169-3536", "Engineering", "Multidisciplinary", 3.9, "Q2", true, "https://ieeexplore.ieee.org/xpl/RecentIssue.jsp?punumber=6287639", 45, 70.0, []string{"engineering", "open access", "multidisciplinary"}},
		{"Frontiers in Artificial Intelligence", "Frontiers Media", "2624-8212", "Computer Science", "Artificial Intelligence", 3.0, "Q2", true, "https://www.frontiersin.org/journals/artificial-intelligence", 60, 45.0, []string{"artificial intelligence", "machine learning", "open access"}},
		{"Bioinformatics", "Oxford University Press", "1367-4803", "Biology", "Computational Biology", 4.4, "Q1", false, "https://academic.oup.com/bioinformatics", 45, 25.0, []string{"bioinformatics", "computational biology", "genomics"}},
		{"Neural Networks", "Elsevier", "0893-6080", "Computer Science", "Neural Networks", 7.8, "Q1", false, "https://www.journals.elsevier.com/neural-networks", 90, 20.0, []string{"neural networks", "deep learning", "machine learning"}},
		{"Journal of Clinical Oncology", "American Society of Clinical Oncology", "0732-183X", "Medicine", "Oncology", 50.7, "Q1", false, "https://ascopubs.org/journal/jco", 30, 10.0, []string{"oncology", "clinical trials", "cancer research"}},
		{"Angewandte Chemie", "Wiley", "1433-7851", "Chemistry", "General Chemistry", 16.6, "Q1", false, "https://onlinelibrary.wiley.com/journal/15213773", 30, 18.0, []string{"chemistry", "materials", "catalysis"}},
	}

	for _, j := range journals {
		_, err := db.Exec(
			`INSERT INTO journals (name, publisher, issn, category, subcategory, impact_factor, quartile, open_access, website_url, review_time_days, acceptance_rate, keywords)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			j.name, j.publisher, j.issn, j.category, j.subcategory, j.impactFactor, j.quartile, j.openAccess, j.websiteURL, j.reviewTimeDays, j.acceptanceRate, pq.Array(j.keywords),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
