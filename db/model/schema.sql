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
    barcode VARCHAR(100),
    store_owner_id INTEGER REFERENCES store_owners(id),
    customer_phone VARCHAR(20),
    note VARCHAR(255),
    from_city VARCHAR(100),
    to_city VARCHAR(100),
    to_subcity VARCHAR(100),
    to_specific_location VARCHAR(255),
    status VARCHAR(50),
    current_location VARCHAR(100),
    price INTEGER,
    fdelivery_fee INTEGER,
    total_price INTEGER,
    warehouse_id INTEGER REFERENCES warehouses(id),
    created_at TIMESTAMP DEFAULT now()
);
