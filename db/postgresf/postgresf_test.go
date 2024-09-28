package postgresf

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
	_ "github.com/lib/pq"
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
	AuthorID        int64 `gofacto:"foreignKey,struct:Author"`
	CategoryID      int64 `gofacto:"foreignKey,struct:Category,table:categories"`
	SubCategoryID   int64 `gofacto:"foreignKey,struct:SubCategory,table:sub_categories"`
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

type Category struct {
	ID          int64
	Name        string
	AuthorID    int64 `gofacto:"foreignKey,struct:Author"`
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SubCategory struct {
	ID          int64
	CategoryID  int64 `gofacto:"foreignKey,struct:Category,table:categories"`
	AuthorID    int64 `gofacto:"foreignKey,struct:Author"`
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type testingSuite struct {
	db        *sql.DB
	authorF   *gofacto.Factory[Author]
	bookF     *gofacto.Factory[Book]
	categoryF *gofacto.Factory[Category]
}

func bookBlueprint(i int) Book {
	mockGenre := "Fiction"
	return Book{Genre: &mockGenre}
}

func (s *testingSuite) setupSuite() {
	// Start PostgreSQL Docker container
	port := docker.RunDocker(docker.ImagePostgres)
	dba, err := sql.Open("postgres", fmt.Sprintf("postgres://postgres:postgres@localhost:%s/postgres?sslmode=disable", port))
	if err != nil {
		log.Fatalf("sql.Open failed: %s", err)
	}

	// Set the database connection
	s.db = dba

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

	// Set up gofacto factories
	s.authorF = gofacto.New(Author{}).WithDB(NewConfig(s.db))
	s.bookF = gofacto.New(Book{}).WithDB(NewConfig(s.db)).WithBlueprint(bookBlueprint)
	s.categoryF = gofacto.New(Category{}).WithDB(NewConfig(s.db)).WithStorageName("categories")
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

	if _, err := s.db.Exec("DELETE FROM categories"); err != nil {
		return err
	}

	if _, err := s.db.Exec("DELETE FROM sub_categories"); err != nil {
		return err
	}

	s.authorF.Reset()
	s.bookF.Reset()
	s.categoryF.Reset()

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

func TestPostgresf(t *testing.T) {
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
	author, err := findAuthor(s.db, "SELECT * FROM authors WHERE id = $1", mockAuthor.ID)
	if err != nil {
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

	// prepare expected data
	authors, err := findAuthors(s.db, "SELECT * FROM authors WHERE id IN ($1, $2, $3)", mockAuthors[0].ID, mockAuthors[1].ID, mockAuthors[2].ID)
	if err != nil {
		t.Fatalf("Failed to find authors: %s", err)
	}

	// assertion
	if err := testutils.CompareVal(mockAuthors, authors, "BirthDate", "LastPublicationTime"); err != nil {
		t.Fatalf("Inserted authors are not the same as the mock authors: %s", err)
	}
}

func (s *testingSuite) TestWithOne(t *testing.T) {
	for _, fn := range map[string]func(*testingSuite, *testing.T){
		"when on builder, insert with association correctly":                              withOne_OnBuilder,
		"when on builder list, insert with association correctly":                         withOne_OnBuilderList,
		"when on builder with multi level association, insert with association correctly": withOne_OnBuilderWithMultiLevelAssociation,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(s, t)
			if err := s.tearDownTest(); err != nil {
				t.Fatalf("Failed to tear down test: %s", err)
			}
		})
	}
}

func withOne_OnBuilder(s *testingSuite, t *testing.T) {
	// prepare mock data
	mockAuthor := Author{}
	mockCategory, err := s.categoryF.Build(mockCTX).WithOne(&mockAuthor).Insert()
	if err != nil {
		t.Fatalf("Failed to insert: %s", err)
	}

	// prepare expected data
	category, err := findCategory(s.db, "SELECT * FROM categories WHERE id = $1", mockCategory.ID)
	if err != nil {
		t.Fatalf("Failed to find category: %s", err)
	}

	author, err := findAuthor(s.db, "SELECT * FROM authors WHERE id = $1", mockCategory.AuthorID)
	if err != nil {
		t.Fatalf("Failed to find author: %s", err)
	}

	// assertion
	if err := testutils.CompareVal(mockCategory, category, "CreatedAt", "UpdatedAt"); err != nil {
		t.Fatalf("Inserted category is not the same as the mock category: %s", err)
	}

	if err := testutils.CompareVal(mockAuthor, author, "BirthDate", "LastPublicationTime"); err != nil {
		t.Fatalf("Inserted author is not the same as the mock author: %s", err)
	}

	// check if the association is correctly set
	if mockCategory.AuthorID != mockAuthor.ID {
		t.Fatalf("Inserted category author id is not the same as the mock author id: %d", mockCategory.AuthorID)
	}
}

func withOne_OnBuilderList(s *testingSuite, t *testing.T) {
	// prepare mock data
	mockRating := 4.5
	mockAuthor := Author{Rating: &mockRating}
	mockCategories, err := s.categoryF.
		BuildList(mockCTX, 3).
		WithOne(&mockAuthor).
		Insert()
	if err != nil {
		t.Fatalf("Failed to insert categories: %s", err)
	}

	// prepare expected data
	categories, err := findCategories(s.db, "SELECT * FROM categories WHERE author_id = $1", mockCategories[0].AuthorID)
	if err != nil {
		t.Fatalf("Failed to find categories: %s", err)
	}

	author, err := findAuthor(s.db, "SELECT * FROM authors WHERE id = $1", mockCategories[0].AuthorID)
	if err != nil {
		t.Fatalf("Failed to find author: %s", err)
	}

	// assertion
	if err := testutils.CompareVal(mockCategories, categories, "CreatedAt", "UpdatedAt"); err != nil {
		t.Fatalf("Inserted categories are not the same as the mock categories: %s", err)
	}

	if err := testutils.CompareVal(mockAuthor, author, "BirthDate", "LastPublicationTime"); err != nil {
		t.Fatalf("Inserted author is not the same as the mock author: %s", err)
	}

	// check if the association is correctly set
	for _, c := range categories {
		if c.AuthorID != mockAuthor.ID {
			t.Fatalf("Inserted category author id is not the same as the mock author id: %d", c.AuthorID)
		}
	}
}

func withOne_OnBuilderWithMultiLevelAssociation(s *testingSuite, t *testing.T) {
	// prepare mock data
	mockAuthor := Author{}
	mockCategory := Category{}
	mockSubCategory := SubCategory{}

	mockBook, err := s.bookF.
		Build(mockCTX).
		WithOne(&mockAuthor, &mockCategory, &mockSubCategory).
		Insert()
	if err != nil {
		t.Fatalf("Failed to insert book: %s", err)
	}

	// prepare expected data
	author, err := findAuthor(s.db, "SELECT * FROM authors WHERE id = $1", mockBook.AuthorID)
	if err != nil {
		t.Fatalf("Failed to find author: %s", err)
	}

	category, err := findCategory(s.db, "SELECT * FROM categories WHERE author_id = $1", mockBook.AuthorID)
	if err != nil {
		t.Fatalf("Failed to find category: %s", err)
	}

	subCategory, err := findSubCategory(s.db, "SELECT * FROM sub_categories WHERE author_id = $1", mockBook.AuthorID)
	if err != nil {
		t.Fatalf("Failed to find sub category: %s", err)
	}

	book, err := findBook(s.db, "SELECT * FROM books WHERE id = $1", mockBook.ID)
	if err != nil {
		t.Fatalf("Failed to find book: %s", err)
	}

	// assertion
	if err := testutils.CompareVal(mockAuthor, author, "BirthDate", "LastPublicationTime"); err != nil {
		t.Fatalf("Inserted author is not the same as the mock author: %s", err)
	}

	if err := testutils.CompareVal(mockCategory, category, "CreatedAt", "UpdatedAt"); err != nil {
		t.Fatalf("Inserted category is not the same as the mock category: %s", err)
	}

	if err := testutils.CompareVal(mockSubCategory, subCategory, "CreatedAt", "UpdatedAt"); err != nil {
		t.Fatalf("Inserted sub category is not the same as the mock sub category: %s", err)
	}

	// trim the value of CHAR(13)
	// when selecting the value of CHAR(13) from postgres, it will include the padding space
	trimStr := strings.TrimSpace(*book.ISBN)
	book.ISBN = &trimStr
	if err := testutils.CompareVal(mockBook, book, "CreatedAt", "UpdatedAt", "PublicationDate"); err != nil {
		t.Fatalf("Inserted book is not the same as the mock book: %s", err)
	}

	// check if the association is correctly set
	// check book association
	if mockBook.AuthorID != mockAuthor.ID {
		t.Fatalf("Inserted book author id is not the same as the mock author id: %d", mockBook.AuthorID)
	}
	if mockBook.CategoryID != mockCategory.ID {
		t.Fatalf("Inserted book category id is not the same as the mock category id: %d", mockBook.CategoryID)
	}
	if mockBook.SubCategoryID != mockSubCategory.ID {
		t.Fatalf("Inserted book sub category id is not the same as the mock sub category id: %d", mockBook.SubCategoryID)
	}

	// check category association
	if mockCategory.AuthorID != mockAuthor.ID {
		t.Fatalf("Inserted category author id is not the same as the mock author id: %d", mockCategory.AuthorID)
	}

	// check sub category association
	if mockSubCategory.AuthorID != mockAuthor.ID {
		t.Fatalf("Inserted sub category author id is not the same as the mock author id: %d", mockSubCategory.AuthorID)
	}
	if mockSubCategory.CategoryID != mockCategory.ID {
		t.Fatalf("Inserted sub category category id is not the same as the mock category id: %d", mockSubCategory.CategoryID)
	}
}

func (s *testingSuite) TestWithMany(t *testing.T) {
	for _, fn := range map[string]func(*testingSuite, *testing.T){
		"when association and factory value same amount, insert with association correctly":                  withMany_AssocAndFactoryValSameAmount,
		"when association and factory value different amount, insert with association correctly":             withMany_AssocAndFactoryValDiffAmount,
		"when multi level association and factory value same amount, insert with association correctly":      withMany_MultiLevelSameAmount,
		"when multi level association and factory value different amount, insert with association correctly": withMany_MultiLevelDiffAmount,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(s, t)
			if err := s.tearDownTest(); err != nil {
				t.Fatalf("Failed to tear down test: %s", err)
			}
		})
	}
}

