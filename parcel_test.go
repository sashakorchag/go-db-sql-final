package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)

	parcel.Number = id

	parc, err := store.Get(parcel.Number)
	assert.NoError(t, err)
	assert.Equal(t, parcel, parc)

	err = store.Delete(parcel.Number)
	require.NoError(t, err)

	_, err = store.Get(parcel.Number)
	require.Error(t, err)

}

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)
	parcel.Number = id
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)
	parc, err := store.Get(parcel.Number)
	assert.NoError(t, err)
	assert.Equal(t, newAddress, parc.Address)
}

func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)
	parcel.Number = id
	err = store.SetStatus(parcel.Number, ParcelStatusSent)
	require.NoError(t, err)
	parc, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, parc.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.Greater(t, id, 0)

		parcels[i].Number = id
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels))

	parcelMap := make(map[int]Parcel)
	for _, p := range parcels {
		parcelMap[p.Number] = p
	}

	for _, parcel := range storedParcels {
		expectedParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists, "Посылка с ID %d не найдена в мапе", parcel.Number)
		require.Equal(t, expectedParcel, parcel, "Данные посылки с ID %d не совпадают", parcel.Number)
	}
}
