package main

import (
	"database/sql"
	"fmt"
)

// Parcel представляет собой структуру для хранения информации о посылке.
type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

// ParcelStore представляет собой структуру для работы с базой данных посылок.
type ParcelStore struct {
	db *sql.DB
}

// NewParcelStore создает новый экземпляр ParcelStore с указанным соединением с базой данных.
func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add добавляет новую посылку в базу данных и возвращает ее номер.
func (s ParcelStore) Add(p Parcel) (int, error) {
	query := `INSERT INTO parcels (client, status, address, created_at) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to add parcel: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return int(id), nil
}

// Get получает посылку по ее номеру.
func (s ParcelStore) Get(number int) (Parcel, error) {
	query := `SELECT number, client, status, address, created_at FROM parcels WHERE number = ?`
	row := s.db.QueryRow(query, number)

	p := Parcel{}
	if err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return Parcel{}, fmt.Errorf("no parcel found with number %d", number)
		}
		return Parcel{}, fmt.Errorf("failed to get parcel: %w", err)
	}

	return p, nil
}

// GetByClient получает все посылки для указанного клиента.
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	query := `SELECT number, client, status, address, created_at FROM parcels WHERE client = ?`
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get parcels for client %d: %w", client, err)
	}
	defer rows.Close()

	var res []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan parcel: %w", err)
		}
		res = append(res, p)
	}

	// Проверка на наличие ошибок после завершения цикла
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return res, nil
}

// SetStatus обновляет статус посылки по ее номеру.
func (s ParcelStore) SetStatus(number int, status string) error {
	query := `UPDATE parcels SET status = ? WHERE number = ?`
	_, err := s.db.Exec(query, status, number)
	if err != nil {
		return fmt.Errorf("failed to update status for parcel %d: %w", number, err)
	}
	return nil
}

// SetAddress обновляет адрес посылки, если статус равен "registered".
func (s ParcelStore) SetAddress(number int, address string) error {
	// Проверяем статус перед изменением адреса
	var status string
	err := s.db.QueryRow(`SELECT status FROM parcels WHERE number = ?`, number).Scan(&status)
	if err != nil {
		return fmt.Errorf("failed to get parcel status for %d: %w", number, err)
	}

	if status != "registered" {
		return fmt.Errorf("cannot change address for parcel %d with status %s", number, status)
	}

	query := `UPDATE parcels SET address = ? WHERE number = ?`
	_, err = s.db.Exec(query, address, number)
	if err != nil {
		return fmt.Errorf("failed to update address for parcel %d: %w", number, err)
	}
	return nil
}

// Delete удаляет посылку, если статус равен "registered".
func (s ParcelStore) Delete(number int) error {
	// Проверяем статус перед удалением
	var status string
	err := s.db.QueryRow(`SELECT status FROM parcels WHERE number = ?`, number).Scan(&status)
	if err != nil {
		return fmt.Errorf("failed to get parcel status for %d: %w", number, err)
	}

	if status != "registered" {
		return fmt.Errorf("cannot delete parcel %d with status %s", number, status)
	}

	query := `DELETE FROM parcels WHERE number = ?`
	_, err = s.db.Exec(query, number)
	if err != nil {
		return fmt.Errorf("failed to delete parcel %d: %w", number, err)
	}
	return nil
}
