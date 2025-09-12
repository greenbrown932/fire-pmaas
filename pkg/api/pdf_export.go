package api

import (
	"bytes"
	"fmt"
	"html/template"
	"os/exec"
	"strings"
	"time"

	"github.com/greenbrown932/fire-pmaas/pkg/models"
)

// PDFReportGenerator handles PDF generation for reports
type PDFReportGenerator struct {
	Template *template.Template
}

// NewPDFReportGenerator creates a new PDF report generator
func NewPDFReportGenerator() *PDFReportGenerator {
	return &PDFReportGenerator{
		Template: loadPDFTemplates(),
	}
}

// GeneratePDFReport generates a PDF from report data
func (g *PDFReportGenerator) GeneratePDFReport(reportData *models.ReportData, reportInfo *models.CustomReport) ([]byte, error) {
	// Generate HTML content from template
	htmlContent, err := g.generateHTMLContent(reportData, reportInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HTML content: %w", err)
	}

	// Convert HTML to PDF using wkhtmltopdf (if available) or basic HTML-to-PDF
	pdfData, err := g.convertHTMLToPDF(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HTML to PDF: %w", err)
	}

	return pdfData, nil
}

// generateHTMLContent creates HTML content for the PDF
func (g *PDFReportGenerator) generateHTMLContent(reportData *models.ReportData, reportInfo *models.CustomReport) (string, error) {
	data := struct {
		Report      *models.CustomReport
		Data        *models.ReportData
		GeneratedAt time.Time
		Title       string
	}{
		Report:      reportInfo,
		Data:        reportData,
		GeneratedAt: time.Now(),
		Title:       fmt.Sprintf("%s Report", reportInfo.Name),
	}

	var buf bytes.Buffer
	if err := g.Template.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// convertHTMLToPDF converts HTML content to PDF
func (g *PDFReportGenerator) convertHTMLToPDF(htmlContent string) ([]byte, error) {
	// Try to use wkhtmltopdf if available
	if isWkhtmltopdfAvailable() {
		return g.convertWithWkhtmltopdf(htmlContent)
	}

	// Fall back to simple HTML-to-PDF conversion
	return g.generateSimplePDF(htmlContent)
}

// isWkhtmltopdfAvailable checks if wkhtmltopdf is installed
func isWkhtmltopdfAvailable() bool {
	_, err := exec.LookPath("wkhtmltopdf")
	return err == nil
}

// convertWithWkhtmltopdf uses wkhtmltopdf to convert HTML to PDF
func (g *PDFReportGenerator) convertWithWkhtmltopdf(htmlContent string) ([]byte, error) {
	cmd := exec.Command("wkhtmltopdf",
		"--page-size", "A4",
		"--margin-top", "0.75in",
		"--margin-right", "0.75in",
		"--margin-bottom", "0.75in",
		"--margin-left", "0.75in",
		"--encoding", "UTF-8",
		"--print-media-type",
		"-", // read from stdin
		"-", // write to stdout
	)

	cmd.Stdin = bytes.NewReader([]byte(htmlContent))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("wkhtmltopdf error: %v, stderr: %s", err, stderr.String())
	}

	return out.Bytes(), nil
}

// generateSimplePDF creates a basic PDF without external dependencies
func (g *PDFReportGenerator) generateSimplePDF(htmlContent string) ([]byte, error) {
	// This is a simplified PDF generation for demonstration
	// In production, you'd want to use a proper PDF library like gofpdf or similar

	pdfContent := fmt.Sprintf(`%%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj

2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj

3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
/Resources <<
  /Font <<
    /F1 5 0 R
  >>
>>
>>
endobj

4 0 obj
<<
/Length %d
>>
stream
BT
/F1 12 Tf
50 750 Td
(Fire PMAAS Report) Tj
0 -20 Td
(Generated: %s) Tj
0 -40 Td
(This is a simplified PDF export.) Tj
0 -20 Td
(For full formatting, install wkhtmltopdf.) Tj
ET
endstream
endobj

5 0 obj
<<
/Type /Font
/Subtype /Type1
/BaseFont /Helvetica
>>
endobj

xref
0 6
0000000000 65535 f
0000000010 00000 n
0000000053 00000 n
0000000110 00000 n
0000000281 00000 n
0000000456 00000 n
trailer
<<
/Size 6
/Root 1 0 R
>>
startxref
543
%%%%EOF`,
		180, // approximate length of the content stream
		time.Now().Format("2006-01-02 15:04:05"),
	)

	return []byte(pdfContent), nil
}

