package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Parcel struct {
	Number    int
	Client    int
	Status    stringAddress   string
	CreatedAt string
}

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	result, err := s.db.Exec(
		"INSERT INTO parcels (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	var p Parcel
	err := s.db.QueryRow(
		"SELECT number, client, status, address, created_at FROM parcels WHERE number = ?",
		number,
	).Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(
		"SELECT number, client, status, address, created_at FROM parcels WHERE client = ?",
		client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parcels []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, err
		}
		parcels = append(parcels, p)
	}
	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec("UPDATE parcels SET status = ? WHERE number = ?", status, number)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	var status string
	err := s.db.QueryRow("SELECT status FROM parcels WHERE number = ?", number).Scan(&status)
	if err != nil {
		return err}
	if status != "registered" {
		return fmt.Errorf("address can only be changed if status is 'registered'")
	}
	_, err = s.db.Exec("UPDATE parcels SET address = ? WHERE number = ?", address, number)
	return err
}

func (s ParcelStore) Delete(number int) error {
	var status string
	err := s.db.QueryRow("SELECT status FROM parcels WHERE number = ?", number).Scan(&status)
	if err != nil {
		return err}
	if status != "registered" {
		return fmt.Errorf("parcel can only be deleted if status is 'registered'")
	}
	_, err = s.db.Exec("DELETE FROM parcels WHERE number = ?", number)
	return err
}

func main() {
	// Подключение к базе данных SQLite
	db, err := sql.Open("sqlite3", "./parcels.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создание таблицы parcels, если она не существует
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS parcels (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT)`)
	if err != nil {
		log.Fatal(err)
	}

	// Создание нового экземпляра ParcelStore
	store := NewParcelStore(db)

	// Пример использования
	now := time.Now().Format(time.RFC3339)
	parcel := Parcel{
		Client:1,
		Status:    "registered",
		Address:   "123 Main St",
		CreatedAt: now,
	}

	// Добавление посылки
	id, err := store.Add(parcel)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Added parcel with ID: %d\n", id)

	// Получение посылки
	p, err := store.Get(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Retrieved parcel: %+v\n", p)

	// Обновление статуса посылки
	err = store.SetStatus(id, "shipped")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated parcel status to 'shipped'")

	// Попытка изменения адреса (должна завершиться ошибкой)
	err = store.SetAddress(id, "456 Elm St")
	if err != nil {
		fmt.Println("Error updating address:", err)
	}

	// Удаление посылки (должна завершиться ошибкой)
	err = store.Delete(id)
	if err != nil {
		fmt.Println("Error deleting parcel:", err)
	}
}