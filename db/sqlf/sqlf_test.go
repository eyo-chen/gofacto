package sqlf

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/eyo-chen/gofacto"
	"github.com/eyo-chen/gofacto/internal/docker"
	"github.com/eyo-chen/gofacto/internal/testutils"
	"github.com/eyo-chen/gofacto/utils"
	_ "github.com/go-sql-driver/mysql"
)

var (
	mockCTX = context.Background()
)

type Author struct {
	ID                  int64
	FirstName           string
	LastName            string
	BirthDate           *time.Time
	Nationality         *string
	Email               *string
	Biography           *string
	IsActive            bool
	Rating              *float64
	BooksWritten        *int32
	LastPublicationTime time.Time
	WebsiteURL          *string
	FanCount            *int64
	ProfilePicture      []byte
}

type Book struct {
	ID              int64
	AuthorID        int64 `gofacto:"Author,authors"`
	Title           string
	ISBN            *string
	PublicationDate *time.Time
	Genre           *string
	Price           *float64
	PageCount       *int32
	Description     *string
	InStock         bool
	CoverImage      []byte
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func setupTest() *sql.DB {
	port := docker.RunDocker(docker.ImageMySQL)
	dba, err := sql.Open("mysql", fmt.Sprintf("root:root@(localhost:%s)/mysql?parseTime=true", port))
	if err != nil {
		log.Fatalf("sql.Open failed: %s", err)
	}

	// Read SQL file
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("Failed to read schema.sql: %s", err)
	}

	// Split SQL file content into individual statements
	queries := strings.Split(string(schema), ";")

	// Execute SQL statements one by one
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		if _, err := dba.Exec(query); err != nil {
			log.Fatalf("Failed to execute query: %s, error: %s", query, err)
		}
	}

	return dba
}

func tearDownTest(db *sql.DB) {
	db.Close()
	docker.PurgeDocker()
}

func TestInsert(t *testing.T) {
	// prepare setup
	db := setupTest()
	defer tearDownTest(db)
	f := gofacto.New(Author{}).SetConfig(gofacto.Config[Author]{
		DB: &Config{DB: db},
	})

	// prepare mock data
	mockAuthor, err := f.Build(mockCTX).Insert()
	if err != nil {
		t.Fatalf("Failed to insert author: %s", err)
	}

	// verify the inserted data
	stmt := "SELECT * FROM authors WHERE id = ?"
	row := db.QueryRow(stmt, mockAuthor.ID)
	var author Author
	if err := row.Scan(
		&author.ID,
		&author.FirstName,
		&author.LastName,
		&author.BirthDate,
		&author.Nationality,
		&author.Email,
		&author.Biography,
		&author.IsActive,
		&author.Rating,
		&author.BooksWritten,
		&author.LastPublicationTime,
		&author.WebsiteURL,
		&author.FanCount,
		&author.ProfilePicture,
	); err != nil {
		t.Fatalf("Failed to scan author: %s", err)
	}

	// assertion
	if err := testutils.CompareVal(mockAuthor, author, "BirthDate", "LastPublicationTime"); err != nil {
		t.Fatalf("Inserted author is not the same as the mock author: %s", err)
	}
}

func TestInsertList(t *testing.T) {
	// prepare setup
	db := setupTest()
	defer tearDownTest(db)
	f := gofacto.New(Author{}).SetConfig(gofacto.Config[Author]{
		DB: &Config{DB: db},
	})

	// prepare mock data
	mockAuthors, err := f.BuildList(mockCTX, 3).Insert()
	if err != nil {
		t.Fatalf("Failed to insert books: %s", err)
	}

	// verify the inserted data
	stmt := "SELECT * FROM authors"
	rows, err := db.Query(stmt)
	if err != nil {
		t.Fatalf("Failed to query authors: %s", err)
	}
	defer rows.Close()

	var authors []Author
	for rows.Next() {
		var author Author
		if err := rows.Scan(
			&author.ID,
			&author.FirstName,
			&author.LastName,
			&author.BirthDate,
			&author.Nationality,
			&author.Email,
			&author.Biography,
			&author.IsActive,
			&author.Rating,
			&author.BooksWritten,
			&author.LastPublicationTime,
			&author.WebsiteURL,
			&author.FanCount,
			&author.ProfilePicture,
		); err != nil {
			t.Fatalf("Failed to scan author: %s", err)
		}

		authors = append(authors, author)
	}

	// assertion
	if err := testutils.CompareVal(mockAuthors, authors, "BirthDate", "LastPublicationTime"); err != nil {
		t.Fatalf("Inserted authors are not the same as the mock authors: %s", err)
	}
}

