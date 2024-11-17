package usecases_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/appxpy/sphere-api/internal/models"
	"github.com/appxpy/sphere-api/internal/storage"
	"github.com/appxpy/sphere-api/internal/usecases"
	"github.com/tidwall/geodesic"
)

// GeolocationUsecaseTestSuite defines the suite structure
type GeolocationUsecaseTestSuite struct {
	suite.Suite
	usecase *usecases.GeolocationUsecase
	repo    *storage.ClientRepository

	// Test variables moved to SetupTest
	client1ID string
	client2ID string
	client3ID string

	pos1 *models.Position
	pos2 *models.Position
	pos3 *models.Position

	client1 *models.ClientInfo
	client2 *models.ClientInfo
	client3 *models.ClientInfo
}

// SetupTest initializes the necessary components and variables before each test
func (t *GeolocationUsecaseTestSuite) SetupTest() {
	// Initialize the repository and usecase
	t.repo = storage.NewClientRepository()
	t.usecase = usecases.NewGeolocationUsecase(t.repo)

	// Initialize test variables
	t.client1ID = "client1"
	t.client2ID = "client2"
	t.client3ID = "client3"

	// Positions for clients
	t.pos1 = &models.Position{Latitude: 55.755820, Longitude: 37.617633} // Moscow
	t.pos2 = &models.Position{Latitude: 50.450514, Longitude: 30.523440} // Kiev
	t.pos3 = &models.Position{Latitude: 52.373036, Longitude: 4.892413}  // Amsterdam

	t.client1 = &models.ClientInfo{
		ID: t.client1ID,
	}
	t.client2 = &models.ClientInfo{
		ID: t.client2ID,
	}
	t.client3 = &models.ClientInfo{
		ID: t.client3ID,
	}
}

// TestClientsUpdatingPositions tests clients updating their positions
func (t *GeolocationUsecaseTestSuite) TestClientsUpdatingPositions() {
	// Add clients
	t.repo.AddClient(t.client1)
	t.repo.AddClient(t.client2)
	t.repo.AddClient(t.client3)

	t.usecase.UpdatePosition(t.client1ID, t.pos1.Latitude, t.pos1.Longitude)
	t.usecase.UpdatePosition(t.client2ID, t.pos2.Latitude, t.pos2.Longitude)
	t.usecase.UpdatePosition(t.client3ID, t.pos3.Latitude, t.pos3.Longitude)

	t.Require().Equal(t.client1.Position.ClosestClientID, t.client2ID, "Moscow should be closest to Kiev than to Amsterdam")
	t.Require().Equal(t.client2.Position.ClosestClientID, t.client1ID, "Kiev should be closest to Moscow than to Amsterdam")
	t.Require().Equal(t.client3.Position.ClosestClientID, t.client2ID, "Amsterdam should be closest to Kiev than to Moscow")

	// Let's simulate client1 (Moscow) moving to a new location (Las Vegas)
	newPos1 := &models.Position{Latitude: 36.1699, Longitude: -115.1398} // Las Vegas
	t.usecase.UpdatePosition(t.client1ID, newPos1.Latitude, newPos1.Longitude)

	client1, _ := t.repo.GetClient(t.client1ID)
	client2, _ := t.repo.GetClient(t.client2ID)
	client3, _ := t.repo.GetClient(t.client3ID)

	t.Require().Equal(client1.Position.ClosestClientID, t.client3ID, "Las Vegas should be closest to Amsterdam than to Kiev")
	t.Require().Equal(client2.Position.ClosestClientID, t.client3ID, "Kiev should be closest to Amsterdam than to Las Vegas")
	t.Require().Equal(client3.Position.ClosestClientID, t.client2ID, "Amsterdam should be closest to Kiev than to Las Vegas")

	t.Require().Equal(client1.Position.Distance, calculateDistance(client1, client3), "Distance between client1 (Las Vegas) and client3 (Amsterdam) calculated incorrectly!")
	t.Require().Equal(client2.Position.Distance, calculateDistance(client2, client3), "Distance between client2 (Kiev) and client3 (Amsterdam) calculated incorrectly!")
	t.Require().Equal(client3.Position.Distance, calculateDistance(client2, client3), "Distance between client3 (Amsterdam) and client2 (Kiev) calculated incorrectly!")

	// Return Las Vegas back to Moscow
	t.usecase.UpdatePosition(t.client1ID, t.pos1.Latitude, t.pos1.Longitude)

	t.Require().Equal(t.client1.Position.ClosestClientID, t.client2ID, "Moscow should be closest to Kiev than to Amsterdam")
	t.Require().Equal(t.client2.Position.ClosestClientID, t.client1ID, "Kiev should be closest to Moscow than to Amsterdam")
	t.Require().Equal(t.client3.Position.ClosestClientID, t.client2ID, "Amsterdam should be closest to Kiev than to Moscow")
}

// calculateDistance is a helper function to compute the geodesic distance between two points
func calculateDistance(from, to *models.ClientInfo) float64 {
	var distance float64
	geodesic.WGS84.Inverse(
		from.Position.Latitude,
		from.Position.Longitude,
		to.Position.Latitude,
		to.Position.Longitude,
		&distance,
		nil,
		nil,
	)
	return distance
}

// TestGeolocationUsecaseTestSuite runs the test suite
func TestGeolocationUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(GeolocationUsecaseTestSuite))
}
