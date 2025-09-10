-- Users table for storing user profiles and authentication data
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    keycloak_id VARCHAR(255) UNIQUE, -- Keycloak user ID for external auth
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20),
    profile_picture_url TEXT,
    email_verified BOOLEAN DEFAULT FALSE,
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(255), -- For TOTP MFA
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'suspended', 'inactive'
    last_login TIMESTAMPTZ,
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Roles table for defining user roles and permissions
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL, -- 'admin', 'property_manager', 'tenant', 'viewer'
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions TEXT[], -- Array of permission strings
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User roles junction table (many-to-many relationship)
CREATE TABLE user_roles (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    assigned_by INT REFERENCES users(id),
    UNIQUE(user_id, role_id)
);

-- User sessions table for tracking user sessions
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add user_id to tenants table to link tenants with user accounts
ALTER TABLE tenants ADD COLUMN user_id INT REFERENCES users(id) ON DELETE SET NULL;

-- Insert default roles
INSERT INTO roles (name, display_name, description, permissions) VALUES
('admin', 'Administrator', 'Full system access with all permissions', ARRAY[
    'users.create', 'users.read', 'users.update', 'users.delete',
    'properties.create', 'properties.read', 'properties.update', 'properties.delete',
    'tenants.create', 'tenants.read', 'tenants.update', 'tenants.delete',
    'leases.create', 'leases.read', 'leases.update', 'leases.delete',
    'payments.create', 'payments.read', 'payments.update', 'payments.delete',
    'maintenance.create', 'maintenance.read', 'maintenance.update', 'maintenance.delete',
    'roles.manage', 'system.settings'
]),
('property_manager', 'Property Manager', 'Manage properties, tenants, and maintenance', ARRAY[
    'properties.create', 'properties.read', 'properties.update', 'properties.delete',
    'tenants.create', 'tenants.read', 'tenants.update', 'tenants.delete',
    'leases.create', 'leases.read', 'leases.update', 'leases.delete',
    'payments.read', 'payments.update',
    'maintenance.create', 'maintenance.read', 'maintenance.update', 'maintenance.delete'
]),
('tenant', 'Tenant', 'View own information and submit maintenance requests', ARRAY[
    'profile.read', 'profile.update',
    'lease.read.own', 'payments.read.own',
    'maintenance.create.own', 'maintenance.read.own'
]),
('viewer', 'Viewer', 'Read-only access to basic information', ARRAY[
    'properties.read', 'tenants.read', 'maintenance.read'
]);

-- Create indexes for better performance
CREATE INDEX idx_users_keycloak_id ON users(keycloak_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at);
