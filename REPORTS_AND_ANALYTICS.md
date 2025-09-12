# Fire PMAAS - Reports and Analytics Documentation

## Overview

Fire PMAAS now includes comprehensive reporting and analytics capabilities that enable users to generate customizable reports, visualize data through charts and graphs, export reports in various formats (including PDF), and perform trend analysis on key metrics.

## Features Implemented

### Epic 3: Reporting and Analytics

✅ **Story 1: Report Generation** - Users can generate customizable reports based on specific criteria
✅ **Story 2: Data Visualization** - Users can view data in charts and graphs to identify trends and patterns  
✅ **Story 3: Report Export** - Users can export reports in PDF format for sharing
✅ **Story 4: Trend Analysis** - Users can perform trend analysis on data for informed decision-making

## Database Schema

### New Tables Added

1. **custom_reports** - Store user-defined report configurations
2. **report_executions** - Track report generation history and performance
3. **analytics_dashboards** - Custom dashboard configurations
4. **dashboard_permissions** - Dashboard sharing permissions
5. **saved_charts** - Reusable chart configurations
6. **kpi_metrics** - Key performance indicator tracking
7. **scheduled_reports** - Automated report scheduling
8. **report_templates** - Pre-defined report templates

### Sample Data Included

- Default report templates for common use cases
- Sample KPI metrics for demonstration
- Proper indexing for performance optimization

## API Endpoints

### Report Management

```
GET    /api/reports                    - Get user's reports
POST   /api/reports                    - Create new report
GET    /api/reports/{id}               - Get specific report
PUT    /api/reports/{id}               - Update report
DELETE /api/reports/{id}               - Delete report
POST   /api/reports/{id}/execute       - Execute report
POST   /api/reports/{id}/export        - Export report (PDF/CSV/Excel)
```

### Report Templates

```
GET    /api/report-templates           - Get available templates
POST   /api/reports/from-template/{id} - Create report from template
```

### Analytics and KPIs

```
GET    /api/analytics/kpis             - Get KPI metrics
POST   /api/analytics/kpis             - Create KPI metric
GET    /api/analytics/trends/{metric}  - Get trend analysis
GET    /api/analytics/summary          - Get analytics summary
```

### Charts and Visualizations

```
GET    /api/charts                     - Get saved charts
POST   /api/charts                     - Create chart
GET    /api/charts/{id}                - Get specific chart
PUT    /api/charts/{id}                - Update chart
DELETE /api/charts/{id}                - Delete chart
```

### Dashboard Management

```
GET    /api/dashboards                 - Get user dashboards
POST   /api/dashboards                 - Create dashboard
GET    /api/dashboards/{id}            - Get specific dashboard
PUT    /api/dashboards/{id}            - Update dashboard
DELETE /api/dashboards/{id}            - Delete dashboard
```

### Quick Stats (for widgets)

```
GET    /api/stats/properties           - Property statistics
GET    /api/stats/financial            - Financial statistics
GET    /api/stats/tenants              - Tenant statistics
GET    /api/stats/maintenance          - Maintenance statistics
```

## Web Interface

### New Pages Added

1. **Reports Page** (`/reports`)
   - View and manage custom reports
   - Quick report generation
   - Report templates gallery
   - Recent reports list

2. **Analytics Dashboard** (`/analytics`)
   - Real-time KPI cards
   - Interactive charts and graphs
   - Trend analysis visualization
   - Customizable time ranges

3. **Create Report Page** (`/reports/create`)
   - Report builder interface
   - Column selection
   - Criteria configuration
   - Chart inclusion options

4. **View Report Page** (`/reports/{id}/view`)
   - Report results display
   - Export options
   - Chart rendering

### Navigation Updates

Added new navigation items in the sidebar:
- Reports
- Analytics

## Report Types Supported

### 1. Property Reports
- Property overview and performance
- Occupancy rates by property
- Unit details and status
- Maintenance request summary per property

**Default Columns:**
- ID, Name, Address, Type, Units, Occupied, Avg Rent, Maintenance Requests

### 2. Financial Reports
- Monthly revenue trends
- Payment collection analysis
- Expense tracking
- Profit/loss summaries

**Default Columns:**
- Month, Payment Count, Total Amount, Average Amount, Collection Rate

