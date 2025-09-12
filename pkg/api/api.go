package api

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/greenbrown932/fire-pmaas/pkg/middleware"
	"github.com/greenbrown932/fire-pmaas/pkg/models"
)

func RegisterRoutes(r *chi.Mux) {
	// Serve static files (CSS, JS, images)
	workDir, _ := filepath.Abs("./")
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(filesDir)))

	// Register user API routes (includes public and protected routes)
	RegisterUserRoutes(r)

	// Register report and analytics API routes
	RegisterReportRoutes(r)

	// API Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})
	r.Get("/callback", middleware.HandleCallback)

	// Web routes (protected by authentication)
	r.Group(func(auth chi.Router) {
		auth.Use(middleware.LoadUserFromToken)
		auth.Use(middleware.RequireLogin)

		// Properties routes
		auth.Get("/properties/new", handleNewPropertyForm)
		auth.Post("/properties/new", handleCreateProperty)
		auth.Get("/properties/import", handleImportPropertyForm)
		auth.Post("/properties/import", handleImportProperty)

		// Dashboard and UI routes
		auth.Get("/", handleDashboard)
		auth.Get("/profile", handleProfilePage)

		// Properties routes with role-based access
		auth.Group(func(props chi.Router) {
			props.Use(middleware.RequireAnyRole("admin", "property_manager", "viewer"))
			props.Get("/properties", handleProperties)
			props.Get("/properties/{id}", handlePropertyDetail)
		})

		// Tenant management routes
		auth.Group(func(tenants chi.Router) {
			tenants.Use(middleware.RequireAnyRole("admin", "property_manager", "viewer"))
			tenants.Get("/tenants", handleTenants)
		})

		// Maintenance routes
		auth.Group(func(maint chi.Router) {
			maint.Use(middleware.RequireAnyRole("admin", "property_manager", "tenant"))
			maint.Get("/maintenance", handleMaintenance)
		})

		// Reports and Analytics routes
		auth.Group(func(reports chi.Router) {
			reports.Use(middleware.RequireAnyRole("admin", "property_manager", "viewer"))
			reports.Get("/reports", handleReportsPage)
			reports.Get("/analytics", handleAnalyticsPage)
			reports.Get("/reports/create", handleCreateReportPage)
			reports.Get("/reports/{id}/view", handleViewReportPage)
		})

		// Admin routes
		auth.Group(func(admin chi.Router) {
			admin.Use(middleware.RequireAnyRole("admin", "property_manager"))
			admin.Get("/admin/users", handleAdminUsersPage)
		})
	})
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	properties, err := models.GetProperties()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Title      string
		Properties []models.PropertyDetail
		Stats      struct {
			TotalProperties int
			Occupied        int
			Vacant          int
			Maintenance     int
			MonthlyRevenue  float64
		}
	}{
		Title:      "Property Management Dashboard",
		Properties: properties,
	}

	// Calculate stats
	for _, p := range properties {
		data.Stats.TotalProperties++
		switch p.Status {
		case "active":
			data.Stats.Occupied++
			data.Stats.MonthlyRevenue += p.Rent
		case "ended":
			data.Stats.Vacant++
		case "pending":
			data.Stats.Maintenance++
		}
	}

	renderTemplate(w, "dashboard.html", data)
}

func handleProperties(w http.ResponseWriter, r *http.Request) {
	tags := r.URL.Query()["tags"]

	var properties []models.PropertyDetail
	var err error

	if len(tags) > 0 {
		properties, err = models.GetPropertiesByTags(tags)
	} else {
		properties, err = models.GetProperties()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Title      string
		Properties []models.PropertyDetail
	}{
		Title:      "All Properties",
		Properties: properties,
	}
	renderTemplate(w, "properties.html", data)
}

func handlePropertyDetail(w http.ResponseWriter, r *http.Request) {
	// TODO: In a real app, you'd parse the ID and look up the property
	properties, err := models.GetProperties()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Title    string
		Property models.PropertyDetail
	}{
		Title:    "Property Details",
		Property: properties[0], // Mock - using first property
	}

	renderTemplate(w, "property-detail.html", data)
}

func handleTenants(w http.ResponseWriter, r *http.Request) {
	properties, err := models.GetProperties()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Title      string
		Properties []models.PropertyDetail
	}{
		Title:      "Tenant Management",
		Properties: properties,
	}
	renderTemplate(w, "tenants.html", data)
}

