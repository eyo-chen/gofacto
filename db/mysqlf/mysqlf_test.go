package mysqlf

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/eyo-chen/gofacto"
	"github.com/eyo-chen/gofacto/internal/docker"
	"github.com/eyo-chen/gofacto/internal/testutils"
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

type testingSuite struct {
	db      *sql.DB
	authorF *gofacto.Factory[Author]
	bookF   *gofacto.Factory[Book]
}

func (s *testingSuite) setupSuite() {
	// start MySQL Docker container
	port := docker.RunDocker(docker.ImageMySQL)
	dba, err := sql.Open("mysql", fmt.Sprintf("root:root@(localhost:%s)/mysql?parseTime=true", port))
	if err != nil {
		log.Fatalf("sql.Open failed: %s", err)
	}

	// set the database connection
	s.db = dba

	// read SQL file
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("Failed to read schema.sql: %s", err)
	}

	// split SQL file content into individual statements
	queries := strings.Split(string(schema), ";")

	// execute SQL statements one by one
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		if _, err := dba.Exec(query); err != nil {
			log.Fatalf("Failed to execute query: %s, error: %s", query, err)
		}
	}

	// set up gofacto factories
	s.authorF = gofacto.New(Author{}).SetConfig(gofacto.Config[Author]{
		DB: NewConfig(s.db),
	})
	s.bookF = gofacto.New(Book{}).SetConfig(gofacto.Config[Book]{
		DB: NewConfig(s.db),
	})
}

func (s *testingSuite) tearDownSuite() error {
	if err := s.db.Close(); err != nil {
		return err
	}

	docker.PurgeDocker()

	return nil
}

func (s *testingSuite) tearDownTest() error {
	if _, err := s.db.Exec("DELETE FROM authors"); err != nil {
		return err
	}

	if _, err := s.db.Exec("DELETE FROM books"); err != nil {
		return err
	}

	s.authorF.Reset()
	s.bookF.Reset()

	return nil
}

func (s *testingSuite) Run(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"TestInsert", s.TestInsert},
		{"TestInsertList", s.TestInsertList},
		{"TestWithOne", s.TestWithOne},
		{"TestWithMany", s.TestWithMany},
		{"TestListWithOne", s.TestListWithOne},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.fn(t)
			if err := s.tearDownTest(); err != nil {
				t.Fatalf("Failed to tear down test: %s", err)
			}
		})
	}
}

func TestMySQLf(t *testing.T) {
	s := testingSuite{}
	s.setupSuite()
	defer func() {
		if err := s.tearDownSuite(); err != nil {
			t.Fatalf("Failed to tear down suite: %s", err)
		}
	}()

	s.Run(t)
}

func (s *testingSuite) TestInsert(t *testing.T) {
	// prepare mock data
	mockAuthor, err := s.authorF.Build(mockCTX).Insert()
	if err != nil {
		t.Fatalf("Failed to insert author: %s", err)
	}

	// prepare expected data
	stmt := "SELECT * FROM authors WHERE id = ?"
	row := s.db.QueryRow(stmt, mockAuthor.ID)
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

func (s *testingSuite) TestInsertList(t *testing.T) {
	// prepare mock data
	mockAuthors, err := s.authorF.BuildList(mockCTX, 3).Insert()
	if err != nil {
		t.Fatalf("Failed to insert books: %s", err)
	}

	// prepare expected data
	stmt := "SELECT * FROM authors WHERE id IN (?, ?, ?)"
	rows, err := s.db.Query(stmt, mockAuthors[0].ID, mockAuthors[1].ID, mockAuthors[2].ID)
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

func (s *testingSuite) TestWithOne(t *testing.T) {
	// prepare mock data
	mockAuthor := Author{}
	mockGenre := "Science"
	ow := Book{Genre: &mockGenre} // set correct enum value
	mockBook, err := s.bookF.Build(mockCTX).Overwrite(ow).WithOne(&mockAuthor).Insert()
	if err != nil {
		t.Fatalf("Failed to insert: %s", err)
	}

	// prepare expected data
	bookStmt := "SELECT * FROM books WHERE author_id = ?"
	bookRow := s.db.QueryRow(bookStmt, mockBook.AuthorID)
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
	authorRow := s.db.QueryRow(authorStmt, mockBook.AuthorID)
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

func (s *testingSuite) TestWithMany(t *testing.T) {
	// prepare mock data
	mockAnyAuthors := make([]interface{}, 3)
	for i := 0; i < 3; i++ {
		mockAnyAuthors[i] = &Author{}
	}
	mockGenre := "Science"
	ow := Book{Genre: &mockGenre} // set correct enum value
	mockBooks, err := s.bookF.BuildList(mockCTX, 3).Overwrite(ow).WithMany(mockAnyAuthors).Insert()
	if err != nil {
		t.Fatalf("Failed to insert books: %s", err)
	}

	mockAuthors := make([]Author, 3)
	for i := 0; i < 3; i++ {
		mockAuthors[i] = *mockAnyAuthors[i].(*Author)
	}

	// prepare expected data
	bookStmt := "SELECT * FROM books WHERE author_id = ?"
	authorStmt := "SELECT * FROM authors WHERE id = ?"

	// loop through each data to check association connection
	for i := 0; i < 3; i++ {
		bookRow := s.db.QueryRow(bookStmt, mockBooks[i].AuthorID)
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

		authorRow := s.db.QueryRow(authorStmt, mockBooks[i].AuthorID)
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

func (s *testingSuite) TestListWithOne(t *testing.T) {
	// prepare mock data
	mockAuthor := Author{}
	mockGenre := "Science"
	ow := Book{Genre: &mockGenre} // set correct enum value
	mockBooks, err := s.bookF.BuildList(mockCTX, 3).Overwrite(ow).WithOne(&mockAuthor).Insert()
	if err != nil {
		t.Fatalf("Failed to insert books: %s", err)
	}

	// prepare expected data
	bookStmt := "SELECT * FROM books WHERE author_id = ?"
	bookRows, err := s.db.Query(bookStmt, mockBooks[0].AuthorID)
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
	authorRow := s.db.QueryRow(authorStmt, mockBooks[0].AuthorID)
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