func withMany_AssocAndFactoryValSameAmount(s *testingSuite, t *testing.T) {
	// prepare mock data
	mockAuthorAmount := 3
	mockAnyAuthors := make([]interface{}, mockAuthorAmount)
	mockRating := 2.0
	for i := 0; i < mockAuthorAmount; i++ {
		mockAnyAuthors[i] = &Author{Rating: &mockRating}
	}
	mockCategories, err := s.categoryF.BuildList(mockCTX, 3).WithMany(mockAnyAuthors).Insert()
	if err != nil {
		t.Fatalf("Failed to insert categories: %s", err)
	}

	mockAuthors := make([]Author, mockAuthorAmount)
	for i := 0; i < mockAuthorAmount; i++ {
		mockAuthors[i] = *mockAnyAuthors[i].(*Author)
	}

	// prepare expected data
	// loop through each data to check association connection
	for i := 0; i < mockAuthorAmount; i++ {
		category, err := findCategory(s.db, "SELECT * FROM categories WHERE id = $1", mockCategories[i].ID)
		if err != nil {
			t.Fatalf("Failed to find category: %s", err)
		}

		author, err := findAuthor(s.db, "SELECT * FROM authors WHERE id = $1", mockCategories[i].AuthorID)
		if err != nil {
			t.Fatalf("Failed to find author: %s", err)
		}

		// assertion
		if err := testutils.CompareVal(mockCategories[i], category, "CreatedAt", "UpdatedAt"); err != nil {
			t.Fatalf("Inserted category is not the same as the mock category: %s", err)
		}

		if err := testutils.CompareVal(mockAuthors[i], author, "BirthDate", "LastPublicationTime"); err != nil {
			t.Fatalf("Inserted author is not the same as the mock author: %s", err)
		}

		// check if the association is correctly set
		if mockCategories[i].AuthorID != mockAuthors[i].ID {
			t.Fatalf("Inserted category author id is not the same as the mock author id: %d", mockCategories[i].AuthorID)
		}
	}
}

