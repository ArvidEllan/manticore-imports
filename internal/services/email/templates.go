package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

type QuoteReceivedData struct {
	CustomerName  string
	Reference     string
	ProductName   string
	Quantity      int
	SourceCountry string
}

type StatusUpdatedData struct {
	CustomerName string
	Reference    string
	Status       string
}

func RenderQuoteReceived(data QuoteReceivedData) (html, text string, err error) {
	html, err = render("templates/quote_received.html", data)
	if err != nil {
		return "", "", err
	}
	text = fmt.Sprintf(
		"Hi %s,\n\nYour import quote request %s has been received.\nProduct: %s\nQuantity: %d\nSource country: %s\n\nTrack your request using your reference and email.",
		data.CustomerName, data.Reference, data.ProductName, data.Quantity, data.SourceCountry,
	)
	return html, text, nil
}

func RenderStatusUpdated(data StatusUpdatedData) (html, text string, err error) {
	html, err = render("templates/status_updated.html", data)
	if err != nil {
		return "", "", err
	}
	text = fmt.Sprintf(
		"Hi %s,\n\nYour import request %s status has been updated to: %s",
		data.CustomerName, data.Reference, data.Status,
	)
	return html, text, nil
}

func render(name string, data any) (string, error) {
	tmpl, err := template.ParseFS(templateFS, name)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", name, err)
	}
	return buf.String(), nil
}
