package ris

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/deepwrite/serivces/gateway/models/reference"
)

type RISParser struct {
	entries []RISEntry
}

type RISEntry struct {
	Type     string            `json:"type"`
	Fields   map[string]string `json:"fields"`
	Authors  []string          `json:"authors"`
	Keywords []string          `json:"keywords"`
}

func NewRISParser() *RISParser {
	return &RISParser{
		entries: make([]RISEntry, 0),
	}
}

func (p *RISParser) Parse(content string) []RISEntry {
	p.entries = make([]RISEntry, 0)

	lines := strings.Split(content, "\n")
	var currentEntry *RISEntry

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "TY  - ") {
			currentEntry = &RISEntry{
				Type:     strings.TrimPrefix(line, "TY  - "),
				Fields:   make(map[string]string),
				Authors:  make([]string, 0),
				Keywords: make([]string, 0),
			}
			continue
		}

		if line == "ER  -" {
			if currentEntry != nil {
				p.entries = append(p.entries, *currentEntry)
				currentEntry = nil
			}
			continue
		}

		if currentEntry == nil {
			continue
		}

		if len(line) < 6 {
			continue
		}

		tag := line[0:2]
		value := strings.TrimSpace(line[6:])

		switch tag {
		case "AU", "A1", "A2", "A3", "A4":
			currentEntry.Authors = append(currentEntry.Authors, value)
		case "KW":
			currentEntry.Keywords = append(currentEntry.Keywords, value)
		default:
			currentEntry.Fields[tag] = value
		}
	}

	return p.entries
}

func (p *RISParser) ToReference(entry RISEntry) *reference.Reference {
	ref := &reference.Reference{
		Title:     entry.Fields["T1"],
		DOI:       entry.Fields["DO"],
		URL:       entry.Fields["UR"],
		Abstract:  entry.Fields["AB"],
		Volume:    entry.Fields["VL"],
		Issue:     entry.Fields["IS"],
		Pages:     entry.Fields["SP"],
		Publisher: entry.Fields["PB"],
		Language:  entry.Fields["LA"],
	}

	if entry.Fields["T1"] == "" {
		ref.Title = entry.Fields["TI"]
	}

	if entry.Fields["EP"] != "" && ref.Pages != "" {
		ref.Pages = ref.Pages + "-" + entry.Fields["EP"]
	}

	if yearStr := entry.Fields["PY"]; yearStr != "" {
		if year, err := strconv.Atoi(extractYear(yearStr)); err == nil {
			ref.Year = &year
		}
	}

	ref.Source = entry.Fields["J2"]
	if ref.Source == "" {
		ref.Source = entry.Fields["JO"]
	}
	if ref.Source == "" {
		ref.Source = entry.Fields["T2"]
	}

	if len(entry.Authors) > 0 {
		authors := make([]reference.Author, len(entry.Authors))
		for i, name := range entry.Authors {
			authors[i] = parseAuthorName(name)
		}
		authorsJSON, _ := json.Marshal(authors)
		ref.Authors = authorsJSON
	}

	if len(entry.Keywords) > 0 {
		keywordsJSON, _ := json.Marshal(entry.Keywords)
		ref.Keywords = keywordsJSON
	}

	ref.Type = mapRISType(entry.Type)

	return ref
}

func parseAuthorName(name string) reference.Author {
	name = strings.TrimSpace(name)

	if strings.Contains(name, ",") {
		parts := strings.SplitN(name, ",", 2)
		return reference.Author{
			Family: strings.TrimSpace(parts[0]),
			Given:  strings.TrimSpace(parts[1]),
		}
	}

	parts := strings.Fields(name)
	if len(parts) >= 2 {
		return reference.Author{
			Given:  strings.Join(parts[:len(parts)-1], " "),
			Family: parts[len(parts)-1],
		}
	}

	return reference.Author{
		Literal: name,
	}
}

func extractYear(s string) string {
	re := regexp.MustCompile(`\d{4}`)
	match := re.FindString(s)
	return match
}

