package mongof

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eyo-chen/gofacto"
	"github.com/eyo-chen/gofacto/internal/docker"
	"github.com/eyo-chen/gofacto/internal/testutils"
)

var (
	mockCTX = context.Background()
)

type Person struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty"`
	Name         string               `bson:"name"`
	Age          int                  `bson:"age"`
	Salary       int64                `bson:"salary"`
	Rating       float64              `bson:"rating"`
	IsActive     bool                 `bson:"is_active"`
	DateOfBirth  time.Time            `bson:"date_of_birth"`
	ProfileImage []byte               `bson:"profile_image"`
	Tags         []string             `bson:"tags"`
	Address      Address              `bson:"address"`
	Attributes   map[string]int       `bson:"attributes"`
	Balance      primitive.Decimal128 `bson:"balance"`
	LastLogin    primitive.Timestamp  `bson:"last_login"`
	Description  interface{}          `bson:"description,omitempty"`
	// Pattern      primitive.Regex      `bson:"pattern"`
	Notes  *string `bson:"notes,omitempty"`
	Orders []Order `bson:"orders"`
}

type Address struct {
	Street  string `bson:"street"`
	City    string `bson:"city"`
	State   string `bson:"state"`
	ZipCode string `bson:"zip_code"`
}

type Order struct {
	Product     string  `bson:"product"`
	Quantity    int     `bson:"quantity"`
	Price       float64 `bson:"price"`
	IsDelivered bool    `bson:"is_delivered"`
}

type testingSuite struct {
	db *mongo.Database
	f  *gofacto.Factory[Person]
}

func (s *testingSuite) setupSuite() {
	// Start MySQL Docker container
	port := docker.RunDocker(docker.ImageMongo)
	client, err := mongo.Connect(mockCTX, options.Client().ApplyURI(fmt.Sprintf("mongodb://localhost:%s", port)))
	if err != nil {
		log.Fatalf("mongo.Connect failed: %s", err)
	}
	s.db = client.Database("mongo")

	s.f = gofacto.New(Person{}).SetConfig(gofacto.Config[Person]{
		DB: NewConfig(s.db),
	})
}

func (s *testingSuite) tearDownSuite() {
	if err := s.db.Client().Disconnect(mockCTX); err != nil {
		log.Fatalf("s.db.Client().Disconnect failed: %s", err)
	}

	docker.PurgeDocker()
}

func (s *testingSuite) tearDownTest() {
	if err := s.db.Drop(mockCTX); err != nil {
		log.Fatalf("s.db.Drop failed: %s", err)
	}
}

func (s *testingSuite) Run(t *testing.T) {
	s.setupSuite()
	defer s.tearDownSuite()

	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"TestInsert", s.TestInsert},
		{"TestInsertList", s.TestInsertList},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.fn(t)
			s.tearDownTest()
		})
	}
}

func TestMongof(t *testing.T) {
	s := testingSuite{}
	s.Run(t)
}

func (s *testingSuite) TestInsert(t *testing.T) {
	// prepare mock data
	mockPerson, err := s.f.Build(mockCTX).Insert()
	if err != nil {
		t.Fatalf("Failed to insert person: %s", err)
	}

	// prepare expected data
	var expResp Person
	if err := s.db.Collection("persons").FindOne(mockCTX, bson.M{"_id": mockPerson.ID}).Decode(&expResp); err != nil {
		t.Fatalf("Failed to find person: %s", err)
	}

	// assertion
	if err := testutils.CompareVal(mockPerson, expResp, "DateOfBirth"); err != nil {
		t.Fatalf("testutils.CompareVal failed: %s", err)
	}
}

func (s *testingSuite) TestInsertList(t *testing.T) {
	// prepare mock data
	mockPersons, err := s.f.BuildList(mockCTX, 3).Insert()
	if err != nil {
		t.Fatalf("Failed to insert persons: %s", err)
	}

	// prepare expected data
	var expResp []Person
	cursor, err := s.db.Collection("persons").Find(mockCTX, bson.M{"_id": bson.M{"$in": []primitive.ObjectID{mockPersons[0].ID, mockPersons[1].ID, mockPersons[2].ID}}})
	if err != nil {
		t.Fatalf("Failed to find persons: %s", err)
	}
	if err := cursor.All(mockCTX, &expResp); err != nil {
		t.Fatalf("Failed to decode persons: %s", err)
	}

	// assertion
	if err := testutils.CompareVal(mockPersons, expResp, "DateOfBirth"); err != nil {
		t.Fatalf("testutils.CompareVal failed: %s", err)
	}
}
