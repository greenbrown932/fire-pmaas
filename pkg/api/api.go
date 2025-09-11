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

	// Register user API routes (includes public and protected routes)
	RegisterUserRoutes(r)

	// API Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})
	r.Get("/callback", middleware.HandleCallback)

	// Web routes (protected by authentication)
	r.Group(func(auth chi.Router) {
		auth.Use(middleware.LoadUserFromToken)
		auth.Use(middleware.RequireLogin)

		// Dashboard and UI routes
		auth.Get("/", handleDashboard)
		auth.Get("/profile", handleProfilePage)

		// Properties routes with role-based access
		auth.Group(func(props chi.Router) {
			props.Use(middleware.RequireAnyRole("admin", "property_manager", "viewer"))
			props.Get("/properties", handleProperties)
			props.Get("/properties/{id}", handlePropertyDetail)
		})

		// Property management routes (create, update, delete)
		auth.Group(func(propsMgmt chi.Router) {
			propsMgmt.Use(middleware.RequireAnyRole("admin", "property_manager"))
			propsMgmt.Post("/properties", handleCreateProperty)
			propsMgmt.Put("/properties/{id}", handleUpdateProperty)
			propsMgmt.Delete("/properties/{id}", handleDeleteProperty)
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
