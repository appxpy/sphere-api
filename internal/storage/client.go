package storage

import (
	"sync"

	"github.com/appxpy/sphere-api/internal/models"
	"github.com/appxpy/sphere-api/internal/util"
	"github.com/dhconnelly/rtreego"
	"github.com/gorilla/websocket"
)

type ClientRepository struct {
	clients                 map[string]*models.ClientInfo
	connections             map[*websocket.Conn]string
	whoReferenceMeAsNearest map[string]map[string]struct{}

	rtree *rtreego.Rtree

	mu sync.RWMutex
}

func NewClientRepository() *ClientRepository {
	return &ClientRepository{
		clients:                 make(map[string]*models.ClientInfo),
		connections:             make(map[*websocket.Conn]string),
		rtree:                   rtreego.NewTree(3, 25, 50), // Инициализируем R-Tree
		whoReferenceMeAsNearest: make(map[string]map[string]struct{}),
	}
}

func (r *ClientRepository) AddClient(client *models.ClientInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[client.ID] = client
	r.connections[client.Connection] = client.ID
	r.whoReferenceMeAsNearest[client.ID] = make(map[string]struct{})

	// Добавляем клиента в R-Tree, если у него есть позиция
	if client.Position != nil {
		r.rtree.Insert(client)
	}
}

func (r *ClientRepository) RemoveClient(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	client, ok := r.clients[id]
	if !ok {
		return
	}

	// Удаляем из R-Tree, если у клиента есть позиция
	if client.Position != nil {
		r.rtree.Delete(client)
	}

	delete(r.clients, id)
	delete(r.connections, client.Connection)
}

func (r *ClientRepository) UpdateClientPosition(id string, position *models.Position) {
	r.mu.Lock()
	defer r.mu.Unlock()
	client, ok := r.clients[id]
	if !ok {
		// Обработка ошибки
		return
	}

	// Удаляем из R-Tree, если позиция существовала
	if client.Position != nil {
		r.rtree.Delete(client)
	}

	// Обновляем позицию
	client.Position = position

	// Вставляем в R-Tree с новой позицией
	if client.Position != nil {
		r.rtree.Insert(client)
	}
}

func (r *ClientRepository) GetClient(id string) (*models.ClientInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	client, exists := r.clients[id]
	return client, exists
}

func (r *ClientRepository) GetClientIDByConnection(connection *websocket.Conn) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.connections[connection]

	return id, exists
}

func (r *ClientRepository) GetAllClients() []*models.ClientInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*models.ClientInfo, 0, len(r.clients))
	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients
}

func (r *ClientRepository) FindNearestClient(clientID string) (*models.ClientInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, exist := r.clients[clientID]
	if !exist {
		return nil, util.ErrClientNotFound
	}

	if client.Position == nil {
		return nil, util.ErrNoPositionProvided
	}

	p := rtreego.Point{client.Position.X, client.Position.Y, client.Position.Z}

	// Получаем ближайших соседей (включая самого клиента)
	results := r.rtree.NearestNeighbors(2, p)

	oldNearest := client.Position.ClosestClientID

	var nearest *models.ClientInfo
	for _, obj := range results {
		otherClient, ok := obj.(*models.ClientInfo)
		if !ok {
			continue
		}
		if otherClient.ID == clientID {
			continue // Пропускаем самого себя
		}
		nearest = otherClient
		break
	}

	if nearest == nil {
		return nil, util.ErrNoClientsAvailable
	}

	if _, ok := r.whoReferenceMeAsNearest[oldNearest]; ok {
		delete(r.whoReferenceMeAsNearest[oldNearest], clientID)
	}

	r.whoReferenceMeAsNearest[nearest.ID][clientID] = struct{}{}

	return nearest, nil
}

func (r *ClientRepository) WhoReferenceMeAsNearest(id string) []string {
	idsSet, ok := r.whoReferenceMeAsNearest[id]
	if !ok {
		return []string{}
	}

	ids := make([]string, 0, len(idsSet))
	for ref, _ := range idsSet {
		ids = append(ids, ref)
	}

	return ids
}

func (r *ClientRepository) DeleteClientFromNearestReferences(id string) {
	delete(r.whoReferenceMeAsNearest, id)
}

func (r *ClientRepository) HeDoesNotReferenceMeAsNearestAnymore(me, him string) {
	delete(r.whoReferenceMeAsNearest[me], him)
}

func (r *ClientRepository) UpdateClientWindowSettings(id string, settings *models.WindowSettings) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[id].WindowSettings = settings
}
