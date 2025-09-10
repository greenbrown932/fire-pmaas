package models

import (
	"time"

	"github.com/greenbrown932/fire-pmaas/pkg/db"
)

// Property represents a property in our system.
type Property struct {
	ID           int // Unique identifier for the property
	Name         string
	Address      string
	PropertyType string
	CreatedAt    time.Time
	UpdatedAt    time.Time // Time the property was last updated
}

// PropertyUnit represents an individual unit within a property.
type PropertyUnit struct {
	ID          int // Unique identifier for the property unit
	PropertyID  int
	UnitNumber  string
	Bedrooms    int
	Bathrooms   int
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time // Time the property unit was last updated
}

// Tenant stores information about individual tenants.
type Tenant struct {
	ID          int // Unique identifier for the tenant
	FirstName   string
	LastName    string
	Email       string
	PhoneNumber string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time // Time the tenant was last updated
}

// Lease connects a tenant to a specific unit for a period of time.
type Lease struct {
	ID          int // Unique identifier for the lease
	UnitID      int
	TenantID    int
	StartDate   time.Time
	EndDate     time.Time
	MonthlyRent float64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time // Time the lease was last updated
}

// PropertyDetail represents a detailed view of a property, including lease and tenant information.
type PropertyDetail struct {
	ID         int    // Unique identifier for the property
	Address    string // Address of the property
	Rent       float64
	Status     string
	Bedrooms   int
	Bathrooms  int
	TenantName string // Name of the tenant currently leasing the property
}

// CreateProperty creates a new property in the database.
func CreateProperty(property *Property) error {
	_, err := db.DB.Exec(`
		INSERT INTO properties (name, address, property_type)
		VALUES ($1, $2, $3)
	`, property.Name, property.Address, property.PropertyType)
	return err
}

// UpdateProperty updates an existing property in the database.
func UpdateProperty(property *Property) error {
	_, err := db.DB.Exec(`
		UPDATE properties
		SET name = $1, address = $2, property_type = $3
		WHERE id = $4
	`, property.Name, property.Address, property.PropertyType, property.ID)
	return err
}

// DeleteProperty deletes a property from the database.
func DeleteProperty(id int) error {
	_, err := db.DB.Exec(`
		DELETE FROM properties
		WHERE id = $1
	`, id)
	return err
}

// GetProperties retrieves a list of properties with details including address, rent, status, and tenant name.
func GetProperties() ([]PropertyDetail, error) {
	// Execute the SQL query to retrieve property details.
	rows, err := db.DB.Query(`
		SELECT
			p.id,                             -- Property ID
			p.address,                        -- Property Address
			COALESCE(l.monthly_rent, 0),      -- Lease Monthly Rent (default to 0 if NULL)
			COALESCE(l.status, 'vacant'),     -- Lease Status (default to 'vacant' if NULL)
			COALESCE(pu.bedrooms, 0),         -- Property Unit Bedrooms (default to 0 if NULL)
			COALESCE(pu.bathrooms, 0),        -- Property Unit Bathrooms (default to 0 if NULL)
			COALESCE(t.first_name || ' ' || t.last_name, '') AS tenant_name -- Tenant Full Name (empty if NULL)
		FROM properties p                                     -- From the properties table
		LEFT JOIN property_units pu ON p.id = pu.property_id   -- Join with property_units table
		LEFT JOIN leases l ON pu.id = l.unit_id AND l.status = 'active' -- Only active leases
		LEFT JOIN tenants t ON l.tenant_id = t.id             -- Join with tenants table
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var properties []PropertyDetail
	for rows.Next() {
		var p PropertyDetail
		if err := rows.Scan(&p.ID, &p.Address, &p.Rent, &p.Status, &p.Bedrooms, &p.Bathrooms, &p.TenantName); err != nil {
			return nil, err
		}
		properties = append(properties, p)
	}
	return properties, nil
}
