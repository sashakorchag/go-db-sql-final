package main

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test address",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// setupTestDB настраивает in-memory базу данных для тестирования
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE parcels (
			number INTEGER PRIMARY KEY AUTOINCREMENT,
			client INTEGER,
			status TEXT,
			address TEXT,
			created_at TEXT	)
	`)
	require.NoError(t, err)

	return db
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Get
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, storedParcel.Client)
	require.Equal(t, parcel.Status, storedParcel.Status)
	require.Equal(t, parcel.Address, storedParcel.Address)

	// Delete
	err = store.Delete(id)
	require.NoError(t, err)

	// Check that the parcel cannot be retrieved anymore
	_, err = store.Get(id)
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err) // Проверка на конкретную ошибку "не найдено"
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// Check
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, storedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// Add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Set status
	newStatus := ParcelStatusDelivered
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// Check
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, storedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	// Задаем всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// AddparcelMap := make(map[int]Parcel)
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// Get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// Check
	for _, parcel := range storedParcels {
		expected, exists := parcelMap[parcel.Number]
		require.True(t, exists)
		require.Equal(t, expected.Client, parcel.Client)
		require.Equal(t, expected.Status, parcel.Status)
		require.Equal(t, expected.Address, parcel.Address)
	}
}

// TestDeleteNonRegisteredParcel проверяет удаление посылки со статусом, отличным от "зарегистрирована"
func TestDeleteNonRegisteredParcel(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	parcel.Status = ParcelStatusDelivered // Установим статус, отличный от "зарегистрирована"

	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.Delete(id)
	require.Error(t, err)
	require.Equal(t, "parcel can only be deleted if status is 'registered'", err.Error())
}
