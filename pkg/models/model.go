package models

var Properties = []Property{
	{1, "123 Main St, Boston, MA", 2500, "Occupied", 2, 1, 900, "John Smith"},
	{2, "456 Oak Ave, Cambridge, MA", 3200, "Vacant", 3, 2, 1200, ""},
	{3, "789 Pine St, Somerville, MA", 2800, "Occupied", 2, 2, 1000, "Sarah Johnson"},
	{4, "321 Elm Dr, Brighton, MA", 2200, "Maintenance", 1, 1, 700, ""},
}

// Property represents a property in our system
type Property struct {
	ID         int
	Address    string
	Rent       int
	Status     string
	Bedrooms   int
	Bathrooms  int
	SquareFeet int
	Tenant     string
}
