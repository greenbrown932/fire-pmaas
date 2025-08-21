package api

import (
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

	r.Group(func(auth chi.Router) {
		auth.Use(middleware.RequireLogin)
		auth.Get("/", handleDashboard)
		auth.Get("/properties", handleProperties)
		auth.Get("/properties/{id}", handlePropertyDetail)
		auth.Get("/tenants", handleTenants)
		auth.Get("/maintenance", handleMaintenance)
		auth.Get("/callback", middleware.HandleCallback)
	})

}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title      string
		Properties []models.Property
		Stats      struct {
			TotalProperties int
			Occupied        int
			Vacant          int
			Maintenance     int
			MonthlyRevenue  int
		}
	}{
		Title:      "Property Management Dashboard",
		Properties: models.Properties,
	}

	// Calculate stats
	for _, p := range models.Properties {
		data.Stats.TotalProperties++
		switch p.Status {
		case "Occupied":
			data.Stats.Occupied++
			data.Stats.MonthlyRevenue += p.Rent
		case "Vacant":
			data.Stats.Vacant++
		case "Maintenance":
			data.Stats.Maintenance++
		}
	}

	renderTemplate(w, "dashboard.html", data)
}

func handleProperties(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title      string
		Properties []models.Property
	}{
		Title:      "All Properties",
		Properties: models.Properties,
	}
	renderTemplate(w, "properties.html", data)
}

func handlePropertyDetail(w http.ResponseWriter, r *http.Request) {
	// In a real app, you'd parse the ID and look up the property
	data := struct {
		Title    string
		Property models.Property
	}{
		Title:    "Property Details",
		Property: models.Properties[0], // Mock - using first property
	}
	renderTemplate(w, "property-detail.html", data)
}

func handleTenants(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title      string
		Properties []models.Property
	}{
		Title:      "Tenant Management",
		Properties: models.Properties,
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
	}
}