func TestWithOne(t *testing.T) {
	// prepare setup
	db := setupTest()
	defer tearDownTest(db)
	f := gofacto.New(Book{}).SetConfig(gofacto.Config[Book]{
		DB: &Config{DB: db},
	})

	// prepare mock data
	mockAuthor := Author{}
	mockGenre := "Science"
	ow := Book{Genre: &mockGenre} // set correct enum value
	mockBook, err := f.Build(mockCTX).Overwrite(ow).WithOne(&mockAuthor).Insert()
	if err != nil {
		t.Fatalf("Failed to insert: %s", err)
	}

	// verify the inserted data
	bookStmt := "SELECT * FROM books WHERE author_id = ?"
	bookRow := db.QueryRow(bookStmt, mockBook.AuthorID)
	var book Book
	if err := bookRow.Scan(
		&book.ID,
		&book.AuthorID,
		&book.Title,
		&book.ISBN,
		&book.PublicationDate,
		&book.Genre,
		&book.Price,
		&book.PageCount,
		&book.Description,
		&book.InStock,
		&book.CoverImage,
		&book.CreatedAt,
		&book.UpdatedAt,
	); err != nil {
		t.Fatalf("Failed to scan book: %s", err)
	}

	authorStmt := "SELECT * FROM authors WHERE id = ?"
	authorRow := db.QueryRow(authorStmt, mockBook.AuthorID)
	var author Author
	if err := authorRow.Scan(
		&author.ID,
		&author.FirstName,
		&author.LastName,
		&author.BirthDate,
		&author.Nationality,
		&author.Email,
		&author.Biography,
		&author.IsActive,
		&author.Rating,
		&author.BooksWritten,
		&author.LastPublicationTime,
		&author.WebsiteURL,
		&author.FanCount,
		&author.ProfilePicture,
	); err != nil {
		t.Fatalf("Failed to scan author: %s", err)
	}

	// assertion
	if err := testutils.CompareVal(mockBook, book, "PublicationDate", "CreatedAt", "UpdatedAt"); err != nil {
		t.Fatalf("Inserted book is not the same as the mock book: %s", err)
	}

	if err := testutils.CompareVal(mockAuthor, author, "BirthDate", "LastPublicationTime"); err != nil {
		t.Fatalf("Inserted author is not the same as the mock author: %s", err)
	}
}