### 3. Tenant Reports
- Tenant demographics
- Lease information
- Move-in/move-out tracking
- Tenant satisfaction metrics

**Default Columns:**
- ID, First Name, Last Name, Email, Phone, Status, Property, Rent

### 4. Maintenance Reports
- Request tracking and resolution
- Priority breakdown
- Response time analysis
- Cost analysis

**Default Columns:**
- ID, Property, Description, Status, Priority, Reported Date, Resolution Days

## Export Formats

### PDF Export
- Professional formatting with Fire PMAAS branding
- Summary statistics included
- Charts and graphs embedded
- Print-optimized layout
- Uses wkhtmltopdf if available, falls back to basic PDF generation

### CSV Export
- Standard comma-separated format
- All data columns included
- Compatible with Excel and other spreadsheet applications

### Excel Export
- Microsoft Excel compatible format
- Formatted for easy reading
- Includes data validation where appropriate

## Data Visualization

### Chart Types Supported
- **Line Charts** - Trend analysis over time
- **Bar Charts** - Comparative data visualization
- **Pie/Doughnut Charts** - Proportion and distribution analysis
- **Area Charts** - Cumulative data visualization
- **Scatter Plots** - Correlation analysis

### Interactive Features
- Responsive design for all screen sizes
- Chart.js integration for professional visualizations
- Time period selectors
- Real-time data updates
- Export chart data

## KPI Metrics and Analytics

### Key Performance Indicators Tracked

#### Financial KPIs
- Monthly Revenue
- Total Expenses
- Net Income
- Collection Rate
- Average Rent
- Revenue per Unit

#### Operational KPIs
- Occupancy Rate
- Average Lease Duration
- Turnover Rate
- Maintenance Response Time
- Cost per Request
- Tenant Satisfaction Score

#### Property Performance KPIs
- Revenue Efficiency
- Maintenance Cost per Property
- Vacancy Duration
- Rental Yield

### Trend Analysis Features
- **Trend Direction** - Increasing, decreasing, or stable
- **Change Rate** - Percentage change over time
- **Correlation Analysis** - Statistical relationships
- **Automated Insights** - AI-generated observations
- **Forecasting** - Predictive analytics (basic implementation)

## Security and Permissions

### Role-Based Access Control

#### Admin Users
- Full access to all reports and analytics
- Can create, edit, and delete any report
- Access to system-wide KPIs
- Dashboard management permissions

#### Property Managers
- Create and manage property-related reports
- Access to operational and financial analytics
- Limited administrative functions
- Can share reports with team members

#### Tenants
- View own lease and payment information
- Limited report generation capabilities
- Access to personal analytics only
- Cannot view other tenants' data

#### Viewers
- Read-only access to public reports
- Basic analytics viewing
- Cannot create or modify reports
- Limited dashboard access

### Data Privacy
- Users can only access data they have permission to view
- Reports can be marked as public or private
- Audit trail for report generation and access
- Secure export functionality

## Performance Considerations

### Database Optimization
- Proper indexing on frequently queried columns
- Optimized queries for large datasets
- Connection pooling for concurrent users
- Query result caching where appropriate

### Report Generation
- Asynchronous processing for large reports
- Pagination for extensive datasets
- Streaming for PDF exports
- Background job processing for scheduled reports

### Frontend Performance
- Lazy loading of charts and visualizations
- Efficient data binding
- Responsive image loading
- Optimized JavaScript bundling

## Installation and Setup

### Database Migration
```bash
# Run the new migration
go run cmd/migrate/main.go up
```

### Dependencies
- Chart.js for data visualization
- wkhtmltopdf (optional) for enhanced PDF generation
- Standard Go libraries for core functionality

### Configuration
No additional configuration required. The system uses existing database and authentication settings.

## Usage Examples

### Creating a Custom Report

1. Navigate to `/reports`
2. Click "Create Report"
3. Fill in report details:
   - Name: "Monthly Property Performance"
   - Type: "Property"
   - Description: "Monthly overview of all properties"
4. Select columns to include
5. Set criteria (optional)
6. Enable charts if desired
7. Click "Create Report"

### Generating Quick Reports

