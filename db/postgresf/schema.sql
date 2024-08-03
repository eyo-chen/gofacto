CREATE TABLE IF NOT EXISTS authors (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    birth_date DATE,
    nationality VARCHAR(50),
    email VARCHAR(100) UNIQUE,
    biography TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    rating NUMERIC(3,2),
    books_written INTEGER,
    last_publication_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    website_url VARCHAR(255),
    fan_count BIGINT,
    profile_picture BYTEA
);

CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    author_id INTEGER,
    title VARCHAR(255) NOT NULL,
    isbn CHAR(13) UNIQUE,
    publication_date DATE,
    genre TEXT CHECK (genre IN ('Fiction', 'Non-Fiction', 'Science', 'History', 'Biography', 'Other')),
    price NUMERIC(10,2),
    page_count SMALLINT,
    description TEXT,
    in_stock BOOLEAN DEFAULT TRUE,
    cover_image BYTEA,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE SET NULL
);