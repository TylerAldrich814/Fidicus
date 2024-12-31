-- 001_init.up.sql

CREATE TYPE role AS ENUM (
  'access_role_unspecified',
  'access_role_entity',
  'access_role_admin',
  'access_role_account',
  'access_role_read_only'
);

CREATE TABLE entities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Unique Entity ID
  name VARCHAR(256) UNIQUE NOT NULL,             -- Entity Name
  description TEXT,                              -- Optional Entity Description
  account_ids UUID[] default '{}',               -- Entity Account ID's -- Account's associated with this Entity
  created_at TIMESTAMP,                          -- Datetime - When Entity account was created
  updated_at TIMESTAMP                           -- Datetime - when Entity account was last updated.
);

CREATE TABLE accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),      -- Accounts Unique ID
  entity_id UUID NOT NULL,                            -- EntityID References an Entity. Declaring that this account is a subaccount of Entity.
  email VARCHAR(256) UNIQUE NOT NULL,                 -- Accounts Contact Email
  password_hash VARCHAR(256) NOT NULL,                -- Accounts Hashed Password
  role role NOT NULL DEFAULT 'access_role_read_only', -- Accounts Role with parent Entity
  first_name VARCHAR(256),                            -- Accounts First name
  last_name  VARCHAR(256),                            -- Accounts Last name
  cellphone_number VARCHAR(32),                       -- Account's Cellphone Number
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,     -- Datetime - When Account Account was created
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,     -- Datetime - When Account Account was last updated.
  FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);

CREATE TABLE tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id UUID NOT NULL,
  refresh_token TEXT UNIQUE NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE TABLE schemas (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),   -- Schema Unique ID
  entity_id UUID NOT NULL,                         -- Foreign key - Links to parent Entity's ID
  name VARCHAR(256),                               -- Schema's Name
  description TEXT,                                -- Description of the Schema
  blob_url    TEXT NOT NULL,                       -- Blob Storage location
  graph_url   TEXT NOT NULL,                       -- GraphQL Storge Location
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Datetime - When Schema was created
  updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Datetime - When Schema was last updated.
  FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);

CREATE TABLE permissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),  -- A Unique ID refrences a Account's permission.
  entity_id  UUID NOT NULL,                       -- References an Entity's ID
  account_id UUID NOT NULL,                       -- References a account ID
  role role NOT NULL,                             -- The Accounts Role
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Datetime - When Schema was created
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Datetime - When Schema was last updated.
  FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE,
  FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);
