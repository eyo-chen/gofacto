package gormf

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

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
	BornTime            datatypes.Time
	BornTime1           *datatypes.Time
}

type Book struct {
	ID               int64
	AuthorID         int64 `gofacto:"Author,authors"`
	Title            string
	ISBN             *string
	PublicationDate  datatypes.Date
	PublicationDate1 *datatypes.Date
	Genre            *string
	Price            *float64
	PageCount        *int32
	Description      *string
	InStock          bool
	CoverImage       []byte
	Data             datatypes.JSON
	Data1            *datatypes.JSON
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type testingSuite struct {
	db      *gorm.DB
	authorF *gofacto.Factory[Author]
	bookF   *gofacto.Factory[Book]
}

func (s *testingSuite) setupSuite() {
	// start MySQL Docker container
	port := docker.RunDocker(docker.ImageMySQL)
	dsn := fmt.Sprintf("root:root@(localhost:%s)/mysql?parseTime=true", port)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("gorm.Open failed: %s", err)
	}

	// set the database connection
	s.db = db

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
		if err := db.Exec(query).Error; err != nil {
			log.Fatalf("Failed to execute query: %s, error: %s", query, err)
		}
	}

	// set up gofacto factories
	s.authorF = gofacto.New(Author{}).SetConfig(gofacto.Config[Author]{
		DB: NewConfig(db),
	})
	s.bookF = gofacto.New(Book{}).SetConfig(gofacto.Config[Book]{
		DB: NewConfig(db),
	})
}

func (s *testingSuite) tearDownSuite() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		log.Fatalf("Failed to get DB: %s", err)
	}

	sqlDB.Close()
	docker.PurgeDocker()

	return nil
}

func (s *testingSuite) tearDownTest() error {
	if err := s.db.Exec("DELETE FROM authors").Error; err != nil {
		return err
	}

	if err := s.db.Exec("DELETE FROM books").Error; err != nil {
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

func TestGormf(t *testing.T) {
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
	var author Author
	if err := s.db.First(&author, mockAuthor.ID).Error; err != nil {
		t.Fatalf("Failed to find author: %s", err)
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
		t.Fatalf("Failed to insert authors: %s", err)
	}

	// collect the IDs of the inserted mock authors
	ids := make([]int64, len(mockAuthors))
	for i, author := range mockAuthors {
		ids[i] = author.ID
	}

	// prepare expected data
	var authors []Author
	if err := s.db.Where("id IN ?", ids).Find(&authors).Error; err != nil {
		t.Fatalf("Failed to find authors: %s", err)
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
	var book Book
	if err := s.db.Where("author_id = ?", mockBook.AuthorID).First(&book).Error; err != nil {
		t.Fatalf("Failed to find book: %s", err)
	}

	var author Author
	if err := s.db.First(&author, mockBook.AuthorID).Error; err != nil {
		t.Fatalf("Failed to find author: %s", err)
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

	// loop through each data to check association connection
	for i := 0; i < 3; i++ {
		// prepare expected data
		var book Book
		if err := s.db.Where("author_id = ?", mockBooks[i].AuthorID).First(&book).Error; err != nil {
			t.Fatalf("Failed to find book: %s", err)
		}

		var author Author
		if err := s.db.First(&author, mockBooks[i].AuthorID).Error; err != nil {
			t.Fatalf("Failed to find author: %s", err)
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
	var books []Book
	if err := s.db.Where("author_id = ?", mockBooks[0].AuthorID).Find(&books).Error; err != nil {
		t.Fatalf("Failed to find books: %s", err)
	}

	var author Author
	if err := s.db.First(&author, mockBooks[0].AuthorID).Error; err != nil {
		t.Fatalf("Failed to find author: %s", err)
	}

	// assertion
	if err := testutils.CompareVal(mockBooks, books, "PublicationDate", "CreatedAt", "UpdatedAt"); err != nil {
		t.Fatalf("Inserted books are not the same as the mock books: %s", err)
	}

	if err := testutils.CompareVal(mockAuthor, author, "BirthDate", "LastPublicationTime"); err != nil {
		t.Fatalf("Inserted author is not the same as the mock author: %s", err)
	}
}
