package database

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/google/uuid"

	"github.com/ethereum/go-ethereum/common"
)

// Addresses 结构体用于表示地址信息，包括用户地址、热钱包地址和冷钱包地址。
// 它通过GUID进行唯一标识，并存储了地址类型、公钥以及时间戳信息。
type Addresses struct {
	// GUID 是 Address 的唯一标识符，使用 UUID 类型，并且是主键。
	// 在 JSON 中表示为 "guid"。
	GUID uuid.UUID `gorm:"primaryKey" json:"guid"`

	// Address 存储了实际的地址信息，使用 common.Address 类型。
	// 它被序列化为字节存储，并在 JSON 中表示为 "address"。
	Address common.Address `json:"address" gorm:"serializer:bytes"`

	// AddressType 是一个 uint8 类型的字段，用于区分地址的类型。
	// 0 表示用户地址，1 表示热钱包地址（归集地址），2 表示冷钱包地址。
	AddressType uint8 `json:"addressType"`

	// PublicKey 存储了与地址相关的公钥信息，以字符串形式表示。
	// 在 JSON 中表示为 "publicKey"。
	PublicKey string `json:"publicKey"`

	// Timestamp 存储了地址创建的时间戳，为 uint64 类型。
	// 它用于记录地址的创建时间。
	Timestamp int64
}

// AddressesView defines the interface for querying address-related information.
// It includes methods for checking the existence of addresses, querying address details,
// and obtaining wallet information.
type AddressesView interface {
	// AddressExist returns whether the given address exists in the database and
	// the type of the address if it exists. If the address does not exist,
	// returns false and 0.
	AddressExist(address *common.Address) (bool, uint8)
	// QueryAddressesByToAddress returns the Addresses entry with the given address
	// if it exists. If the address does not exist, returns nil and gorm.ErrRecordNotFound.
	QueryAddressesByToAddress(*common.Address) (*Addresses, error)
	// QueryHotWalletInfo returns the Addresses entry with the hot wallet address
	// if it exists. If the address does not exist, returns nil and gorm.ErrRecordNotFound.
	QueryHotWalletInfo() (*Addresses, error)
	// QueryColdWalletInfo returns the Addresses entry with the cold wallet address
	// if it exists. If the address does not exist, returns nil and gorm.ErrRecordNotFound.
	QueryColdWalletInfo() (*Addresses, error)
	// GetAllAddresses returns all Addresses entries in the database.
	// It returns a slice of Addresses and a nil error if successful.
	// If there is an error, it returns a nil slice and the error.
	GetAllAddresses() ([]*Addresses, error)
}

// AddressesDB 定义了一个接口，用于管理地址数据的存储和检索。
// 它继承了 AddressesView 接口，意味着它拥有 AddressesView 接口的所有功能，
// 同时增加了存储地址数据的能力。
type AddressesDB interface {
	AddressesView

	// StoreAddresses 方法用于存储一组地址数据。
	// 参数:
	//   - []Addresses: 一个地址数据的切片，表示要存储的多个地址。
	// 返回值:
	//   - error: 如果存储过程中发生错误，返回一个描述错误的 error 对象；否则返回 nil。
	StoreAddresses([]Addresses) error
}

type addressesDB struct {
	gorm *gorm.DB
}

func (db *addressesDB) AddressExist(address *common.Address) (bool, uint8) {
	var addressEntry Addresses
	err := db.gorm.Table("addresses").Where("address", strings.ToLower(address.String())).First(&addressEntry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, 0
		}
		return false, 0
	}
	return true, addressEntry.AddressType
}

func (db *addressesDB) QueryAddressesByToAddress(address *common.Address) (*Addresses, error) {
	var addressEntry Addresses
	err := db.gorm.Table("addresses").Where("address", strings.ToLower(address.String())).Take(&addressEntry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &addressEntry, nil
}

// NewAddressesDB returns a new instance of the AddressesDB interface, which is
// backed by the given Gorm DB.
//
// The AddressesDB interface provides methods for accessing and manipulating the
// addresses table in the database.
//
// The returned AddressesDB instance is safe for concurrent use by multiple
// goroutines.
func NewAddressesDB(db *gorm.DB) AddressesDB {
	return &addressesDB{gorm: db}
}

// StoreAddresses store address
func (db *addressesDB) StoreAddresses(addressList []Addresses) error {
	result := db.gorm.Table("addresses").CreateInBatches(&addressList, len(addressList))
	return result.Error
}

func (db *addressesDB) QueryHotWalletInfo() (*Addresses, error) {
	var addressEntry Addresses
	err := db.gorm.Table("addresses").Where("address_type", 1).Take(&addressEntry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &addressEntry, nil
}

func (db *addressesDB) QueryColdWalletInfo() (*Addresses, error) {
	var addressEntry Addresses
	err := db.gorm.Table("addresses").Where("address_type", 2).Take(&addressEntry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &addressEntry, nil
}

func (db *addressesDB) GetAllAddresses() ([]*Addresses, error) {
	var addresses []*Addresses
	err := db.gorm.Table("addresses").Find(&addresses).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return addresses, nil
}