// loadPDFTemplates loads HTML templates for PDF generation
func loadPDFTemplates() *template.Template {
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 20px;
            color: #333;
        }
        .header {
            border-bottom: 2px solid #3B82F6;
            padding-bottom: 20px;
            margin-bottom: 30px;
        }
        .header h1 {
            color: #1F2937;
            margin: 0;
            font-size: 24px;
        }
        .header .subtitle {
            color: #6B7280;
            margin: 5px 0 0 0;
            font-size: 14px;
        }
        .meta-info {
            background: #F3F4F6;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 30px;
        }
        .meta-info table {
            width: 100%;
            border-collapse: collapse;
        }
        .meta-info td {
            padding: 5px 0;
            border: none;
        }
        .meta-info .label {
            font-weight: bold;
            width: 150px;
        }
        .data-table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 30px;
        }
        .data-table th,
        .data-table td {
            border: 1px solid #D1D5DB;
            padding: 12px 8px;
            text-align: left;
        }
        .data-table th {
            background-color: #F9FAFB;
            font-weight: bold;
            color: #374151;
        }
        .data-table tr:nth-child(even) {
            background-color: #F9FAFB;
        }
        .summary {
            background: #EFF6FF;
            border-left: 4px solid #3B82F6;
            padding: 20px;
            margin-bottom: 30px;
        }
        .summary h3 {
            margin: 0 0 15px 0;
            color: #1E40AF;
        }
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
        }
        .summary-item {
            background: white;
            padding: 15px;
            border-radius: 6px;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }
        .summary-item .label {
            font-size: 12px;
            color: #6B7280;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .summary-item .value {
            font-size: 20px;
            font-weight: bold;
            color: #1F2937;
            margin-top: 5px;
        }
        .footer {
            margin-top: 50px;
            padding-top: 20px;
            border-top: 1px solid #E5E7EB;
            text-align: center;
            color: #6B7280;
            font-size: 12px;
        }
        @media print {
            body { margin: 0; }
            .header { page-break-inside: avoid; }
            .data-table { page-break-inside: auto; }
            .data-table tr { page-break-inside: avoid; }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.Report.Name}}</h1>
        <p class="subtitle">{{.Report.ReportType | title}} Report</p>
    </div>

    <div class="meta-info">
        <table>
            <tr>
                <td class="label">Generated:</td>
                <td>{{.GeneratedAt.Format "January 2, 2006 at 3:04 PM"}}</td>
                <td class="label">Report Type:</td>
                <td>{{.Report.ReportType | title}}</td>
            </tr>
            <tr>
                <td class="label">Description:</td>
                <td colspan="3">{{if .Report.Description.Valid}}{{.Report.Description.String}}{{else}}No description provided{{end}}</td>
            </tr>
            <tr>
                <td class="label">Total Records:</td>
                <td>{{len .Data.Rows}}</td>
                <td class="label">Columns:</td>
                <td>{{len .Data.Headers}}</td>
            </tr>
        </table>
    </div>

    {{if .Data.Summary}}
    <div class="summary">
        <h3>Summary Statistics</h3>
        <div class="summary-grid">
            {{range $key, $value := .Data.Summary}}
            <div class="summary-item">
                <div class="label">{{$key | replace "_" " " | title}}</div>
                <div class="value">{{$value}}</div>
            </div>
            {{end}}
        </div>
    </div>
    {{end}}

    {{if .Data.Rows}}
    <table class="data-table">
        <thead>
            <tr>
                {{range .Data.Headers}}
                <th>{{.}}</th>
                {{end}}
            </tr>
        </thead>
        <tbody>
            {{range .Data.Rows}}
            <tr>
                {{range $.Data.Headers}}
                <td>{{index $ .}}</td>
                {{end}}
            </tr>
            {{end}}
        </tbody>
    </table>
    {{else}}
    <div style="text-align: center; padding: 50px; color: #6B7280;">
        <p>No data available for this report</p>
    </div>
    {{end}}

    <div class="footer">
        <p>Fire PMAAS - Property Management as a Service</p>
        <p>This report was generated automatically on {{.GeneratedAt.Format "January 2, 2006 at 3:04 PM"}}</p>
    </div>
</body>
</html>`

	// Create template with helper functions
	tmpl := template.New("pdf-report")
	tmpl = tmpl.Funcs(template.FuncMap{
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.Title(s)
		},
		"replace": func(old, new, s string) string {
			return strings.ReplaceAll(s, old, new)
		},
	})

	template.Must(tmpl.Parse(htmlTemplate))
	return tmpl
}

// sanitizeFilename removes unsafe characters from filenames
func sanitizeFilename(filename string) string {
	// Replace unsafe characters with underscores
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	result := filename
	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
