-- 1. Properties Table: Represents a physical building or property.
CREATE TABLE properties (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    property_type VARCHAR(50) NOT NULL, -- e.g., 'Apartment Building', 'Single Family Home'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE properties ADD COLUMN tags TEXT[];


-- 2. Property Units Table: Represents individual units within a property.
CREATE TABLE property_units (
    id SERIAL PRIMARY KEY,
    property_id INT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    unit_number VARCHAR(50), -- e.g., 'Apt 101', 'Unit B'
    bedrooms INT NOT NULL DEFAULT 1,
    bathrooms INT NOT NULL DEFAULT 1,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Tenants Table: Stores information about individual tenants.
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20),
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- e.g., 'active', 'archived'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Leases Table: A linking table that connects tenants to a specific unit for a period of time.
CREATE TABLE leases (
    id SERIAL PRIMARY KEY,
    unit_id INT NOT NULL REFERENCES property_units(id) ON DELETE RESTRICT,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    monthly_rent DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- e.g., 'active', 'ended', 'pending'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Payments Table: Records all payments made by tenants.
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    lease_id INT NOT NULL REFERENCES leases(id) ON DELETE RESTRICT,
    amount DECIMAL(10, 2) NOT NULL,
    payment_date DATE NOT NULL,
    payment_method VARCHAR(50), -- e.g., 'Credit Card', 'Bank Transfer'
    status VARCHAR(50) NOT NULL DEFAULT 'completed', -- e.g., 'completed', 'pending', 'failed'
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 6. Maintenance Requests Table: Tracks maintenance issues for properties.
CREATE TABLE maintenance_requests (
    id SERIAL PRIMARY KEY,
    property_id INT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    reported_by_tenant_id INT REFERENCES tenants(id) ON DELETE SET NULL, -- A tenant might report it
    description TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'reported', -- e.g., 'reported', 'in_progress', 'completed'
    priority VARCHAR(50) DEFAULT 'medium', -- e.g., 'low', 'medium', 'high'
    reported_date DATE NOT NULL DEFAULT CURRENT_DATE,
    completed_date DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
