package main

import (
	"database/sql"
	"fmt"
	"time"

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
	Status    stringAddress   string
	CreatedAt string
}

type ParcelStore interface {
	Add(parcel Parcel) (int, error)
	Get(number int) (Parcel, error)
	GetByClient(client int) ([]Parcel, error)
	SetStatus(number int, status string) error
	SetAddress(number int, address string) error
	Delete(number int) error
}

type SQLiteParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return &SQLiteParcelStore{db: db}
}

func (s *SQLiteParcelStore) Add(parcel Parcel) (int, error) {
	result, err := s.db.Exec(
		"INSERT INTO parcels (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		parcel.Client, parcel.Status, parcel.Address, parcel.CreatedAt,
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

func (s *SQLiteParcelStore) Get(number int) (Parcel, error) {
	var parcel Parcel
	err := s.db.QueryRow(
		"SELECT number, client, status, address, created_at FROM parcels WHERE number = ?",
		number,
	).Scan(&parcel.Number, &parcel.Client, &parcel.Status, &parcel.Address, &parcel.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}
	return parcel, nil
}

func (s *SQLiteParcelStore) GetByClient(client int) ([]Parcel, error) {
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
		var parcel Parcel
		if err := rows.Scan(&parcel.Number, &parcel.Client, &parcel.Status, &parcel.Address, &parcel.CreatedAt); err != nil {
			return nil, err	}
		parcels = append(parcels, parcel)
	}
	if err := rows.Err(); err != nil {
		return nil, err}
	return parcels, nil
}

func (s *SQLiteParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec("UPDATE parcels SET status = ? WHERE number = ?", status, number)
	return err
}

func (s *SQLiteParcelStore) SetAddress(number int, address string) error {
	// Проверяем статус перед обновлением адреса
	var status string
	err := s.db.QueryRow("SELECT status FROM parcels WHERE number = ?", number).Scan(&status)
	if err != nil {
		return err}

	// Изменять адрес можно только если статус 'registered'
	if status != ParcelStatusRegistered {
		return fmt.Errorf("address can only be updated if status is 'registered'")
	}

	_, err = s.db.Exec("UPDATE parcels SET address = ? WHERE number = ?", address, number)
	return err
}

func (s *SQLiteParcelStore) Delete(number int) error {
	// Проверяем статус перед удалением
	var status string
	err := s.db.QueryRow("SELECT status FROM parcels WHERE number = ?", number).Scan(&status)
	if err != nil {
		return err}

	// Удалять можно только если статус 'registered'
	if status != ParcelStatusRegistered {
		return fmt.Errorf("parcel can only be deleted if status is 'registered'")
	}

	_, err = s.db.Exec("DELETE FROM parcels WHERE number = ?", number)
	return err
}