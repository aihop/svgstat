package renderer

import (
	"fmt"
	"html/template"
	"strings"
)

type Renderer struct {
	templates *template.Template
}

type CounterData struct {
	Value  int64
	Label  string
	Theme  string
	Color  string
}

type BadgeData struct {
	Label  string
	Value  string
	Color  string
	Style  string
}

func New() *Renderer {
	return &Renderer{
		templates: loadTemplates(),
	}
}

func loadTemplates() *template.Template {
	tmpl := template.New("svg")

	tmpl = template.Must(tmpl.New("counter").Parse(counterTemplate))
	tmpl = template.Must(tmpl.New("badge").Parse(badgeTemplate))
	tmpl = template.Must(tmpl.New("badge_flat").Parse(badgeFlatTemplate))

	return tmpl
}

func (r *Renderer) RenderCounter(data CounterData) (string, error) {
	if data.Theme == "" {
		data.Theme = "default"
	}
	if data.Color == "" {
		data.Color = "#007bff"
	}

	var buf strings.Builder
	err := r.templates.ExecuteTemplate(&buf, "counter", data)
	if err != nil {
		return "", fmt.Errorf("failed to render counter: %w", err)
	}

	return buf.String(), nil
}

func (r *Renderer) RenderBadge(data BadgeData) (string, error) {
	if data.Color == "" {
		data.Color = "#007bff"
	}
	if data.Style == "" {
		data.Style = "default"
	}

	templateName := "badge"
	if data.Style == "flat" {
		templateName = "badge_flat"
	}

	var buf strings.Builder
	err := r.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		return "", fmt.Errorf("failed to render badge: %w", err)
	}

	return buf.String(), nil
}

const counterTemplate = `
<svg xmlns="http://www.w3.org/2000/svg" width="120" height="40" viewBox="0 0 120 40">
  <style>
    .background { fill: #f8f9fa; }
    .text { font-family: 'Arial', sans-serif; font-size: 20px; font-weight: bold; }
    .label { font-size: 14px; fill: #6c757d; }
    .value { fill: {{.Color}}; }
  </style>
  <rect class="background" width="120" height="40" rx="8" ry="8"/>
  {{if .Label}}
  <text class="text label" x="10" y="17">{{.Label}}</text>
  <text class="text value" x="10" y="36">{{.Value}}</text>
  {{else}}
  <text class="text value" x="60" y="27" text-anchor="middle">{{.Value}}</text>
  {{end}}
</svg>
`

const badgeTemplate = `
<svg xmlns="http://www.w3.org/2000/svg" width="140" height="28" viewBox="0 0 140 28">
  <style>
    .label-bg { fill: #555; }
    .value-bg { fill: {{.Color}}; }
    .text { font-family: 'Arial', sans-serif; font-size: 12px; font-weight: bold; fill: white; }
  </style>
  <rect class="label-bg" width="60" height="28"/>
  <rect class="value-bg" x="60" width="80" height="28"/>
  <rect x="58" width="4" height="28" fill="{{.Color}}"/>
  <text class="text" x="30" y="18" text-anchor="middle">{{.Label}}</text>
  <text class="text" x="100" y="18" text-anchor="middle">{{.Value}}</text>
</svg>
`

const badgeFlatTemplate = `
<svg xmlns="http://www.w3.org/2000/svg" width="140" height="28" viewBox="0 0 140 28">
  <style>
    .label-bg { fill: #555; }
    .value-bg { fill: {{.Color}}; }
    .text { font-family: 'Arial', sans-serif; font-size: 12px; font-weight: normal; fill: white; }
  </style>
  <rect class="label-bg" width="60" height="28"/>
  <rect class="value-bg" x="60" width="80" height="28"/>
  <text class="text" x="30" y="18" text-anchor="middle">{{.Label}}</text>
  <text class="text" x="100" y="18" text-anchor="middle">{{.Value}}</text>
</svg>
`