func withMany_AssocAndFactoryValDiffAmount(s *testingSuite, t *testing.T) {
	// prepare mock data
	mockAuthorAmount := 2
	mockAnyAuthors := make([]interface{}, mockAuthorAmount)
	mockRating := 2.0
	for i := 0; i < mockAuthorAmount; i++ {
		mockAnyAuthors[i] = &Author{Rating: &mockRating}
	}
	mockCategories, err := s.categoryF.BuildList(mockCTX, 3).WithMany(mockAnyAuthors).Insert()
	if err != nil {
		t.Fatalf("Failed to insert categories: %s", err)
	}

	mockAuthors := make([]Author, mockAuthorAmount)
	for i := 0; i < mockAuthorAmount; i++ {
		mockAuthors[i] = *mockAnyAuthors[i].(*Author)
	}

	// prepare expected data
	// loop through each data to check association connection
	for i := 0; i < 3; i++ {
		// because we only have 2 authors, we need to loop back to the last author
		mockAuthorIdx := i
		if mockAuthorIdx >= mockAuthorAmount {
			mockAuthorIdx = mockAuthorAmount - 1
		}

		category, err := findCategory(s.db, "SELECT * FROM categories WHERE id = $1", mockCategories[i].ID)
		if err != nil {
			t.Fatalf("Failed to find category: %s", err)
		}

		author, err := findAuthor(s.db, "SELECT * FROM authors WHERE id = $1", mockCategories[i].AuthorID)
		if err != nil {
			t.Fatalf("Failed to find author: %s", err)
		}

		// assertion
		if err := testutils.CompareVal(mockCategories[i], category, "CreatedAt", "UpdatedAt"); err != nil {
			t.Fatalf("Inserted category is not the same as the mock category: %s", err)
		}

		if err := testutils.CompareVal(mockAuthors[mockAuthorIdx], author, "BirthDate", "LastPublicationTime"); err != nil {
			t.Fatalf("Inserted author is not the same as the mock author: %s", err)
		}

		// check if the association is correctly set
		if mockCategories[i].AuthorID != mockAuthors[mockAuthorIdx].ID {
			t.Fatalf("Inserted category author id is not the same as the mock author id: %d", mockCategories[i].AuthorID)
		}
	}
}

