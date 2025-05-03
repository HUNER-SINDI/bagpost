CREATE TABLE admins (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);


CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    fullname VARCHAR(100) NOT NULL,
    locname VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    phone VARCHAR(20) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE warehousesetter (
    id SERIAL PRIMARY KEY,
    krd VARCHAR(100) NOT NULL,
    ar VARCHAR(100) NOT NULL,
    warehouse_id INTEGER NOT NULL UNIQUE,
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE
);



CREATE TABLE store_owners (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    email VARCHAR(255) UNIQUE,
    password TEXT,
    location_city VARCHAR(100),
    location_address TEXT,
    warehouse_id INTEGER REFERENCES warehouses(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE store_balances (
    id SERIAL PRIMARY KEY,
    store_owner_id INTEGER UNIQUE REFERENCES store_owners(id) ON DELETE CASCADE,
    
    in_store_balance INTEGER DEFAULT 0,
    pending_balance INTEGER DEFAULT 0,
    paid_balance INTEGER DEFAULT 0,
    refused_balance INTEGER DEFAULT 0,

    updated_at TIMESTAMP DEFAULT now()
);


CREATE TABLE storesetter (
    id SERIAL PRIMARY KEY,
    krd VARCHAR(100) NOT NULL,
    ar VARCHAR(100) NOT NULL
);


CREATE TABLE drivers (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    email VARCHAR(255) UNIQUE,
    password TEXT,
    location_city VARCHAR(100),
    location_address TEXT,
    setter_location VARCHAR(255),
    warehouse_id INTEGER REFERENCES warehouses(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE city_routes (
    id SERIAL PRIMARY KEY,
    warehouse_id INTEGER REFERENCES warehouses(id),
    city_name VARCHAR(100),
    price INTEGER NOT NULL
);

CREATE TABLE subcities (
    id SERIAL PRIMARY KEY,
    city_route_id INTEGER REFERENCES city_routes(id),
    subcity_name VARCHAR(100),
    price INTEGER
);

CREATE TABLE deliveries (
    id SERIAL PRIMARY KEY,
    barcode VARCHAR(100) NOT NULL UNIQUE,
    store_owner_id INTEGER NOT NULL REFERENCES store_owners(id),
    customer_phone VARCHAR(20) NOT NULL,
    note VARCHAR(255),  -- Nullable
    from_city VARCHAR(100) NOT NULL,
    to_city VARCHAR(100) NOT NULL,
    to_subcity VARCHAR(100) NOT NULL,
    to_specific_location VARCHAR(255),  -- Nullable
    status VARCHAR(50) NOT NULL,
    price INTEGER NOT NULL,
    fdelivery_fee INTEGER NOT NULL,
    total_price INTEGER NOT NULL,
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
    created_at TIMESTAMP DEFAULT now() NOT NULL
);


CREATE TABLE delivery_routing (
    id SERIAL PRIMARY KEY,
    delivery_id INTEGER NOT NULL REFERENCES deliveries(id) ON DELETE CASCADE,
    setter_krd VARCHAR(100) NOT NULL,
    setter_ar VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE delivery_transfers (
    id SERIAL PRIMARY KEY,
    delivery_id INTEGER REFERENCES deliveries(id) NOT NULL,
    origin_warehouse_id INTEGER REFERENCES warehouses(id) NOT NULL,
    current_warehouse_id INTEGER REFERENCES warehouses(id) NOT NULL,
    transfer_status VARCHAR(50) NOT NULL,
    driver_id INTEGER REFERENCES drivers(id) NULL,
    transferred_at TIMESTAMP DEFAULT now(),
    received_at TIMESTAMP NULL,
    transfer_note TEXT
);



CREATE TABLE ads (
    id SERIAL PRIMARY KEY,
    url VARCHAR(255)
);