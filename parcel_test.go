package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// setupTestDB создает временную базу данных в памяти
func setupTestDB(t *testing.T) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	createTableQuery := `
	CREATE TABLE parcels (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT);`
	_, err = db.Exec(createTableQuery)
	require.NoError(t, err)

	return db, nil
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db, err := setupTestDB(t)
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// get
	retrievedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, retrievedParcel.Client)
	require.Equal(t, parcel.Status, retrievedParcel.Status)
	require.Equal(t, parcel.Address, retrievedParcel.Address)
	require.Equal(t, parcel.CreatedAt, retrievedParcel.CreatedAt)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	// verify deletion
	_, err = store.Get(id)
	require.Error(t, err) // Expect an error since the parcel should be deleted
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db, err := setupTestDB(t)
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	retrievedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, retrievedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := setupTestDB(t)
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// set status
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// check
	retrievedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, retrievedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db, err := setupTestDB(t)
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id 
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		require.Contains(t, parcels, parcel) // Ensure that each retrieved parcel is in the original list
	}
}
