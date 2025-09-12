-- Reports and Analytics Tables for Epic 3

-- Custom reports table for user-created reports
CREATE TABLE custom_reports (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    report_type VARCHAR(50) NOT NULL, -- 'property', 'financial', 'tenant', 'maintenance', 'custom'
    created_by INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    criteria JSONB NOT NULL, -- Flexible JSON criteria for filtering
    columns TEXT[] NOT NULL, -- Array of column names to include
    chart_config JSONB, -- Chart configuration if visualization is needed
    is_public BOOLEAN DEFAULT FALSE, -- Whether other users can see this report
    is_scheduled BOOLEAN DEFAULT FALSE, -- Whether this report runs on schedule
    schedule_cron VARCHAR(100), -- Cron expression for scheduled reports
    last_generated TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Report executions table to track when reports were generated
CREATE TABLE report_executions (
    id SERIAL PRIMARY KEY,
    report_id INT NOT NULL REFERENCES custom_reports(id) ON DELETE CASCADE,
    executed_by INT REFERENCES users(id) ON DELETE SET NULL,
    execution_time TIMESTAMPTZ DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'completed', -- 'running', 'completed', 'failed'
    output_format VARCHAR(20) NOT NULL DEFAULT 'json', -- 'json', 'csv', 'pdf'
    file_path TEXT, -- Path to generated file if applicable
    row_count INT,
    execution_duration_ms INT,
    error_message TEXT,
    parameters JSONB -- Runtime parameters used for this execution
);

-- Analytics dashboards table for custom dashboard configurations
CREATE TABLE analytics_dashboards (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    layout JSONB NOT NULL, -- Dashboard layout configuration
    widgets JSONB NOT NULL, -- Array of widget configurations
    is_default BOOLEAN DEFAULT FALSE, -- Whether this is the default dashboard for the user
    is_public BOOLEAN DEFAULT FALSE, -- Whether other users can see this dashboard
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Dashboard sharing permissions
CREATE TABLE dashboard_permissions (
    id SERIAL PRIMARY KEY,
    dashboard_id INT NOT NULL REFERENCES analytics_dashboards(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    role_name VARCHAR(50), -- Allow access by role instead of specific user
    permission_level VARCHAR(20) NOT NULL DEFAULT 'view', -- 'view', 'edit', 'admin'
    granted_by INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_dashboard_user UNIQUE(dashboard_id, user_id),
    CONSTRAINT unique_dashboard_role UNIQUE(dashboard_id, role_name)
);

-- Saved chart configurations
CREATE TABLE saved_charts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    chart_type VARCHAR(50) NOT NULL, -- 'line', 'bar', 'pie', 'doughnut', 'scatter', 'area'
    data_source VARCHAR(100) NOT NULL, -- Source of data: 'properties', 'payments', 'maintenance', etc.
    config JSONB NOT NULL, -- Full Chart.js configuration
    filters JSONB, -- Data filters applied to the chart
    created_by INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- KPI tracking table for key performance indicators
CREATE TABLE kpi_metrics (
    id SERIAL PRIMARY KEY,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,2) NOT NULL,
    metric_unit VARCHAR(20), -- 'currency', 'percentage', 'count', 'days', etc.
    category VARCHAR(50) NOT NULL, -- 'financial', 'operational', 'tenant_satisfaction', etc.
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    property_id INT REFERENCES properties(id) ON DELETE CASCADE, -- NULL for global metrics
    calculated_by INT REFERENCES users(id) ON DELETE SET NULL,
    calculation_method TEXT, -- Description of how this metric was calculated
    benchmark_value DECIMAL(15,2), -- Target or benchmark value
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Scheduled report jobs
CREATE TABLE scheduled_reports (
    id SERIAL PRIMARY KEY,
    report_id INT NOT NULL REFERENCES custom_reports(id) ON DELETE CASCADE,
    cron_expression VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_run TIMESTAMPTZ,
    next_run TIMESTAMPTZ,
    output_format VARCHAR(20) DEFAULT 'pdf',
    email_recipients TEXT[], -- Array of email addresses to send reports to
    created_by INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Report templates for common report types
CREATE TABLE report_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL, -- 'financial', 'operational', 'compliance', etc.
    template_config JSONB NOT NULL, -- Template configuration
    is_system BOOLEAN DEFAULT FALSE, -- System templates vs user templates
    created_by INT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert default report templates
INSERT INTO report_templates (name, description, category, template_config, is_system) VALUES
('Monthly Revenue Report', 'Summary of rental income and expenses by month', 'financial',
 '{"data_source": "payments", "group_by": "month", "metrics": ["total_income", "total_expenses", "net_income"], "charts": ["line", "bar"]}', true),

('Property Performance Report', 'Occupancy rates and performance metrics by property', 'operational',
 '{"data_source": "properties", "group_by": "property", "metrics": ["occupancy_rate", "avg_rent", "maintenance_cost"], "charts": ["bar", "pie"]}', true),

('Tenant Report', 'Tenant demographics and lease information', 'operational',
 '{"data_source": "tenants", "group_by": "property", "metrics": ["tenant_count", "avg_lease_duration", "turnover_rate"], "charts": ["bar", "pie"]}', true),

('Maintenance Summary', 'Maintenance requests and resolution times', 'operational',
 '{"data_source": "maintenance_requests", "group_by": "month", "metrics": ["request_count", "avg_resolution_time", "cost_per_request"], "charts": ["line", "bar"]}', true),

('Financial Dashboard', 'Comprehensive financial overview', 'financial',
 '{"data_source": "multi", "metrics": ["total_revenue", "total_expenses", "profit_margin", "rent_collection_rate"], "charts": ["line", "doughnut", "bar"]}', true);

-- Create indexes for better query performance
CREATE INDEX idx_custom_reports_created_by ON custom_reports(created_by);
CREATE INDEX idx_custom_reports_type ON custom_reports(report_type);
CREATE INDEX idx_report_executions_report_id ON report_executions(report_id);
CREATE INDEX idx_report_executions_executed_by ON report_executions(executed_by);
CREATE INDEX idx_report_executions_time ON report_executions(execution_time);
CREATE INDEX idx_analytics_dashboards_created_by ON analytics_dashboards(created_by);
CREATE INDEX idx_dashboard_permissions_dashboard_id ON dashboard_permissions(dashboard_id);
CREATE INDEX idx_saved_charts_created_by ON saved_charts(created_by);
CREATE INDEX idx_kpi_metrics_category ON kpi_metrics(category);
CREATE INDEX idx_kpi_metrics_property_id ON kpi_metrics(property_id);
CREATE INDEX idx_kpi_metrics_period ON kpi_metrics(period_start, period_end);
CREATE INDEX idx_scheduled_reports_next_run ON scheduled_reports(next_run);
CREATE INDEX idx_report_templates_category ON report_templates(category);

-- Add some sample KPI data for demonstration
INSERT INTO kpi_metrics (metric_name, metric_value, metric_unit, category, period_start, period_end, calculation_method) VALUES
('Monthly Revenue', 15750.00, 'currency', 'financial', '2024-01-01', '2024-01-31', 'Sum of all rent payments received'),
('Occupancy Rate', 85.50, 'percentage', 'operational', '2024-01-01', '2024-01-31', 'Occupied units / Total units * 100'),
('Average Rent', 1575.00, 'currency', 'financial', '2024-01-01', '2024-01-31', 'Total rent / Number of occupied units'),
('Maintenance Response Time', 2.5, 'days', 'operational', '2024-01-01', '2024-01-31', 'Average time from request to completion'),
('Tenant Satisfaction', 4.2, 'rating', 'tenant_satisfaction', '2024-01-01', '2024-01-31', 'Average of tenant satisfaction surveys');