func mapRISType(risType string) reference.ReferenceType {
	switch strings.ToUpper(risType) {
	case "JOUR", "JFULL":
		return reference.ReferenceTypeArticle
	case "BOOK":
		return reference.ReferenceTypeBook
	case "CHAP":
		return reference.ReferenceTypeBookChapter
	case "CONF", "CPAPER":
		return reference.ReferenceTypeConference
	case "THES", "DISS":
		return reference.ReferenceTypeThesis
	case "RPRT", "TECH":
		return reference.ReferenceTypeReport
	case "ELEC", "ONLINE":
		return reference.ReferenceTypeWeb
	case "UNPB":
		return reference.ReferenceTypePreprint
	default:
		return reference.ReferenceTypeUnknown
	}
}

type RISExporter struct{}

func NewRISExporter() *RISExporter {
	return &RISExporter{}
}

func (e *RISExporter) Export(refs []reference.Reference) string {
	var sb strings.Builder

	for _, ref := range refs {
		entry := e.referenceToRIS(&ref)
		sb.WriteString(entry)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (e *RISExporter) referenceToRIS(ref *reference.Reference) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("TY  - %s\n", e.mapReferenceType(ref.Type)))

	if ref.Title != "" {
		sb.WriteString(fmt.Sprintf("TI  - %s\n", ref.Title))
	}

	if len(ref.Authors) > 0 {
		var authors []reference.Author
		json.Unmarshal(ref.Authors, &authors)
		for _, a := range authors {
			name := a.Literal
			if name == "" {
				if a.Family != "" && a.Given != "" {
					name = fmt.Sprintf("%s, %s", a.Family, a.Given)
				} else if a.Family != "" {
					name = a.Family
				}
			}
			if name != "" {
				sb.WriteString(fmt.Sprintf("AU  - %s\n", name))
			}
		}
	}

	if ref.Year != nil {
		sb.WriteString(fmt.Sprintf("PY  - %d\n", *ref.Year))
	}

	if ref.Source != "" {
		sb.WriteString(fmt.Sprintf("JO  - %s\n", ref.Source))
	}

	if ref.Volume != "" {
		sb.WriteString(fmt.Sprintf("VL  - %s\n", ref.Volume))
	}

	if ref.Issue != "" {
		sb.WriteString(fmt.Sprintf("IS  - %s\n", ref.Issue))
	}

	if ref.Pages != "" {
		pages := strings.Split(ref.Pages, "-")
		sb.WriteString(fmt.Sprintf("SP  - %s\n", pages[0]))
		if len(pages) > 1 {
			sb.WriteString(fmt.Sprintf("EP  - %s\n", pages[1]))
		}
	}

	if ref.Publisher != "" {
		sb.WriteString(fmt.Sprintf("PB  - %s\n", ref.Publisher))
	}

	if ref.DOI != "" {
		sb.WriteString(fmt.Sprintf("DO  - %s\n", ref.DOI))
	}

	if ref.URL != "" {
		sb.WriteString(fmt.Sprintf("UR  - %s\n", ref.URL))
	}

	if ref.Abstract != "" {
		sb.WriteString(fmt.Sprintf("AB  - %s\n", ref.Abstract))
	}

	if len(ref.Keywords) > 0 {
		var keywords []string
		json.Unmarshal(ref.Keywords, &keywords)
		for _, kw := range keywords {
			sb.WriteString(fmt.Sprintf("KW  - %s\n", kw))
		}
	}

	sb.WriteString("ER  -\n")

	return sb.String()
}

func (e *RISExporter) mapReferenceType(refType reference.ReferenceType) string {
	switch refType {
	case reference.ReferenceTypeArticle:
		return "JOUR"
	case reference.ReferenceTypeBook:
		return "BOOK"
	case reference.ReferenceTypeBookChapter:
		return "CHAP"
	case reference.ReferenceTypeConference:
		return "CONF"
	case reference.ReferenceTypeThesis:
		return "THES"
	case reference.ReferenceTypeReport:
		return "RPRT"
	case reference.ReferenceTypeWeb:
		return "ELEC"
	case reference.ReferenceTypePreprint:
		return "UNPB"
	default:
		return "GEN"
	}
}