func withMany_MultiLevelSameAmount(s *testingSuite, t *testing.T) {
	// prepare mock data
	mockAuthor := make([]interface{}, 3)
	for i := 0; i < 3; i++ {
		mockAuthor[i] = &Author{}
	}

	mockCategory := make([]interface{}, 3)
	for i := 0; i < 3; i++ {
		mockCategory[i] = &Category{}
	}

	mockSubCategory := make([]interface{}, 3)
	for i := 0; i < 3; i++ {
		mockSubCategory[i] = &SubCategory{}
	}

	mockBooks, err := s.bookF.
		BuildList(mockCTX, 3).
		WithMany(mockAuthor).
		WithMany(mockCategory).
		WithMany(mockSubCategory).
		Insert()
	if err != nil {
		t.Fatalf("Failed to insert books: %s", err)
	}

	mockAuthors := make([]Author, 3)
	for i := 0; i < 3; i++ {
		mockAuthors[i] = *mockAuthor[i].(*Author)
	}

	mockCategories := make([]Category, 3)
	for i := 0; i < 3; i++ {
		mockCategories[i] = *mockCategory[i].(*Category)
	}

	mockSubCategories := make([]SubCategory, 3)
	for i := 0; i < 3; i++ {
		mockSubCategories[i] = *mockSubCategory[i].(*SubCategory)
	}

	for i := 0; i < 3; i++ {
		// prepare expected data
		author, err := findAuthor(s.db, "SELECT * FROM authors WHERE id = $1", mockBooks[i].AuthorID)
		if err != nil {
			t.Fatalf("Failed to find author: %s", err)
		}

		category, err := findCategory(s.db, "SELECT * FROM categories WHERE id = $1", mockBooks[i].CategoryID)
		if err != nil {
			t.Fatalf("Failed to find category: %s", err)
		}

		subCategory, err := findSubCategory(s.db, "SELECT * FROM sub_categories WHERE id = $1", mockBooks[i].SubCategoryID)
		if err != nil {
			t.Fatalf("Failed to find sub category: %s", err)
		}

		// assertion
		if err := testutils.CompareVal(mockAuthors[i], author, "BirthDate", "LastPublicationTime"); err != nil {
			t.Fatalf("Inserted author is not the same as the mock author: %s", err)
		}

		if err := testutils.CompareVal(mockCategories[i], category, "CreatedAt", "UpdatedAt"); err != nil {
			t.Fatalf("Inserted category is not the same as the mock category: %s", err)
		}

		if err := testutils.CompareVal(mockSubCategories[i], subCategory, "CreatedAt", "UpdatedAt"); err != nil {
			t.Fatalf("Inserted sub category is not the same as the mock sub category: %s", err)
		}

		// check if the association is correctly set
		// check book association
		if mockBooks[i].AuthorID != mockAuthors[i].ID {
			t.Fatalf("Inserted book author id is not the same as the mock author id: %d", mockBooks[i].AuthorID)
		}
		if mockBooks[i].CategoryID != mockCategories[i].ID {
			t.Fatalf("Inserted book category id is not the same as the mock category id: %d", mockBooks[i].CategoryID)
		}
		if mockBooks[i].SubCategoryID != mockSubCategories[i].ID {
			t.Fatalf("Inserted book sub category id is not the same as the mock sub category id: %d", mockBooks[i].SubCategoryID)
		}

		// check category association
		if mockCategories[i].AuthorID != mockAuthors[i].ID {
			t.Fatalf("Inserted category author id is not the same as the mock author id: %d", mockCategories[i].AuthorID)
		}

		// check sub category association
		if mockSubCategories[i].AuthorID != mockAuthors[i].ID {
			t.Fatalf("Inserted sub category author id is not the same as the mock author id: %d", mockSubCategories[i].AuthorID)
		}
		if mockSubCategories[i].CategoryID != mockCategories[i].ID {
			t.Fatalf("Inserted sub category category id is not the same as the mock category id: %d", mockSubCategories[i].CategoryID)
		}
	}
}

