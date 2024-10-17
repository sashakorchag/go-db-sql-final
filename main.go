package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

type ParcelStore interface {
	Add(parcel Parcel) (int, error)
	GetByClient(client int) ([]Parcel, error)
	Delete(number int) error
}

type ParcelService struct {
	store ParcelStore
}

func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store}
}

func (s ParcelService) Register(client int, address string) (Parcel, error) {
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: "2023-10-01T10:00:00Z", // Здесь можно использовать текущее время
	}

	id, err := s.store.Add(parcel)
	if err != nil {
		return parcel, fmt.Errorf("failed to add parcel: %w", err)
	}

	parcel.Number = id
	fmt.Printf("Новая послка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt)

	return parcel, nil
}

func (s ParcelService) PrintClientParcels(client int) error {
	parcels, err := s.store.GetByClient(client)
	if err != nil {
		return fmt.Errorf("failed to get parcels for client %d: %w", client, err)
	}

	fmt.Printf("Посылки клиента %d:\n", client)
	for _, parcel := range parcels {
		fmt.Printf("Посылка № %d, статус: %s, адрес: %s\n", parcel.Number, parcel.Status, parcel.Address)
	}
	return nil
}

func (s ParcelService) ChangeAddress(number int, newAddress string) error {
	// Логика изменения адреса (не реализована в данном примере)
	return nil
}

func (s ParcelService) NextStatus(number int) error {
	// Логика изменения статуса (не реализована в данном примере)
	return nil
}

func (s ParcelService) Delete(number int) error {
	// Логика удаления посылки (не реализована в данном примере)
	return nil
}

func main() {
	// Открытие соединения с базой данных
	db, err := sql.Open("sqlite", "parcels.db")
	if err != nil {
		fmt.Println("Ошибка подключения к базе данных:", err)
		return
	}
	defer db.Close()

	// Создание таблицы, если она не существует_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS parcels (
			number INTEGER PRIMARY KEY AUTOINCREMENT,
			client INTEGER,
			status TEXT,
			address TEXT,
			created_at TEXT
		)
	`)
	if err != nil {
		fmt.Println("Ошибка создания таблицы:", err)
		return
	}

	// Создание объекта ParcelStore
	store := NewParcelStore(db)
	service := NewParcelService(store)

	// Пример использования сервиса
	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	p, err := service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Изменение адреса
	newAddress := "Саратов, д. Верхние Зори, ул. Козлова, д. 25"
	err = service.ChangeAddress(p.Number, newAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Изменение статуса
	err = service.NextStatus(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Вывод посылок клиента
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Попытка удаления отправленной посылки
	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Вывод посылок клиента после удаления
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Регистрация новой посылки
	p, err = service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Удаление новой посылки
	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Вывод посылок клиента после удаления
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}
}