func handleMaintenance(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
	}{
		Title: "Maintenance Requests",
	}
	renderTemplate(w, "maintenance.html", data)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// Parse the base template and the specific template
	t, err := template.ParseFiles("templates/base.html", "templates/"+tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// The following handlers are placeholders.  Implement the logic to
// create, update, and delete properties as needed.

func handleCreateProperty(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("Name")
	address := r.FormValue("Address")
	propertyType := r.FormValue("PropertyType")

	if name == "" || address == "" || propertyType == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	tagsString := r.FormValue("Tags")
	tags := strings.Split(tagsString, ",")

	property := &models.Property{
		Name:         name,
		Address:      address,
		PropertyType: propertyType,
		Tags:         tags,
	}

	if err := models.CreateProperty(property); err != nil {
		http.Error(w, "Error creating property", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/properties", http.StatusSeeOther)
}

func handleUpdateProperty(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Update Property endpoint")
}

func handleDeleteProperty(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Delete Property endpoint")
}

func handleNewPropertyForm(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "property-form.html", nil)
}

func handleImportPropertyForm(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "property-import.html", nil)
}

func handleImportProperty(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form with a maximum file size of 10MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the file from the form data
	file, handler, err := r.FormFile("csvFile")
	if err != nil {
		http.Error(w, "Error retrieving file from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check if the file is a CSV file
	if filepath.Ext(handler.Filename) != ".csv" {
		http.Error(w, "Invalid file type. Only CSV files are allowed.", http.StatusBadRequest)
		return
	}

	// Create a temporary file to store the uploaded CSV file
	tempFile, err := os.CreateTemp("", "upload-*.csv")
	if err != nil {
		http.Error(w, "Error creating temporary file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, "Error copying file to temporary file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Open the temporary file for reading
	csvFile, err := os.Open(tempFile.Name())
	if err != nil {
		http.Error(w, "Error opening temporary file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer csvFile.Close()

	// Read the CSV file
	reader := csv.NewReader(csvFile)
	reader.Comma = ',' // Set the delimiter to comma
	reader.TrimLeadingSpace = true

	// Read the header row
	header, err := reader.Read()
	if err != nil {
		http.Error(w, "Error reading header row: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate the header row
	if len(header) != 3 || header[0] != "Name" || header[1] != "Address" || header[2] != "PropertyType" {
		http.Error(w, "Invalid CSV header format. The header should be: Name,Address,PropertyType", http.StatusBadRequest)
		return
	}

	// Read the data rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Error reading CSV row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Validate the data row
		if len(row) != 3 {
			http.Error(w, "Invalid CSV row format. Each row should have 3 columns.", http.StatusBadRequest)
			return
		}

		name := row[0]
		address := row[1]
		propertyType := row[2]

		// Create a new property
		property := &models.Property{
			Name:         name,
			Address:      address,
			PropertyType: propertyType,
		}

		// Create the property in the database
		if err := models.CreateProperty(property); err != nil {
			http.Error(w, "Error creating property: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Redirect to the properties page
	http.Redirect(w, r, "/properties", http.StatusSeeOther)
}

func handleProfilePage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		User  *models.User
	}{
		Title: "User Profile",
		User:  user,
	}

	renderTemplate(w, "profile.html", data)
}

func handleAdminUsersPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
	}{
		Title: "User Management",
	}

	renderTemplate(w, "admin-users.html", data)
}

func handleReportsPage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		User  *models.User
	}{
		Title: "Reports & Analytics",
		User:  user,
	}

	renderTemplate(w, "reports.html", data)
}

func handleAnalyticsPage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		User  *models.User
	}{
		Title: "Analytics Dashboard",
		User:  user,
	}

	renderTemplate(w, "analytics.html", data)
}

func handleCreateReportPage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
		User  *models.User
	}{
		Title: "Create Report",
		User:  user,
	}

	renderTemplate(w, "create-report.html", data)
}

func handleViewReportPage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	reportID := chi.URLParam(r, "id")

	data := struct {
		Title    string
		User     *models.User
		ReportID string
	}{
		Title:    "View Report",
		User:     user,
		ReportID: reportID,
	}

	renderTemplate(w, "view-report.html", data)
}
