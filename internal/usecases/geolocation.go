package usecases

import (
	"slices"

	"github.com/appxpy/sphere-api/internal/logging"
	"github.com/appxpy/sphere-api/internal/models"
	"github.com/appxpy/sphere-api/internal/storage"
	"github.com/tidwall/geodesic"
)

type GeolocationUsecase struct {
	repo *storage.ClientRepository
}

func NewGeolocationUsecase(repo *storage.ClientRepository) *GeolocationUsecase {
	return &GeolocationUsecase{repo: repo}
}

//func (u *GeolocationUsecase) UpdateHeading(clientID string, heading float64) {
//	if 0.0 > heading || heading > 360.0 {
//		logging.ErrorLogger.Printf("Client %s tried to update heading, which is smaller than 0 or bigger than 360 - %f", clientID, heading)
//		return
//	}
//
//	// Проверяем существует ли клиент
//	var client *models.ClientInfo
//	var exists bool
//
//	if client, exists = u.repo.GetClient(clientID); !exists {
//		logging.ErrorLogger.Printf("Client %s tried to update heading, but it does not exist anymore.", clientID)
//		return
//	}
//
//	if !client.HasPosition() {
//		client.Position = &models.Position{}
//	}
//
//	client.Position.Heading = heading
//}

func (u *GeolocationUsecase) UpdateRelatedClients(clientID string) []string {
	notify := make([]string, 0)

	// Проверяем всех клиентов которые считают нас ближайшими, и проверяем кто стал ближайшим для них
	for _, referencingID := range u.repo.WhoReferenceMeAsNearest(clientID) {
		logging.InfoLogger.Printf("Client %s references me as nearest", referencingID)
		referencingClient, exists := u.repo.GetClient(referencingID)
		if !exists {
			u.repo.HeDoesNotReferenceMeAsNearestAnymore(clientID, referencingID)
			continue
		}

		nearestForRef, _ := u.GetClosestClient(referencingID)
		if nearestForRef == nil {
			referencingClient.Position.ClosestClientID = ""
			referencingClient.Position.Distance = 0
			referencingClient.Position.Azimuth = 0

			notify = append(notify, referencingID)
			continue
		}

		// Вычисляем расстояние и азимут от клиента который на нас ссылается до ближайшего до него
		distance, azimuth, _ := calculateAzimuthAndDistanceBetweenPositions(referencingClient, nearestForRef)
		if referencingClient.Position.ClosestClientID != nearestForRef.ID {
			notify = append(notify, referencingID)
		}

		referencingClient.Position.ClosestClientID = nearestForRef.ID
		referencingClient.Position.Distance = distance
		referencingClient.Position.Azimuth = azimuth
	}

	return notify
}

func (u *GeolocationUsecase) DeleteClientFromNearestReferences(clientID string) {
	u.repo.DeleteClientFromNearestReferences(clientID)
}

func (u *GeolocationUsecase) UpdatePosition(clientID string, lat float64, lon float64) (notify []string) {
	// Проверяем существует ли клиент
	var client *models.ClientInfo
	var exists bool
	notify = make([]string, 0)

	if client, exists = u.repo.GetClient(clientID); !exists {
		logging.ErrorLogger.Printf("Client %s tried to update position, but it does not exist anymore.", clientID)
		return
	}

	if !client.HasPosition() {
		client.Position = &models.Position{}
	}

	// Обновляем X, Y, Z координаты позиции и позицию клиента в репозитории (также обновляет R-Tree) если его геопозиция изменилась
	if client.Position.Latitude != lat || client.Position.Longitude != lon {
		client.Position.Latitude = lat
		client.Position.Longitude = lon

		client.Position.UpdateXYZ()

		u.repo.UpdateClientPosition(clientID, client.Position)
	}

	// Находим нового ближайшего клиента к обновленному клиенту
	nearest, err := u.GetClosestClient(clientID)

	if err != nil {
		// Ближайший клиент не найден
		client.Position.ClosestClientID = ""
		return
	}

	// Вычисляем расстояние и азимут от клиента до его ближайшего клиента
	distanceToNearest, azimuthToNearest, _ := calculateAzimuthAndDistanceBetweenPositions(client, nearest)

	// Обновляем информацию о ближайшем клиенте для текущего клиента
	client.Position.Azimuth = azimuthToNearest
	client.Position.Distance = distanceToNearest
	client.Position.ClosestClientID = nearest.ID
	notify = append(notify, clientID)

	nearestForNearest, _ := u.GetClosestClient(nearest.ID)

	// Вычисляем расстояние и азимут от ближайшего клиента до его ближайшего клиента
	distance, azimuth, _ := calculateAzimuthAndDistanceBetweenPositions(nearest, nearestForNearest)

	nearest.Position.ClosestClientID = nearestForNearest.ID
	nearest.Position.Distance = distance
	nearest.Position.Azimuth = azimuth

	notify = append(notify, u.UpdateRelatedClients(clientID)...)

	if !slices.Contains(notify, nearest.ID) {
		notify = append(notify, nearest.ID)
	}

	return notify
}

func (u *GeolocationUsecase) GetClosestClient(clientID string) (*models.ClientInfo, error) {
	return u.repo.FindNearestClient(clientID)
}

// Пересчитывает азимут и расстояние между двумя клиентами
func calculateAzimuthAndDistanceBetweenPositions(clientA *models.ClientInfo, clientB *models.ClientInfo) (distance, azimuthAtoB, azimuthBtoA float64) {
	geodesic.WGS84.Inverse(clientA.Position.Latitude, clientA.Position.Longitude,
		clientB.Position.Latitude, clientB.Position.Longitude, &distance, &azimuthAtoB, &azimuthBtoA)
	if azimuthAtoB < 0 {
		azimuthAtoB += 360
	}

	if azimuthBtoA < 0 {
		azimuthBtoA += 360
	}

	return
}
