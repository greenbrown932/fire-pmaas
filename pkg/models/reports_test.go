package models

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/greenbrown932/fire-pmaas/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupReportsTestDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	originalDB := db.DB
	db.DB = mockDB

	return mock, func() {
		db.DB = originalDB
		mockDB.Close()
	}
}

func TestCreateCustomReport(t *testing.T) {
	mock, cleanup := setupReportsTestDB(t)
	defer cleanup()

	report := &CustomReport{
		Name:        "Test Property Report",
		Description: NullString("Test description"),
		ReportType:  "property",
		CreatedBy:   1,
		Criteria:    map[string]interface{}{"property_type": "apartment"},
		Columns:     StringArray{"ID", "Name", "Address", "Type"},
		IsPublic:    false,
		IsScheduled: false,
	}

	mock.ExpectQuery(`INSERT INTO custom_reports`).
		WithArgs(report.Name, report.Description, report.ReportType, report.CreatedBy,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			report.IsPublic, report.IsScheduled, report.ScheduleCron).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now()))

	err := CreateCustomReport(report)
	assert.NoError(t, err)
	assert.Equal(t, 1, report.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCustomReports(t *testing.T) {
	mock, cleanup := setupReportsTestDB(t)
	defer cleanup()

	userID := 1
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "report_type", "created_by", "criteria", "columns",
		"chart_config", "is_public", "is_scheduled", "schedule_cron", "last_generated",
		"created_at", "updated_at",
	}).AddRow(1, "Test Report", "Test description", "property", 1,
		`{"property_type": "apartment"}`, `{"ID","Name","Address"}`,
		`{}`, false, false, nil, nil, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM custom_reports`).
		WithArgs(userID).
		WillReturnRows(rows)

	reports, err := GetCustomReports(userID)
	assert.NoError(t, err)
	assert.Len(t, reports, 1)
	assert.Equal(t, "Test Report", reports[0].Name)
	assert.Equal(t, "property", reports[0].ReportType)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteReport(t *testing.T) {
	mock, cleanup := setupReportsTestDB(t)
	defer cleanup()

	reportID := 1
	parameters := map[string]interface{}{"start_date": "2024-01-01"}

	// Mock GetCustomReportByID
	reportRows := sqlmock.NewRows([]string{
		"id", "name", "description", "report_type", "created_by", "criteria", "columns",
		"chart_config", "is_public", "is_scheduled", "schedule_cron", "last_generated",
		"created_at", "updated_at",
	}).AddRow(1, "Property Report", "Test description", "property", 1,
		`{}`, `{"ID","Name","Address"}`,
		`{}`, false, false, nil, nil, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM custom_reports WHERE id = \$1`).
		WithArgs(reportID).
		WillReturnRows(reportRows)

	// Mock the property report query
	dataRows := sqlmock.NewRows([]string{
		"id", "name", "address", "property_type", "unit_count", "occupied_units", "avg_rent", "maintenance_requests",
	}).AddRow(1, "Test Property", "123 Main St", "Apartment", 10, 8, 1500.0, 2).
		AddRow(2, "Another Property", "456 Oak Ave", "Condo", 5, 4, 1200.0, 1)

	mock.ExpectQuery(`SELECT (.+) FROM properties p`).
		WillReturnRows(dataRows)

	// Mock CreateReportExecution
	mock.ExpectQuery(`INSERT INTO report_executions`).
		WithArgs(reportID, sqlmock.AnyArg(), sqlmock.AnyArg(), "completed",
			"json", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	data, err := ExecuteReport(reportID, parameters)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data.Headers, 8) // Expected headers for property report
	assert.Len(t, data.Rows, 2)    // Expected 2 properties

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGeneratePropertyReport(t *testing.T) {
	mock, cleanup := setupReportsTestDB(t)
	defer cleanup()

	report := &CustomReport{
		ReportType: "property",
		Criteria:   map[string]interface{}{},
	}

	dataRows := sqlmock.NewRows([]string{
		"id", "name", "address", "property_type", "unit_count", "occupied_units", "avg_rent", "maintenance_requests",
	}).AddRow(1, "Test Property", "123 Main St", "Apartment", 10, 8, 1500.0, 2)

	mock.ExpectQuery(`SELECT (.+) FROM properties p`).
		WillReturnRows(dataRows)

	data, err := generatePropertyReport(report, map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data.Rows, 1)
	assert.NotNil(t, data.Summary)

	// Check summary calculations
	summary := data.Summary
	assert.Equal(t, 1, summary["total_properties"])
	assert.Equal(t, 10, summary["total_units"])
	assert.Equal(t, 8, summary["total_occupied"])
	assert.Equal(t, 80.0, summary["occupancy_rate"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGenerateFinancialReport(t *testing.T) {
	mock, cleanup := setupReportsTestDB(t)
	defer cleanup()

	report := &CustomReport{
		ReportType: "financial",
		Criteria:   map[string]interface{}{},
	}

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	dataRows := sqlmock.NewRows([]string{
		"month", "payment_count", "total_amount", "avg_amount",
	}).AddRow(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 25, 37500.0, 1500.0).
		AddRow(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), 28, 42000.0, 1500.0)

	mock.ExpectQuery(`SELECT (.+) FROM payments p`).
		WithArgs(startDate, endDate).
		WillReturnRows(dataRows)

	parameters := map[string]interface{}{
		"start_date": "2024-01-01",
		"end_date":   "2024-12-31",
	}

	data, err := generateFinancialReport(report, parameters)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Len(t, data.Rows, 2)
	assert.NotNil(t, data.Summary)

	// Check summary calculations
	summary := data.Summary
	assert.Equal(t, 79500.0, summary["total_amount"])
	assert.Equal(t, 53, summary["total_payments"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateKPIMetric(t *testing.T) {
	mock, cleanup := setupReportsTestDB(t)
	defer cleanup()

	metric := &KPIMetric{
		MetricName:        "Monthly Revenue",
		MetricValue:       15750.0,
		MetricUnit:        NullString("currency"),
		Category:          "financial",
		PeriodStart:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:         time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		CalculationMethod: NullString("Sum of all rent payments"),
	}

	mock.ExpectQuery(`INSERT INTO kpi_metrics`).
		WithArgs(metric.MetricName, metric.MetricValue, metric.MetricUnit,
			metric.Category, metric.PeriodStart, metric.PeriodEnd,
			metric.PropertyID, metric.CalculatedBy, metric.CalculationMethod,
			metric.BenchmarkValue).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(1, time.Now()))

	err := CreateKPIMetric(metric)
	assert.NoError(t, err)
	assert.Equal(t, 1, metric.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetKPIMetrics(t *testing.T) {
	mock, cleanup := setupReportsTestDB(t)
	defer cleanup()

	category := "financial"
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id", "metric_name", "metric_value", "metric_unit", "category",
		"period_start", "period_end", "property_id", "calculated_by",
		"calculation_method", "benchmark_value", "created_at",
	}).AddRow(1, "Monthly Revenue", 15750.0, "currency", "financial",
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		nil, 1, "Sum of payments", nil, time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM kpi_metrics`).
		WithArgs(category, startDate, endDate).
		WillReturnRows(rows)

	metrics, err := GetKPIMetrics(category, startDate, endDate, nil)
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "Monthly Revenue", metrics[0].MetricName)
	assert.Equal(t, 15750.0, metrics[0].MetricValue)
	assert.Equal(t, "financial", metrics[0].Category)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCalculateOccupancyRate(t *testing.T) {
	mock, cleanup := setupReportsTestDB(t)
	defer cleanup()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"total_units", "occupied_units"}).
		AddRow(10, 8)

	mock.ExpectQuery(`SELECT (.+) FROM property_units pu`).
		WithArgs(startDate, endDate).
		WillReturnRows(rows)

	rate, err := CalculateOccupancyRate(startDate, endDate, nil)
	assert.NoError(t, err)
	assert.Equal(t, 80.0, rate)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPerformTrendAnalysis(t *testing.T) {
	mock, cleanup := setupReportsTestDB(t)
	defer cleanup()

	metricName := "Monthly Revenue"
	category := "financial"
	months := 6

	// Mock GetKPIMetrics call
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)

	rows := sqlmock.NewRows([]string{
		"id", "metric_name", "metric_value", "metric_unit", "category",
		"period_start", "period_end", "property_id", "calculated_by",
		"calculation_method", "benchmark_value", "created_at",
	}).AddRow(1, "Monthly Revenue", 15000.0, "currency", "financial",
		startDate, startDate.AddDate(0, 1, 0), nil, 1, "Sum", nil, time.Now()).
		AddRow(2, "Monthly Revenue", 16000.0, "currency", "financial",
			startDate.AddDate(0, 1, 0), startDate.AddDate(0, 2, 0), nil, 1, "Sum", nil, time.Now()).
		AddRow(3, "Monthly Revenue", 17000.0, "currency", "financial",
			startDate.AddDate(0, 2, 0), startDate.AddDate(0, 3, 0), nil, 1, "Sum", nil, time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM kpi_metrics`).
		WithArgs(category, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	analysis, err := PerformTrendAnalysis(metricName, category, months)
	assert.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, metricName, analysis.Metric)
	assert.Equal(t, "increasing", analysis.TrendType) // Based on the increasing values
	assert.Greater(t, analysis.ChangeRate, 0.0)       // Should be positive change
	assert.NotEmpty(t, analysis.Insights)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetReportTemplates(t *testing.T) {
	mock, cleanup := setupReportsTestDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "category", "template_config",
		"is_system", "created_by", "created_at", "updated_at",
	}).AddRow(1, "Monthly Revenue Report", "Revenue summary", "financial",
		`{"data_source": "payments"}`, true, nil, time.Now(), time.Now()).
		AddRow(2, "Property Performance", "Property metrics", "operational",
			`{"data_source": "properties"}`, true, nil, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM report_templates`).
		WillReturnRows(rows)

	templates, err := GetReportTemplates()
	assert.NoError(t, err)
	assert.Len(t, templates, 2)
	assert.Equal(t, "Monthly Revenue Report", templates[0].Name)
	assert.Equal(t, "Property Performance", templates[1].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test helper functions

func TestCalculateChangeRate(t *testing.T) {
	testCases := []struct {
		values   []float64
		expected float64
	}{
		{[]float64{100, 110}, 10.0},
		{[]float64{100, 90}, -10.0},
		{[]float64{100, 100}, 0.0},
		{[]float64{100}, 0.0}, // Single value
		{[]float64{}, 0.0},    // No values
	}

	for _, tc := range testCases {
		result := calculateChangeRate(tc.values)
		assert.InDelta(t, tc.expected, result, 0.01, "Failed for values: %v", tc.values)
	}
}

func TestCalculateCorrelation(t *testing.T) {
	// Test perfect positive correlation
	values := []float64{1, 2, 3, 4, 5}
	correlation := calculateCorrelation(values)
	assert.InDelta(t, 1.0, correlation, 0.01, "Expected perfect positive correlation")

	// Test perfect negative correlation
	values = []float64{5, 4, 3, 2, 1}
	correlation = calculateCorrelation(values)
	assert.InDelta(t, -1.0, correlation, 0.01, "Expected perfect negative correlation")

	// Test no correlation (constant values)
	values = []float64{3, 3, 3, 3, 3}
	correlation = calculateCorrelation(values)
	assert.Equal(t, 0.0, correlation, "Expected zero correlation for constant values")
}

func TestGenerateInsights(t *testing.T) {
	testCases := []struct {
		metricName string
		trendType  string
		changeRate float64
		expected   []string
	}{
		{
			"Monthly Revenue", "increasing", 15.5,
			[]string{"Monthly Revenue has shown an upward trend with 15.5% growth", "This positive revenue trend indicates strong property performance"},
		},
		{
			"Occupancy Rate", "decreasing", -8.2,
			[]string{"Occupancy Rate has declined by 8.2% over the analysis period", "Consider reviewing marketing strategies and rental pricing"},
		},
		{
			"Average Rent", "stable", 1.2,
			[]string{"Average Rent has remained relatively stable with minimal variation"},
		},
	}

	for _, tc := range testCases {
		insights := generateInsights(tc.metricName, tc.trendType, tc.changeRate)
		assert.Equal(t, len(tc.expected), len(insights))
		for i, expected := range tc.expected {
			assert.Equal(t, expected, insights[i])
		}
	}
}

func TestReportsStringArrayValue(t *testing.T) {
	// Test empty array
	arr := StringArray{}
	val, err := arr.Value()
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Test with values
	arr = StringArray{"value1", "value2", "value3"}
	val, err = arr.Value()
	assert.NoError(t, err)
	assert.NotNil(t, val)
}

func TestNullStringHelper(t *testing.T) {
	// Test with empty string
	nullStr := NullString("")
	assert.False(t, nullStr.Valid)

	// Test with non-empty string
	nullStr = NullString("test value")
	assert.True(t, nullStr.Valid)
	assert.Equal(t, "test value", nullStr.String)
}

// Test chart generation functions
func TestGenerateChartsForReport(t *testing.T) {
	reportData := &ReportData{
		Headers: []string{"Month", "Total Amount"},
		Rows: []map[string]interface{}{
			{"Month": "January", "Total Amount": 15000.0},
			{"Month": "February", "Total Amount": 17000.0},
		},
		Summary: map[string]interface{}{
			"total_amount": 32000.0,
		},
	}

	chartConfig := map[string]interface{}{
		"type": "bar",
	}

	charts, err := generateChartsForReport(reportData, chartConfig)
	assert.NoError(t, err)
	assert.Len(t, charts, 1)

	chart := charts[0]
	assert.Equal(t, "bar", chart.Type)
	assert.Equal(t, "Monthly Revenue", chart.Title)
	assert.NotEmpty(t, chart.Data)
}

func TestHasNumericColumn(t *testing.T) {
	rows := []map[string]interface{}{
		{"Text": "hello", "Number": 42, "Float": 3.14},
		{"Text": "world", "Number": 84, "Float": 2.71},
	}

	assert.True(t, hasNumericColumn(rows, "Number"))
	assert.True(t, hasNumericColumn(rows, "Float"))
	assert.False(t, hasNumericColumn(rows, "Text"))
	assert.False(t, hasNumericColumn(rows, "NonExistent"))
}

func TestExtractColumn(t *testing.T) {
	rows := []map[string]interface{}{
		{"Name": "Alice", "Age": 30},
		{"Name": "Bob", "Age": 25},
		{"Name": "Charlie", "Age": 35},
	}

	names := extractColumn(rows, "Name")
	assert.Len(t, names, 3)
	assert.Equal(t, "Alice", names[0])
	assert.Equal(t, "Bob", names[1])
	assert.Equal(t, "Charlie", names[2])

	ages := extractColumn(rows, "Age")
	assert.Len(t, ages, 3)
	assert.Equal(t, 30, ages[0])
	assert.Equal(t, 25, ages[1])
	assert.Equal(t, 35, ages[2])

	// Test non-existent column
	nonExistent := extractColumn(rows, "Height")
	assert.Len(t, nonExistent, 0)
}
