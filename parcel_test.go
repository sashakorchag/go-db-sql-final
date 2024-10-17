package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Подключаем драйвер SQLite
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered, // Используем константу	Address: "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// setupTestDB создает временную базу данных в памяти
func setupTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Создаем таблицу для тестов
	createTableQuery := `
	CREATE TABLE parcels (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT);`
	if _, err := db.Exec(createTableQuery); err != nil {
		return nil, err
	}

	return db, nil
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// get
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, got.Client)
	require.Equal(t, parcel.Status, got.Status)
	require.Equal(t, parcel.Address, got.Address)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	// check that parcel cannot be retrieved
	_, err = store.Get(id)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, got.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// set status
	newStatus := ParcelStatusDelivered // Используем константу
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// check
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, got.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := setupTestDB()
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id // сохраняем добавленную посылку в структуру map	parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		require.Contains(t, parcelMap, parcel.Number)
		// убедитесь, что значения полей полученных посылок заполнены верно
		require.Equal(t, parcelMap[parcel.Number].Client, parcel.Client)
		require.Equal(t, parcelMap[parcel.Number].Status, parcel.Status)
		require.Equal(t, parcelMap[parcel.Number].Address, parcel.Address)
	}
}
