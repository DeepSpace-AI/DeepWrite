package bibtex

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/models/reference"
)

type Entry struct {
	Type       string            `json:"type"`
	Key        string            `json:"key"`
	Fields     map[string]string `json:"fields"`
	RawContent string            `json:"raw_content"`
}

type Parser struct {
	entries []Entry
}

func NewParser() *Parser {
	return &Parser{
		entries: make([]Entry, 0),
	}
}

func (p *Parser) Parse(content string) []Entry {
	p.entries = make([]Entry, 0)

	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	re := regexp.MustCompile(`@(\w+)\s*\{\s*([^,]+)\s*,`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	for i, match := range matches {
		entryType := strings.ToLower(content[match[2]:match[3]])
		key := strings.TrimSpace(content[match[4]:match[5]])

		var endPos int
		if i < len(matches)-1 {
			endPos = matches[i+1][0]
		} else {
			endPos = len(content)
		}

		entryContent := content[match[1]:endPos]
		fields := p.parseFields(entryContent)

		p.entries = append(p.entries, Entry{
			Type:       entryType,
			Key:        key,
			Fields:     fields,
			RawContent: entryContent,
		})
	}

	return p.entries
}

func (p *Parser) parseFields(content string) map[string]string {
	fields := make(map[string]string)

	braceCount := 0
	inField := false
	currentField := ""
	currentValue := ""
	quoteChar := byte(0)

	for i := 0; i < len(content); i++ {
		c := content[i]

		if c == '{' {
			braceCount++
			if inField && braceCount == 1 {
				continue
			}
		} else if c == '}' {
			braceCount--
			if inField && braceCount == 0 {
				fields[strings.ToLower(strings.TrimSpace(currentField))] = strings.TrimSpace(currentValue)
				currentField = ""
				currentValue = ""
				inField = false
				continue
			}
		}

		if braceCount == 0 && c == '=' && !inField {
			inField = true
			braceCount = 0
			continue
		}

		if braceCount == 0 && c == ',' && !inField {
			continue
		}

		if !inField && c != '@' && c != '{' && c != '}' && c != ',' && c != '=' && c != '\n' && c != '\r' && c != ' ' && c != '\t' {
			currentField += string(c)
		} else if !inField && (c == ' ' || c == '\n' || c == '\t') && currentField != "" {
		} else if inField && braceCount > 0 {
			if quoteChar == 0 && (c == '"' || c == '\'') {
				quoteChar = c
				continue
			}
			if quoteChar != 0 && c == quoteChar && braceCount == 1 {
				quoteChar = 0
				continue
			}
			currentValue += string(c)
		}
	}

	return fields
}

func (p *Parser) Entries() []Entry {
	return p.entries
}

func (p *Parser) ToReference(entry Entry) *reference.Reference {
	ref := &reference.Reference{
		Title:       cleanBraces(entry.Fields["title"]),
		DOI:         entry.Fields["doi"],
		URL:         entry.Fields["url"],
		Abstract:    cleanBraces(entry.Fields["abstract"]),
		Volume:      entry.Fields["volume"],
		Issue:       entry.Fields["number"],
		Pages:       entry.Fields["pages"],
		Publisher:   entry.Fields["publisher"],
		Language:    entry.Fields["language"],
		CitationKey: entry.Key,
		BibtexRaw:   entry.RawContent,
	}

	if yearStr := entry.Fields["year"]; yearStr != "" {
		if year, err := strconv.Atoi(strings.TrimSpace(cleanBraces(yearStr))); err == nil {
			ref.Year = &year
		}
	}

	ref.Source = cleanBraces(entry.Fields["journal"])
	if ref.Source == "" {
		ref.Source = cleanBraces(entry.Fields["booktitle"])
	}
	if ref.Source == "" {
		ref.Source = cleanBraces(entry.Fields["publisher"])
	}

	authors := parseAuthors(entry.Fields["author"])
	if len(authors) > 0 {
		authorsJSON, _ := json.Marshal(authors)
		ref.Authors = authorsJSON
	}

	ref.Type = mapBibtexType(entry.Type)

	if keywords := entry.Fields["keywords"]; keywords != "" {
		kwList := strings.Split(cleanBraces(keywords), ",")
		for i, kw := range kwList {
			kwList[i] = strings.TrimSpace(kw)
		}
		if len(kwList) > 0 {
			keywordsJSON, _ := json.Marshal(kwList)
			ref.Keywords = keywordsJSON
		}
	}

	return ref
}

func parseAuthors(authorStr string) []reference.Author {
	authorStr = cleanBraces(authorStr)
	if authorStr == "" {
		return nil
	}

	parts := strings.Split(authorStr, " and ")
	authors := make([]reference.Author, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		author := reference.Author{}

		if strings.Contains(part, ",") {
			p := strings.SplitN(part, ",", 2)
			author.Family = strings.TrimSpace(p[0])
			if len(p) > 1 {
				author.Given = strings.TrimSpace(p[1])
			}
		} else if strings.Contains(part, " ") {
			p := strings.Split(part, " ")
			if len(p) >= 2 {
				author.Given = strings.Join(p[:len(p)-1], " ")
				author.Family = p[len(p)-1]
			} else {
				author.Family = part
			}
		} else {
			author.Literal = part
		}

		authors = append(authors, author)
	}

	return authors
}

func mapBibtexType(bibtexType string) reference.ReferenceType {
	switch strings.ToLower(bibtexType) {
	case "article":
		return reference.ReferenceTypeArticle
	case "book", "booklet":
		return reference.ReferenceTypeBook
	case "inbook", "incollection":
		return reference.ReferenceTypeBookChapter
	case "inproceedings", "conference":
		return reference.ReferenceTypeConference
	case "phdthesis", "mastersthesis":
		return reference.ReferenceTypeThesis
	case "techreport", "manual":
		return reference.ReferenceTypeReport
	case "misc", "unpublished":
		return reference.ReferenceTypePreprint
	case "online":
		return reference.ReferenceTypeWeb
	default:
		return reference.ReferenceTypeUnknown
	}
}

func cleanBraces(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	return strings.TrimSpace(s)
}

type Exporter struct{}

func NewExporter() *Exporter {
	return &Exporter{}
}

func (e *Exporter) Export(refs []reference.Reference) string {
	var sb strings.Builder

	for _, ref := range refs {
		entry := e.referenceToBibtex(&ref)
		sb.WriteString(entry)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func (e *Exporter) referenceToBibtex(ref *reference.Reference) string {
	var sb strings.Builder

	entryType := e.mapReferenceType(ref.Type)
	key := ref.CitationKey
	if key == "" {
		key = e.generateKey(ref)
	}

	sb.WriteString(fmt.Sprintf("@%s{%s,\n", entryType, key))

	if ref.Title != "" {
		sb.WriteString(fmt.Sprintf("  title = {%s},\n", ref.Title))
	}

	if len(ref.Authors) > 0 {
		var authors []reference.Author
		json.Unmarshal(ref.Authors, &authors)
		if len(authors) > 0 {
			authorStr := e.formatAuthors(authors)
			sb.WriteString(fmt.Sprintf("  author = {%s},\n", authorStr))
		}
	}

	if ref.Year != nil {
		sb.WriteString(fmt.Sprintf("  year = {%d},\n", *ref.Year))
	}

	if ref.Source != "" {
		if ref.Type == reference.ReferenceTypeArticle {
			sb.WriteString(fmt.Sprintf("  journal = {%s},\n", ref.Source))
		} else if ref.Type == reference.ReferenceTypeConference {
			sb.WriteString(fmt.Sprintf("  booktitle = {%s},\n", ref.Source))
		} else if ref.Type == reference.ReferenceTypeBook {
			sb.WriteString(fmt.Sprintf("  publisher = {%s},\n", ref.Source))
		}
	}

	if ref.Volume != "" {
		sb.WriteString(fmt.Sprintf("  volume = {%s},\n", ref.Volume))
	}

	if ref.Issue != "" {
		sb.WriteString(fmt.Sprintf("  number = {%s},\n", ref.Issue))
	}

	if ref.Pages != "" {
		sb.WriteString(fmt.Sprintf("  pages = {%s},\n", ref.Pages))
	}

	if ref.DOI != "" {
		sb.WriteString(fmt.Sprintf("  doi = {%s},\n", ref.DOI))
	}

	if ref.URL != "" {
		sb.WriteString(fmt.Sprintf("  url = {%s},\n", ref.URL))
	}

	if ref.Abstract != "" {
		abstract := strings.ReplaceAll(ref.Abstract, "\n", " ")
		sb.WriteString(fmt.Sprintf("  abstract = {%s},\n", abstract))
	}

	sb.WriteString("}")

	return sb.String()
}

func (e *Exporter) mapReferenceType(refType reference.ReferenceType) string {
	switch refType {
	case reference.ReferenceTypeArticle:
		return "article"
	case reference.ReferenceTypeBook:
		return "book"
	case reference.ReferenceTypeBookChapter:
		return "incollection"
	case reference.ReferenceTypeConference:
		return "inproceedings"
	case reference.ReferenceTypeThesis:
		return "phdthesis"
	case reference.ReferenceTypeReport:
		return "techreport"
	case reference.ReferenceTypeWeb:
		return "online"
	case reference.ReferenceTypePreprint:
		return "unpublished"
	default:
		return "misc"
	}
}

func (e *Exporter) formatAuthors(authors []reference.Author) string {
	parts := make([]string, len(authors))
	for i, a := range authors {
		if a.Literal != "" {
			parts[i] = a.Literal
		} else if a.Family != "" && a.Given != "" {
			parts[i] = fmt.Sprintf("%s, %s", a.Family, a.Given)
		} else if a.Family != "" {
			parts[i] = a.Family
		}
	}
	return strings.Join(parts, " and ")
}

func (e *Exporter) generateKey(ref *reference.Reference) string {
	parts := make([]string, 0)

	var authors []reference.Author
	if len(ref.Authors) > 0 {
		json.Unmarshal(ref.Authors, &authors)
	}

	if len(authors) > 0 {
		family := authors[0].Family
		if family == "" {
			family = authors[0].Literal
		}
		family = strings.ToLower(strings.ReplaceAll(family, " ", ""))
		family = regexp.MustCompile(`[^a-z]`).ReplaceAllString(family, "")
		if family != "" {
			parts = append(parts, family)
		}
	}

	if ref.Year != nil {
		parts = append(parts, fmt.Sprintf("%d", *ref.Year))
	}

	if ref.Title != "" {
		words := strings.Fields(ref.Title)
		for _, w := range words {
			w = strings.ToLower(w)
			w = regexp.MustCompile(`[^a-z]`).ReplaceAllString(w, "")
			if len(w) > 3 {
				parts = append(parts, w)
				break
			}
		}
	}

	if len(parts) == 0 {
		return fmt.Sprintf("ref_%d", time.Now().Unix())
	}

	return strings.Join(parts, "")
}