func TestWithMany(t *testing.T) {
	// prepare setup
	db := setupTest()
	defer tearDownTest(db)
	f := gofacto.New(Book{}).SetConfig(gofacto.Config[Book]{
		DB: &Config{DB: db},
	})

	// prepare mock data
	mockAnyAuthors := utils.CvtToAnysWithOWs[Author](3)
	mockGenre := "Science"
	ow := Book{Genre: &mockGenre} // set correct enum value
	mockBooks, err := f.BuildList(mockCTX, 3).Overwrite(ow).WithMany(mockAnyAuthors).Insert()
	if err != nil {
		t.Fatalf("Failed to insert books: %s", err)
	}

	mockAuthors := utils.CvtToT[Author](mockAnyAuthors)

	// verify the inserted data
	bookStmt := "SELECT * FROM books WHERE author_id = ?"
	authorStmt := "SELECT * FROM authors WHERE id = ?"

	for i := 0; i < 3; i++ {
		bookRow := db.QueryRow(bookStmt, mockBooks[i].AuthorID)
		var book Book
		if err := bookRow.Scan(
			&book.ID,
			&book.AuthorID,
			&book.Title,
			&book.ISBN,
			&book.PublicationDate,
			&book.Genre,
			&book.Price,
			&book.PageCount,
			&book.Description,
			&book.InStock,
			&book.CoverImage,
			&book.CreatedAt,
			&book.UpdatedAt,
		); err != nil {
			t.Fatalf("Failed to scan book: %s", err)
		}

		authorRow := db.QueryRow(authorStmt, mockBooks[i].AuthorID)
		var author Author
		if err := authorRow.Scan(
			&author.ID,
			&author.FirstName,
			&author.LastName,
			&author.BirthDate,
			&author.Nationality,
			&author.Email,
			&author.Biography,
			&author.IsActive,
			&author.Rating,
			&author.BooksWritten,
			&author.LastPublicationTime,
			&author.WebsiteURL,
			&author.FanCount,
			&author.ProfilePicture,
		); err != nil {
			t.Fatalf("Failed to scan author: %s", err)
		}

		// assertion
		if err := testutils.CompareVal(mockBooks[i], book, "PublicationDate", "CreatedAt", "UpdatedAt"); err != nil {
			t.Fatalf("Inserted book is not the same as the mock book: %s", err)
		}

		if err := testutils.CompareVal(mockAuthors[i], author, "BirthDate", "LastPublicationTime"); err != nil {
			t.Fatalf("Inserted author is not the same as the mock author: %s", err)
		}
	}
}

func TestListWithOne(t *testing.T) {
	// prepare setup
	db := setupTest()
	defer tearDownTest(db)
	f := gofacto.New(Book{}).SetConfig(gofacto.Config[Book]{
		DB: &Config{DB: db},
	})

	// prepare mock data
	mockAuthor := Author{}
	mockGenre := "Science"
	ow := Book{Genre: &mockGenre} // set correct enum value
	mockBooks, err := f.BuildList(mockCTX, 3).Overwrite(ow).WithOne(&mockAuthor).Insert()
	if err != nil {
		t.Fatalf("Failed to insert books: %s", err)
	}

	// verify the inserted data
	bookStmt := "SELECT * FROM books WHERE author_id = ?"
	bookRows, err := db.Query(bookStmt, mockBooks[0].AuthorID)
	if err != nil {
		t.Fatalf("Failed to query books: %s", err)
	}

	var books []Book
	for bookRows.Next() {
		var book Book
		if err := bookRows.Scan(
			&book.ID,
			&book.AuthorID,
			&book.Title,
			&book.ISBN,
			&book.PublicationDate,
			&book.Genre,
			&book.Price,
			&book.PageCount,
			&book.Description,
			&book.InStock,
			&book.CoverImage,
			&book.CreatedAt,
			&book.UpdatedAt,
		); err != nil {
			t.Fatalf("Failed to scan book: %s", err)
		}

		books = append(books, book)
	}

	authorStmt := "SELECT * FROM authors WHERE id = ?"
	authorRow := db.QueryRow(authorStmt, mockBooks[0].AuthorID)
	var author Author
	if err := authorRow.Scan(
		&author.ID,
		&author.FirstName,
		&author.LastName,
		&author.BirthDate,
		&author.Nationality,
		&author.Email,
		&author.Biography,
		&author.IsActive,
		&author.Rating,
		&author.BooksWritten,
		&author.LastPublicationTime,
		&author.WebsiteURL,
		&author.FanCount,
		&author.ProfilePicture,
	); err != nil {
		t.Fatalf("Failed to scan author: %s", err)
	}

	// assertion
	if err := testutils.CompareVal(mockBooks, books, "PublicationDate", "CreatedAt", "UpdatedAt"); err != nil {
		t.Fatalf("Inserted books are not the same as the mock books: %s", err)
	}

	if err := testutils.CompareVal(mockAuthor, author, "BirthDate", "LastPublicationTime"); err != nil {
		t.Fatalf("Inserted author is not the same as the mock author: %s", err)
	}
}