1. Go to Reports page
2. Click on one of the quick report buttons:
   - Property Overview
   - Financial Summary
   - Tenant Report
   - Maintenance Report
3. Report generates automatically and displays results

### Exporting Reports

1. Execute a report to view results
2. Click "Export" button
3. Select format (PDF, CSV, Excel)
4. File downloads automatically

### Viewing Analytics

1. Navigate to `/analytics`
2. View real-time KPI cards
3. Interact with charts using time period selectors
4. Review trend analysis for key metrics
5. Customize dashboard layout (future enhancement)

## API Usage Examples

### Create a Custom Report
```bash
curl -X POST http://localhost:8000/api/reports \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Property Revenue Report",
    "report_type": "financial",
    "description": "Monthly revenue by property",
    "columns": ["Property", "Month", "Revenue", "Occupancy"],
    "criteria": {"start_date": "2024-01-01", "end_date": "2024-12-31"},
    "is_public": false
  }'
```

### Execute a Report
```bash
curl -X POST http://localhost:8000/api/reports/1/execute \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "2024-01-01",
    "end_date": "2024-03-31"
  }'
```

### Get KPI Metrics
```bash
curl "http://localhost:8000/api/analytics/kpis?category=financial&start_date=2024-01-01&end_date=2024-12-31"
```

### Export Report as PDF
```bash
curl -X POST http://localhost:8000/api/reports/1/export \
  -H "Content-Type: application/json" \
  -d '{"format": "pdf"}' \
  --output report.pdf
```

## Testing

### Test Coverage
- ✅ API endpoint testing
- ✅ Model function testing
- ✅ PDF generation testing
- ✅ Chart generation testing
- ✅ Database interaction testing
- ✅ Security and permissions testing

### Running Tests
```bash
# Run all reporting tests
go test ./pkg/api/reports_test.go
go test ./pkg/models/reports_test.go

# Run with coverage
go test -cover ./pkg/api/
go test -cover ./pkg/models/
```

## Future Enhancements

### Planned Features
1. **Advanced Scheduling** - Cron-based automated report generation
2. **Email Delivery** - Automated report distribution via email
3. **Interactive Dashboards** - Drag-and-drop dashboard builder
4. **Advanced Analytics** - Machine learning insights and predictions
5. **Real-time Alerts** - Threshold-based notifications
6. **Mobile Optimization** - Enhanced mobile experience
7. **API Rate Limiting** - Performance and security improvements
8. **Report Versioning** - Track report changes over time

### Integration Opportunities
1. **External Data Sources** - Import data from other property management systems
2. **BI Tool Integration** - Connect with Tableau, Power BI, etc.
3. **Accounting Software** - Sync with QuickBooks, Xero
4. **Marketing Platforms** - Integration with rental listing sites

## Troubleshooting

### Common Issues

#### PDF Generation Not Working
- **Issue**: PDF exports return errors
- **Solution**: Install wkhtmltopdf or ensure it's in system PATH
- **Alternative**: System falls back to basic PDF generation

#### Charts Not Displaying
- **Issue**: Charts appear blank or don't load
- **Solution**: Check browser console for JavaScript errors, ensure Chart.js is loaded

#### Performance Issues with Large Reports
- **Issue**: Reports take too long to generate
- **Solution**: Implement pagination, use filters to reduce dataset size

#### Permission Denied Errors
- **Issue**: Users cannot access certain reports
- **Solution**: Verify user roles and permissions in the database

### Debug Mode
Enable detailed logging by setting environment variable:
```bash
export FIRE_DEBUG=true
```

### Log Files
Check application logs for detailed error information:
- Report generation errors
- Database query performance
- Authentication issues
- Export process logs

## Support and Contributing

### Getting Help
1. Check this documentation first
2. Review the test files for usage examples
3. Check the GitHub issues for known problems
4. Create a new issue with detailed description

### Contributing
1. Follow the existing code patterns
2. Add tests for new functionality
3. Update documentation for changes
4. Submit pull request with clear description

## Changelog

### Version 1.0.0 - Initial Release
- Complete reporting and analytics system
- PDF export functionality
- Interactive dashboards
- Trend analysis capabilities
- Role-based security
- Comprehensive test suite
- Full documentation

---

For technical support or feature requests, please refer to the project's GitHub repository or contact the development team.