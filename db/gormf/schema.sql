CREATE TABLE IF NOT EXISTS authors (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    birth_date DATE,
    nationality VARCHAR(50),
    email VARCHAR(100) UNIQUE,
    biography TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    rating DECIMAL(3,2),
    books_written INT UNSIGNED,
    last_publication_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    website_url VARCHAR(255),
    fan_count BIGINT UNSIGNED,
    profile_picture BLOB
);

CREATE TABLE IF NOT EXISTS categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    author_id INT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS sub_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    category_id INT NOT NULL,
    author_id INT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS books (
    id INT AUTO_INCREMENT PRIMARY KEY,
    author_id INT,
    category_id INT,
    sub_category_id INT,
    title VARCHAR(255) NOT NULL,
    isbn CHAR(13) UNIQUE,
    publication_date DATE,
    genre ENUM('Fiction', 'Non-Fiction', 'Science', 'History', 'Biography', 'Other'),
    price DECIMAL(10,2),
    page_count SMALLINT UNSIGNED,
    description TEXT,
    in_stock BOOLEAN DEFAULT TRUE,
    cover_image MEDIUMBLOB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE SET NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
    FOREIGN KEY (sub_category_id) REFERENCES sub_categories(id) ON DELETE SET NULL
);