func withMany_MultiLevelDiffAmount(s *testingSuite, t *testing.T) {
	// prepare mock data
	mockAuthorAmount := 1
	mockAuthor := make([]interface{}, mockAuthorAmount)
	for i := 0; i < mockAuthorAmount; i++ {
		mockAuthor[i] = &Author{}
	}

	mockCategoryAmount := 2
	mockCategory := make([]interface{}, mockCategoryAmount)
	for i := 0; i < mockCategoryAmount; i++ {
		mockCategory[i] = &Category{}
	}

	mockSubCategoryAmount := 3
	mockSubCategory := make([]interface{}, mockSubCategoryAmount)
	for i := 0; i < mockSubCategoryAmount; i++ {
		mockSubCategory[i] = &SubCategory{}
	}

	mockBooks, err := s.bookF.
		BuildList(mockCTX, 3).
		WithMany(mockAuthor).
		WithMany(mockCategory).
		WithMany(mockSubCategory).
		Insert()
	if err != nil {
		t.Fatalf("Failed to insert books: %s", err)
	}

	mockAuthors := make([]Author, mockAuthorAmount)
	for i := 0; i < mockAuthorAmount; i++ {
		mockAuthors[i] = *mockAuthor[i].(*Author)
	}

	mockCategories := make([]Category, mockCategoryAmount)
	for i := 0; i < mockCategoryAmount; i++ {
		mockCategories[i] = *mockCategory[i].(*Category)
	}

	mockSubCategories := make([]SubCategory, mockSubCategoryAmount)
	for i := 0; i < mockSubCategoryAmount; i++ {
		mockSubCategories[i] = *mockSubCategory[i].(*SubCategory)
	}

	for i := 0; i < 3; i++ {
		mockAuthorIdx := i
		if mockAuthorIdx >= mockAuthorAmount {
			mockAuthorIdx = mockAuthorAmount - 1
		}

		mockCategoryIdx := i
		if mockCategoryIdx >= mockCategoryAmount {
			mockCategoryIdx = mockCategoryAmount - 1
		}

		mockSubCategoryIdx := i
		if mockSubCategoryIdx >= mockSubCategoryAmount {
			mockSubCategoryIdx = mockSubCategoryAmount - 1
		}

		// prepare expected data
		author, err := findAuthor(s.db, "SELECT * FROM authors WHERE id = $1", mockBooks[i].AuthorID)
		if err != nil {
			t.Fatalf("Failed to find author: %s", err)
		}

		category, err := findCategory(s.db, "SELECT * FROM categories WHERE id = $1", mockBooks[i].CategoryID)
		if err != nil {
			t.Fatalf("Failed to find category: %s", err)
		}

		subCategory, err := findSubCategory(s.db, "SELECT * FROM sub_categories WHERE id = $1", mockBooks[i].SubCategoryID)
		if err != nil {
			t.Fatalf("Failed to find sub category: %s", err)
		}

		// assertion
		if err := testutils.CompareVal(mockAuthors[mockAuthorIdx], author, "BirthDate", "LastPublicationTime"); err != nil {
			t.Fatalf("Inserted author is not the same as the mock author: %s", err)
		}

		if err := testutils.CompareVal(mockCategories[mockCategoryIdx], category, "CreatedAt", "UpdatedAt"); err != nil {
			t.Fatalf("Inserted category is not the same as the mock category: %s", err)
		}

		if err := testutils.CompareVal(mockSubCategories[mockSubCategoryIdx], subCategory, "CreatedAt", "UpdatedAt"); err != nil {
			t.Fatalf("Inserted sub category is not the same as the mock sub category: %s", err)
		}

		// check if the association is correctly set
		// check book association
		if mockBooks[i].AuthorID != mockAuthors[mockAuthorIdx].ID {
			t.Fatalf("Inserted book author id is not the same as the mock author id: %d", mockBooks[i].AuthorID)
		}
		if mockBooks[i].CategoryID != mockCategories[mockCategoryIdx].ID {
			t.Fatalf("Inserted book category id is not the same as the mock category id: %d", mockBooks[i].CategoryID)
		}
		if mockBooks[i].SubCategoryID != mockSubCategories[mockSubCategoryIdx].ID {
			t.Fatalf("Inserted book sub category id is not the same as the mock sub category id: %d", mockBooks[i].SubCategoryID)
		}

		// check category association
		if mockCategories[mockCategoryIdx].AuthorID != mockAuthors[mockAuthorIdx].ID {
			t.Fatalf("Inserted category author id is not the same as the mock author id: %d", mockCategories[mockCategoryIdx].AuthorID)
		}

		// check sub category association
		if mockSubCategories[mockSubCategoryIdx].AuthorID != mockAuthors[mockAuthorIdx].ID {
			t.Fatalf("Inserted sub category author id is not the same as the mock author id: %d", mockSubCategories[mockSubCategoryIdx].AuthorID)
		}
		if mockSubCategories[mockSubCategoryIdx].CategoryID != mockCategories[mockCategoryIdx].ID {
			t.Fatalf("Inserted sub category category id is not the same as the mock category id: %d", mockSubCategories[i].CategoryID)
		}
	}
}

