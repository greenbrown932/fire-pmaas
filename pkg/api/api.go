package api

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/greenbrown932/fire-pmaas/pkg/middleware"
	"github.com/greenbrown932/fire-pmaas/pkg/models"
)

func RegisterRoutes(r *chi.Mux) {
	// Serve static files (CSS, JS, images)
	workDir, _ := filepath.Abs("./")
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(filesDir)))

	// API Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	r.Get("/callback", middleware.HandleCallback)
	r.Group(func(auth chi.Router) {
		auth.Use(middleware.RequireLogin)
		auth.Get("/", handleDashboard)
		auth.Get("/properties", handleProperties)
		auth.Get("/properties/{id}", handlePropertyDetail)
		auth.Post("/properties", handleCreateProperty)
		auth.Put("/properties/{id}", handleUpdateProperty)
		auth.Delete("/properties/{id}", handleDeleteProperty)
		auth.Get("/tenants", handleTenants)
		auth.Get("/maintenance", handleMaintenance)

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
	properties, err := models.GetProperties()
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
	fmt.Fprint(w, "Create Property endpoint")
}

func handleUpdateProperty(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Update Property endpoint")
}

func handleDeleteProperty(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Delete Property endpoint")
}
