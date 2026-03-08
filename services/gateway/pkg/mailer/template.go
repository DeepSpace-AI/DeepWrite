package mailer

import (
	"bytes"
	"html/template"
)

func RenderHTMLTemplateString(tmpl string, data any) (string, error) {
	parsed, err := template.New("mail_html").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := parsed.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func RenderHTMLTemplateFile(path string, data any) (string, error) {
	parsed, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := parsed.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func RenderHTMLTemplateWithLayout(layoutPath, contentPath string, data any) (string, error) {
	parsed, err := template.ParseFiles(layoutPath, contentPath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := parsed.ExecuteTemplate(&buf, "base.html", data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