func findAuthor(db *sql.DB, stmt string, args ...any) (Author, error) {
	row := db.QueryRow(stmt, args...)
	var author Author
	err := row.Scan(
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
	)
	return author, err
}

func findSubCategory(db *sql.DB, stmt string, args ...any) (SubCategory, error) {
	row := db.QueryRow(stmt, args...)
	var subCategory SubCategory
	err := row.Scan(
		&subCategory.ID,
		&subCategory.CategoryID,
		&subCategory.AuthorID,
		&subCategory.Name,
		&subCategory.Description,
		&subCategory.CreatedAt,
		&subCategory.UpdatedAt,
	)
	return subCategory, err
}

func findBook(db *sql.DB, stmt string, args ...any) (Book, error) {
	row := db.QueryRow(stmt, args...)
	var book Book
	err := row.Scan(
		&book.ID,
		&book.AuthorID,
		&book.CategoryID,
		&book.SubCategoryID,
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
	)
	return book, err
}

func findCategory(db *sql.DB, stmt string, args ...any) (Category, error) {
	row := db.QueryRow(stmt, args...)
	var category Category
	err := row.Scan(
		&category.ID,
		&category.AuthorID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	return category, err
}

func findCategories(db *sql.DB, stmt string, args ...any) ([]Category, error) {
	rows, err := db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(
			&category.ID,
			&category.AuthorID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func findAuthors(db *sql.DB, stmt string, args ...any) ([]Author, error) {
	rows, err := db.Query(stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []Author
	for rows.Next() {
		var author Author
		err := rows.Scan(
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
		)
		if err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}

	return authors, nil
}
