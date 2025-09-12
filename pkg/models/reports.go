package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/greenbrown932/fire-pmaas/pkg/db"
	"github.com/lib/pq"
)

// CustomReport represents a user-defined report configuration
type CustomReport struct {
	ID            int                    `json:"id"`
	Name          string                 `json:"name"`
	Description   sql.NullString         `json:"description,omitempty"`
	ReportType    string                 `json:"report_type"`
	CreatedBy     int                    `json:"created_by"`
	Criteria      map[string]interface{} `json:"criteria"`
	Columns       StringArray            `json:"columns"`
	ChartConfig   map[string]interface{} `json:"chart_config,omitempty"`
	IsPublic      bool                   `json:"is_public"`
	IsScheduled   bool                   `json:"is_scheduled"`
	ScheduleCron  sql.NullString         `json:"schedule_cron,omitempty"`
	LastGenerated sql.NullTime           `json:"last_generated,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ReportExecution represents a report execution instance
type ReportExecution struct {
	ID                  int                    `json:"id"`
	ReportID            int                    `json:"report_id"`
	ExecutedBy          sql.NullInt32          `json:"executed_by,omitempty"`
	ExecutionTime       time.Time              `json:"execution_time"`
	Status              string                 `json:"status"`
	OutputFormat        string                 `json:"output_format"`
	FilePath            sql.NullString         `json:"file_path,omitempty"`
	RowCount            sql.NullInt32          `json:"row_count,omitempty"`
	ExecutionDurationMs sql.NullInt32          `json:"execution_duration_ms,omitempty"`
	ErrorMessage        sql.NullString         `json:"error_message,omitempty"`
	Parameters          map[string]interface{} `json:"parameters,omitempty"`
}

// AnalyticsDashboard represents a custom analytics dashboard
type AnalyticsDashboard struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description sql.NullString         `json:"description,omitempty"`
	CreatedBy   int                    `json:"created_by"`
	Layout      map[string]interface{} `json:"layout"`
	Widgets     []interface{}          `json:"widgets"`
	IsDefault   bool                   `json:"is_default"`
	IsPublic    bool                   `json:"is_public"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// SavedChart represents a saved chart configuration
type SavedChart struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description sql.NullString         `json:"description,omitempty"`
	ChartType   string                 `json:"chart_type"`
	DataSource  string                 `json:"data_source"`
	Config      map[string]interface{} `json:"config"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	CreatedBy   int                    `json:"created_by"`
	IsPublic    bool                   `json:"is_public"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// KPIMetric represents a key performance indicator
type KPIMetric struct {
	ID                int             `json:"id"`
	MetricName        string          `json:"metric_name"`
	MetricValue       float64         `json:"metric_value"`
	MetricUnit        sql.NullString  `json:"metric_unit,omitempty"`
	Category          string          `json:"category"`
	PeriodStart       time.Time       `json:"period_start"`
	PeriodEnd         time.Time       `json:"period_end"`
	PropertyID        sql.NullInt32   `json:"property_id,omitempty"`
	CalculatedBy      sql.NullInt32   `json:"calculated_by,omitempty"`
	CalculationMethod sql.NullString  `json:"calculation_method,omitempty"`
	BenchmarkValue    sql.NullFloat64 `json:"benchmark_value,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
}

// ReportTemplate represents a predefined report template
type ReportTemplate struct {
	ID             int                    `json:"id"`
	Name           string                 `json:"name"`
	Description    sql.NullString         `json:"description,omitempty"`
	Category       string                 `json:"category"`
	TemplateConfig map[string]interface{} `json:"template_config"`
	IsSystem       bool                   `json:"is_system"`
	CreatedBy      sql.NullInt32          `json:"created_by,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// ReportData represents the structure of report output
type ReportData struct {
	Headers []string                 `json:"headers"`
	Rows    []map[string]interface{} `json:"rows"`
	Summary map[string]interface{}   `json:"summary,omitempty"`
	Charts  []ChartData              `json:"charts,omitempty"`
}

// ChartData represents chart configuration and data
type ChartData struct {
	Type   string                 `json:"type"`
	Title  string                 `json:"title"`
	Data   map[string]interface{} `json:"data"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// TrendAnalysis represents trend analysis results
type TrendAnalysis struct {
	Metric      string          `json:"metric"`
	Period      string          `json:"period"`
	TrendType   string          `json:"trend_type"` // "increasing", "decreasing", "stable"
	ChangeRate  float64         `json:"change_rate"`
	Correlation float64         `json:"correlation"`
	Forecast    []ForecastPoint `json:"forecast,omitempty"`
	Insights    []string        `json:"insights"`
}

// ForecastPoint represents a forecasted data point
type ForecastPoint struct {
	Date           time.Time `json:"date"`
	PredictedValue float64   `json:"predicted_value"`
	Confidence     float64   `json:"confidence"`
}

// Report creation and management functions

// CreateCustomReport creates a new custom report
func CreateCustomReport(report *CustomReport) error {
	criteriaJSON, err := json.Marshal(report.Criteria)
	if err != nil {
		return err
	}

	chartConfigJSON, err := json.Marshal(report.ChartConfig)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO custom_reports (name, description, report_type, created_by, criteria, columns,
								  chart_config, is_public, is_scheduled, schedule_cron)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	return db.DB.QueryRow(query, report.Name, report.Description, report.ReportType,
		report.CreatedBy, criteriaJSON, pq.Array(report.Columns), chartConfigJSON,
		report.IsPublic, report.IsScheduled, report.ScheduleCron).
		Scan(&report.ID, &report.CreatedAt, &report.UpdatedAt)
}

// GetCustomReports retrieves custom reports for a user
func GetCustomReports(userID int) ([]CustomReport, error) {
	query := `
		SELECT id, name, description, report_type, created_by, criteria, columns,
			   chart_config, is_public, is_scheduled, schedule_cron, last_generated,
			   created_at, updated_at
		FROM custom_reports
		WHERE created_by = $1 OR is_public = true
		ORDER BY updated_at DESC`

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []CustomReport
	for rows.Next() {
		var report CustomReport
		var criteriaJSON, chartConfigJSON []byte

		err := rows.Scan(&report.ID, &report.Name, &report.Description, &report.ReportType,
			&report.CreatedBy, &criteriaJSON, pq.Array(&report.Columns),
			&chartConfigJSON, &report.IsPublic, &report.IsScheduled, &report.ScheduleCron,
			&report.LastGenerated, &report.CreatedAt, &report.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if err := json.Unmarshal(criteriaJSON, &report.Criteria); err != nil {
			return nil, err
		}

		if len(chartConfigJSON) > 0 {
			if err := json.Unmarshal(chartConfigJSON, &report.ChartConfig); err != nil {
				return nil, err
			}
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// ExecuteReport generates report data based on report configuration
func ExecuteReport(reportID int, parameters map[string]interface{}) (*ReportData, error) {
	// Get report configuration
	report, err := GetCustomReportByID(reportID)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	// Build and execute query based on report type and criteria
	data, err := buildAndExecuteReportQuery(report, parameters)
	if err != nil {
		return nil, err
	}

	// Record execution
	execution := &ReportExecution{
		ReportID:            reportID,
		ExecutionTime:       startTime,
		Status:              "completed",
		OutputFormat:        "json",
		RowCount:            sql.NullInt32{Int32: int32(len(data.Rows)), Valid: true},
		ExecutionDurationMs: sql.NullInt32{Int32: int32(time.Since(startTime).Milliseconds()), Valid: true},
		Parameters:          parameters,
	}

	if err := CreateReportExecution(execution); err != nil {
		// Log error but don't fail the report generation
		fmt.Printf("Failed to record report execution: %v\n", err)
	}

	return data, nil
}

// GetCustomReportByID retrieves a specific custom report
func GetCustomReportByID(id int) (*CustomReport, error) {
	report := &CustomReport{}
	var criteriaJSON, chartConfigJSON []byte

	query := `
		SELECT id, name, description, report_type, created_by, criteria, columns,
			   chart_config, is_public, is_scheduled, schedule_cron, last_generated,
			   created_at, updated_at
		FROM custom_reports WHERE id = $1`

	err := db.DB.QueryRow(query, id).Scan(&report.ID, &report.Name, &report.Description,
		&report.ReportType, &report.CreatedBy, &criteriaJSON, pq.Array(&report.Columns),
		&chartConfigJSON, &report.IsPublic, &report.IsScheduled, &report.ScheduleCron,
		&report.LastGenerated, &report.CreatedAt, &report.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if err := json.Unmarshal(criteriaJSON, &report.Criteria); err != nil {
		return nil, err
	}

	if len(chartConfigJSON) > 0 {
		if err := json.Unmarshal(chartConfigJSON, &report.ChartConfig); err != nil {
			return nil, err
		}
	}

	return report, nil
}

// CreateReportExecution records a report execution
func CreateReportExecution(execution *ReportExecution) error {
	parametersJSON, err := json.Marshal(execution.Parameters)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO report_executions (report_id, executed_by, execution_time, status,
									 output_format, file_path, row_count, execution_duration_ms,
									 error_message, parameters)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	return db.DB.QueryRow(query, execution.ReportID, execution.ExecutedBy,
		execution.ExecutionTime, execution.Status, execution.OutputFormat,
		execution.FilePath, execution.RowCount, execution.ExecutionDurationMs,
		execution.ErrorMessage, parametersJSON).Scan(&execution.ID)
}

// buildAndExecuteReportQuery builds and executes the appropriate query for a report
func buildAndExecuteReportQuery(report *CustomReport, parameters map[string]interface{}) (*ReportData, error) {
	var data *ReportData
	var err error

	switch report.ReportType {
	case "property":
		data, err = generatePropertyReport(report, parameters)
	case "financial":
		data, err = generateFinancialReport(report, parameters)
	case "tenant":
		data, err = generateTenantReport(report, parameters)
	case "maintenance":
		data, err = generateMaintenanceReport(report, parameters)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", report.ReportType)
	}

	if err != nil {
		return nil, err
	}

	// Generate charts if chart config is provided
	if report.ChartConfig != nil && len(report.ChartConfig) > 0 {
		charts, err := generateChartsForReport(data, report.ChartConfig)
		if err == nil {
			data.Charts = charts
		}
	}

	return data, nil
}

// generatePropertyReport generates property-related reports
func generatePropertyReport(report *CustomReport, parameters map[string]interface{}) (*ReportData, error) {
	// Base query for property reports
	query := `
		SELECT p.id, p.name, p.address, p.property_type,
			   COUNT(pu.id) as unit_count,
			   COUNT(CASE WHEN l.status = 'active' THEN 1 END) as occupied_units,
			   COALESCE(AVG(l.monthly_rent), 0) as avg_rent,
			   COUNT(mr.id) as maintenance_requests
		FROM properties p
		LEFT JOIN property_units pu ON p.id = pu.property_id
		LEFT JOIN leases l ON pu.id = l.unit_id
		LEFT JOIN maintenance_requests mr ON p.id = mr.property_id
		WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	// Apply criteria filters
	if propertyIDs, ok := report.Criteria["property_ids"].([]interface{}); ok && len(propertyIDs) > 0 {
		argCount++
		query += fmt.Sprintf(" AND p.id = ANY($%d)", argCount)
		args = append(args, pq.Array(propertyIDs))
	}

	query += " GROUP BY p.id, p.name, p.address, p.property_type ORDER BY p.name"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	headers := []string{"ID", "Name", "Address", "Type", "Units", "Occupied", "Avg Rent", "Maintenance Requests"}
	data := &ReportData{
		Headers: headers,
		Rows:    []map[string]interface{}{},
	}

	for rows.Next() {
		var id, unitCount, occupiedUnits, maintenanceRequests int
		var name, address, propertyType string
		var avgRent float64

		err := rows.Scan(&id, &name, &address, &propertyType, &unitCount,
			&occupiedUnits, &avgRent, &maintenanceRequests)
		if err != nil {
			return nil, err
		}

		row := map[string]interface{}{
			"ID":                   id,
			"Name":                 name,
			"Address":              address,
			"Type":                 propertyType,
			"Units":                unitCount,
			"Occupied":             occupiedUnits,
			"Avg Rent":             avgRent,
			"Maintenance Requests": maintenanceRequests,
		}

		data.Rows = append(data.Rows, row)
	}

	// Calculate summary statistics
	if len(data.Rows) > 0 {
		data.Summary = calculatePropertySummary(data.Rows)
	}

	return data, nil
}

// generateFinancialReport generates financial reports
func generateFinancialReport(report *CustomReport, parameters map[string]interface{}) (*ReportData, error) {
	// Get date range from parameters or use default
	startDate := time.Now().AddDate(0, -1, 0) // Default to last month
	endDate := time.Now()

	if start, ok := parameters["start_date"].(string); ok {
		if parsed, err := time.Parse("2006-01-02", start); err == nil {
			startDate = parsed
		}
	}

	if end, ok := parameters["end_date"].(string); ok {
		if parsed, err := time.Parse("2006-01-02", end); err == nil {
			endDate = parsed
		}
	}

	query := `
		SELECT
			DATE_TRUNC('month', p.payment_date) as month,
			COUNT(p.id) as payment_count,
			SUM(p.amount) as total_amount,
			AVG(p.amount) as avg_amount
		FROM payments p
		WHERE p.payment_date BETWEEN $1 AND $2
		GROUP BY DATE_TRUNC('month', p.payment_date)
		ORDER BY month`

	rows, err := db.DB.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	headers := []string{"Month", "Payment Count", "Total Amount", "Average Amount"}
	data := &ReportData{
		Headers: headers,
		Rows:    []map[string]interface{}{},
	}

	for rows.Next() {
		var month time.Time
		var paymentCount int
		var totalAmount, avgAmount float64

		err := rows.Scan(&month, &paymentCount, &totalAmount, &avgAmount)
		if err != nil {
			return nil, err
		}

		row := map[string]interface{}{
			"Month":          month.Format("2006-01"),
			"Payment Count":  paymentCount,
			"Total Amount":   totalAmount,
			"Average Amount": avgAmount,
		}

		data.Rows = append(data.Rows, row)
	}

	// Calculate financial summary
	if len(data.Rows) > 0 {
		data.Summary = calculateFinancialSummary(data.Rows)
	}

	return data, nil
}

// generateTenantReport generates tenant-related reports
func generateTenantReport(report *CustomReport, parameters map[string]interface{}) (*ReportData, error) {
	query := `
		SELECT t.id, t.first_name, t.last_name, t.email, t.phone_number, t.status,
			   p.name as property_name, l.monthly_rent, l.start_date, l.end_date
		FROM tenants t
		LEFT JOIN leases l ON t.id = l.tenant_id AND l.status = 'active'
		LEFT JOIN property_units pu ON l.unit_id = pu.id
		LEFT JOIN properties p ON pu.property_id = p.id
		ORDER BY t.last_name, t.first_name`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	headers := []string{"ID", "First Name", "Last Name", "Email", "Phone", "Status", "Property", "Rent", "Start Date", "End Date"}
	data := &ReportData{
		Headers: headers,
		Rows:    []map[string]interface{}{},
	}

	for rows.Next() {
		var id int
		var firstName, lastName, email, status string
		var phoneNumber, propertyName sql.NullString
		var monthlyRent sql.NullFloat64
		var startDate, endDate sql.NullTime

		err := rows.Scan(&id, &firstName, &lastName, &email, &phoneNumber, &status,
			&propertyName, &monthlyRent, &startDate, &endDate)
		if err != nil {
			return nil, err
		}

		row := map[string]interface{}{
			"ID":         id,
			"First Name": firstName,
			"Last Name":  lastName,
			"Email":      email,
			"Status":     status,
		}

		if phoneNumber.Valid {
			row["Phone"] = phoneNumber.String
		}
		if propertyName.Valid {
			row["Property"] = propertyName.String
		}
		if monthlyRent.Valid {
			row["Rent"] = monthlyRent.Float64
		}
		if startDate.Valid {
			row["Start Date"] = startDate.Time.Format("2006-01-02")
		}
		if endDate.Valid {
			row["End Date"] = endDate.Time.Format("2006-01-02")
		}

		data.Rows = append(data.Rows, row)
	}

	return data, nil
}

// generateMaintenanceReport generates maintenance-related reports
func generateMaintenanceReport(report *CustomReport, parameters map[string]interface{}) (*ReportData, error) {
	query := `
		SELECT mr.id, p.name as property_name, mr.description, mr.status, mr.priority,
			   mr.reported_date, mr.completed_date,
			   CASE
				   WHEN mr.completed_date IS NOT NULL THEN
					   EXTRACT(days FROM mr.completed_date - mr.reported_date)
				   ELSE NULL
			   END as resolution_days
		FROM maintenance_requests mr
		JOIN properties p ON mr.property_id = p.id
		ORDER BY mr.reported_date DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	headers := []string{"ID", "Property", "Description", "Status", "Priority", "Reported Date", "Completed Date", "Resolution Days"}
	data := &ReportData{
		Headers: headers,
		Rows:    []map[string]interface{}{},
	}

	for rows.Next() {
		var id int
		var propertyName, description, status, priority string
		var reportedDate time.Time
		var completedDate sql.NullTime
		var resolutionDays sql.NullFloat64

		err := rows.Scan(&id, &propertyName, &description, &status, &priority,
			&reportedDate, &completedDate, &resolutionDays)
		if err != nil {
			return nil, err
		}

		row := map[string]interface{}{
			"ID":            id,
			"Property":      propertyName,
			"Description":   description,
			"Status":        status,
			"Priority":      priority,
			"Reported Date": reportedDate.Format("2006-01-02"),
		}

		if completedDate.Valid {
			row["Completed Date"] = completedDate.Time.Format("2006-01-02")
		}
		if resolutionDays.Valid {
			row["Resolution Days"] = int(resolutionDays.Float64)
		}

		data.Rows = append(data.Rows, row)
	}

	return data, nil
}

// Helper functions for calculations and chart generation

// calculatePropertySummary calculates summary statistics for property reports
func calculatePropertySummary(rows []map[string]interface{}) map[string]interface{} {
	totalProperties := len(rows)
	totalUnits := 0
	totalOccupied := 0
	totalRent := 0.0

	for _, row := range rows {
		if units, ok := row["Units"].(int); ok {
			totalUnits += units
		}
		if occupied, ok := row["Occupied"].(int); ok {
			totalOccupied += occupied
		}
		if rent, ok := row["Avg Rent"].(float64); ok {
			totalRent += rent
		}
	}

	occupancyRate := 0.0
	if totalUnits > 0 {
		occupancyRate = float64(totalOccupied) / float64(totalUnits) * 100
	}

	avgRent := 0.0
	if totalProperties > 0 {
		avgRent = totalRent / float64(totalProperties)
	}

	return map[string]interface{}{
		"total_properties": totalProperties,
		"total_units":      totalUnits,
		"total_occupied":   totalOccupied,
		"occupancy_rate":   math.Round(occupancyRate*100) / 100,
		"average_rent":     math.Round(avgRent*100) / 100,
	}
}

// calculateFinancialSummary calculates summary statistics for financial reports
func calculateFinancialSummary(rows []map[string]interface{}) map[string]interface{} {
	totalAmount := 0.0
	totalPayments := 0

	for _, row := range rows {
		if amount, ok := row["Total Amount"].(float64); ok {
			totalAmount += amount
		}
		if count, ok := row["Payment Count"].(int); ok {
			totalPayments += count
		}
	}

	avgPaymentAmount := 0.0
	if totalPayments > 0 {
		avgPaymentAmount = totalAmount / float64(totalPayments)
	}

	return map[string]interface{}{
		"total_amount":           math.Round(totalAmount*100) / 100,
		"total_payments":         totalPayments,
		"average_payment_amount": math.Round(avgPaymentAmount*100) / 100,
		"reporting_period":       fmt.Sprintf("%d months", len(rows)),
	}
}

// generateChartsForReport generates chart configurations based on report data
func generateChartsForReport(data *ReportData, chartConfig map[string]interface{}) ([]ChartData, error) {
	var charts []ChartData

	// Example: Generate a bar chart for property data
	if len(data.Rows) > 0 {
		// Check if we have numeric data to chart
		if hasNumericColumn(data.Rows, "Total Amount") {
			chart := ChartData{
				Type:  "bar",
				Title: "Monthly Revenue",
				Data: map[string]interface{}{
					"labels": extractColumn(data.Rows, "Month"),
					"datasets": []map[string]interface{}{
						{
							"label":           "Revenue",
							"data":            extractColumn(data.Rows, "Total Amount"),
							"backgroundColor": "rgba(54, 162, 235, 0.2)",
							"borderColor":     "rgba(54, 162, 235, 1)",
							"borderWidth":     1,
						},
					},
				},
			}
			charts = append(charts, chart)
		}
	}

	return charts, nil
}

// hasNumericColumn checks if a column contains numeric data
func hasNumericColumn(rows []map[string]interface{}, columnName string) bool {
	for _, row := range rows {
		if val, exists := row[columnName]; exists {
			switch val.(type) {
			case int, int32, int64, float32, float64:
				return true
			}
		}
	}
	return false
}

// extractColumn extracts values from a specific column
func extractColumn(rows []map[string]interface{}, columnName string) []interface{} {
	var values []interface{}
	for _, row := range rows {
		if val, exists := row[columnName]; exists {
			values = append(values, val)
		}
	}
	return values
}

// KPI and Analytics functions

// GetKPIMetrics retrieves KPI metrics for a specific period and category
func GetKPIMetrics(category string, startDate, endDate time.Time, propertyID *int) ([]KPIMetric, error) {
	query := `
		SELECT id, metric_name, metric_value, metric_unit, category, period_start, period_end,
			   property_id, calculated_by, calculation_method, benchmark_value, created_at
		FROM kpi_metrics
		WHERE category = $1 AND period_start >= $2 AND period_end <= $3`

	args := []interface{}{category, startDate, endDate}

	if propertyID != nil {
		query += " AND property_id = $4"
		args = append(args, *propertyID)
	}

	query += " ORDER BY period_start DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []KPIMetric
	for rows.Next() {
		var metric KPIMetric
		err := rows.Scan(&metric.ID, &metric.MetricName, &metric.MetricValue,
			&metric.MetricUnit, &metric.Category, &metric.PeriodStart,
			&metric.PeriodEnd, &metric.PropertyID, &metric.CalculatedBy,
			&metric.CalculationMethod, &metric.BenchmarkValue, &metric.CreatedAt)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// CreateKPIMetric creates a new KPI metric
func CreateKPIMetric(metric *KPIMetric) error {
	query := `
		INSERT INTO kpi_metrics (metric_name, metric_value, metric_unit, category,
							   period_start, period_end, property_id, calculated_by,
							   calculation_method, benchmark_value)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	return db.DB.QueryRow(query, metric.MetricName, metric.MetricValue,
		metric.MetricUnit, metric.Category, metric.PeriodStart, metric.PeriodEnd,
		metric.PropertyID, metric.CalculatedBy, metric.CalculationMethod,
		metric.BenchmarkValue).Scan(&metric.ID, &metric.CreatedAt)
}

// CalculateOccupancyRate calculates the occupancy rate for a given period
func CalculateOccupancyRate(startDate, endDate time.Time, propertyID *int) (float64, error) {
	query := `
		SELECT
			COUNT(DISTINCT pu.id) as total_units,
			COUNT(DISTINCT CASE WHEN l.status = 'active' AND
				  l.start_date <= $2 AND l.end_date >= $1 THEN pu.id END) as occupied_units
		FROM property_units pu
		LEFT JOIN leases l ON pu.id = l.unit_id`

	args := []interface{}{startDate, endDate}

	if propertyID != nil {
		query += " WHERE pu.property_id = $3"
		args = append(args, *propertyID)
	}

	var totalUnits, occupiedUnits int
	err := db.DB.QueryRow(query, args...).Scan(&totalUnits, &occupiedUnits)
	if err != nil {
		return 0, err
	}

	if totalUnits == 0 {
		return 0, nil
	}

	return float64(occupiedUnits) / float64(totalUnits) * 100, nil
}

// PerformTrendAnalysis analyzes trends in KPI metrics
func PerformTrendAnalysis(metricName string, category string, months int) (*TrendAnalysis, error) {
	// Get historical data for the metric
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)

	metrics, err := GetKPIMetrics(category, startDate, endDate, nil)
	if err != nil {
		return nil, err
	}

	// Filter for specific metric
	var values []float64
	var dates []time.Time
	for _, metric := range metrics {
		if metric.MetricName == metricName {
			values = append(values, metric.MetricValue)
			dates = append(dates, metric.PeriodStart)
		}
	}

	if len(values) < 2 {
		return nil, fmt.Errorf("insufficient data for trend analysis")
	}

	// Calculate trend
	changeRate := calculateChangeRate(values)
	trendType := "stable"
	if changeRate > 5 {
		trendType = "increasing"
	} else if changeRate < -5 {
		trendType = "decreasing"
	}

	analysis := &TrendAnalysis{
		Metric:      metricName,
		Period:      fmt.Sprintf("%d months", months),
		TrendType:   trendType,
		ChangeRate:  changeRate,
		Correlation: calculateCorrelation(values),
		Insights:    generateInsights(metricName, trendType, changeRate),
	}

	return analysis, nil
}

// Helper functions for trend analysis

// calculateChangeRate calculates the percentage change rate
func calculateChangeRate(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	first := values[0]
	last := values[len(values)-1]

	if first == 0 {
		return 0
	}

	return ((last - first) / first) * 100
}

// calculateCorrelation calculates correlation coefficient
func calculateCorrelation(values []float64) float64 {
	n := len(values)
	if n < 2 {
		return 0
	}

	// Simple correlation with time (assuming linear time series)
	sumX, sumY, sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0, 0.0

	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	numerator := float64(n)*sumXY - sumX*sumY
	denominator := math.Sqrt((float64(n)*sumX2 - sumX*sumX) * (float64(n)*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// generateInsights generates textual insights based on trend analysis
func generateInsights(metricName, trendType string, changeRate float64) []string {
	var insights []string

	switch trendType {
	case "increasing":
		insights = append(insights, fmt.Sprintf("%s has shown an upward trend with %.1f%% growth", metricName, changeRate))
		if strings.Contains(strings.ToLower(metricName), "revenue") {
			insights = append(insights, "This positive revenue trend indicates strong property performance")
		}
	case "decreasing":
		insights = append(insights, fmt.Sprintf("%s has declined by %.1f%% over the analysis period", metricName, math.Abs(changeRate)))
		if strings.Contains(strings.ToLower(metricName), "occupancy") {
			insights = append(insights, "Consider reviewing marketing strategies and rental pricing")
		}
	case "stable":
		insights = append(insights, fmt.Sprintf("%s has remained relatively stable with minimal variation", metricName))
	}

	return insights
}

// GetReportTemplates retrieves available report templates
func GetReportTemplates() ([]ReportTemplate, error) {
	query := `
		SELECT id, name, description, category, template_config, is_system,
			   created_by, created_at, updated_at
		FROM report_templates
		ORDER BY is_system DESC, category, name`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []ReportTemplate
	for rows.Next() {
		var template ReportTemplate
		var configJSON []byte

		err := rows.Scan(&template.ID, &template.Name, &template.Description,
			&template.Category, &configJSON, &template.IsSystem,
			&template.CreatedBy, &template.CreatedAt, &template.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configJSON, &template.TemplateConfig); err != nil {
			return nil, err
		}

		templates = append(templates, template)
	}

	return templates, nil
}
