package types

import "encoding/binary"

const (
	// ModuleName defines the module name
	ModuleName = "sharding"

	// StoreKey is string representation of the store key for marker
	StoreKey = ModuleName
)

var (
	// PetKeyPrefix is the key for the stored pet
	PetKeyPrefix = []byte{0x01}
	// PetInfoPrefix is the key for the stored pet info
	PetInfoPrefix = []byte{0x02}
	// OwnerPrefix is the key for the stored pet owner
	OwnerPrefix = []byte{0x03}
	// NamePrefix is the key for the stored pet name
	NamePrefix = []byte{0x05}
	// ColorPrefix is the key for the stored pet color
	ColorPrefix = []byte{0x06}
	// SpotsPrefix is the key for the stored pet spotss
	SpotsPrefix = []byte{0x07}
)

// GetPetKey is a function to get the key for the stored pet object
func GetPetKey(petID uint64) []byte {
	petIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(petIDBytes, petID)

	key := PetKeyPrefix
	key = append(key, petIDBytes...)
	return key
}

// GetPetKey is a function to get the key for the stored pet object
func GetPetInfoKey(petID uint64) []byte {
	petIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(petIDBytes, petID)

	key := PetInfoPrefix
	key = append(key, petIDBytes...)
	return key
}

// GetPetOwnerKey is a function to get the key for the stored pet owner
func GetPetOwnerKey(petID uint64) []byte {
	petIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(petIDBytes, petID)

	key := OwnerPrefix
	key = append(key, petIDBytes...)
	return key
}

// GetPetNameKey is a function to get the key for the stored pet name
func GetPetNameKey(petID uint64) []byte {
	petIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(petIDBytes, petID)

	key := NamePrefix
	key = append(key, petIDBytes...)
	return key
}

// GetPetColorKey is a function to get the key for the stored pet color
func GetPetColorKey(petID uint64) []byte {
	petIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(petIDBytes, petID)

	key := ColorPrefix
	key = append(key, petIDBytes...)
	return key
}

// GetPetSpotsKey is a function to get the key for the stored pet spots
func GetPetSpotsKey(petID uint64) []byte {
	petIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(petIDBytes, petID)

	key := SpotsPrefix
	key = append(key, petIDBytes...)
	return key
}
