-- Drop all reporting and analytics tables in reverse dependency order

DROP TABLE IF EXISTS scheduled_reports;
DROP TABLE IF EXISTS report_templates;
DROP TABLE IF EXISTS kpi_metrics;
DROP TABLE IF EXISTS saved_charts;
DROP TABLE IF EXISTS dashboard_permissions;
DROP TABLE IF EXISTS analytics_dashboards;
DROP TABLE IF EXISTS report_executions;
DROP TABLE IF EXISTS custom_reports;
