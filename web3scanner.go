// Package web3scanner provides functionality for scanning and managing Ethereum addresses.
package web3scanner

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/google/uuid"

	"github.com/qiaopengjun5162/web3scanner/config"
	"github.com/qiaopengjun5162/web3scanner/database"
)

type Web3Scanner struct {
	db *database.DB

	shutdown context.CancelCauseFunc
	stopped  atomic.Bool
}

// NewWeb3Scanner creates a new instance of Web3Scanner.
//
// It takes a context, a configuration and a shutdown function. The context is used
// for database operations and the shutdown function is used to cancel the context
// when the Web3Scanner is shut down.
//
// The function returns a pointer to the new Web3Scanner instance and an error.
// The error is set if there was an error creating the database connection.
func NewWeb3Scanner(ctx context.Context, cfg *config.Config, shutdown context.CancelCauseFunc) (*Web3Scanner, error) {
	dba, err := database.NewDB(ctx, cfg.MasterDB)
	if err != nil {
		log.Error("init database fail", err)
		return nil, err
	}
	out := &Web3Scanner{
		db:       dba,
		shutdown: shutdown,
	}
	return out, nil
}

// Start starts the Web3Scanner.
//
// It takes a context and stores an address in the database. It then retrieves all
// addresses from the database and prints them out.
//
// The function returns an error if there was an error storing or retrieving the
// addresses.
func (ws *Web3Scanner) Start(_ context.Context) error {
	fmt.Println("Web3Scanner start .........")
	var batchAddress []database.Addresses
	addressItem := database.Addresses{
		GUID:        uuid.New(),
		Address:     common.HexToAddress("0x0fa09C3A328792253f8dee7116848723b72a6d2e"),
		AddressType: 1,
		PublicKey:   "0x0fa09C3A328792253f8dee7116848723b72a6d2e",
		Timestamp:   uint64(time.Now().Unix()),
	}
	batchAddress = append(batchAddress, addressItem)
	err := ws.db.Addresses.StoreAddresses(batchAddress)
	if err != nil {
		fmt.Println("store address fail")
		return err
	}

	addrList, err := ws.db.Addresses.GetAllAddresses()
	if err != nil {
		return err
	}
	for _, item := range addrList {
		fmt.Println("=======print address list==========")
		fmt.Println(item.Address)
		fmt.Println(item.Timestamp)
		fmt.Println(item.AddressType)
		fmt.Println("=======print address list==========")
	}
	return nil
}

// Stop stops the Web3Scanner.
//
// It prints a message to the console. It's currently a no-op, but it's a
// placeholder for future code that will do something more interesting.
func (ws *Web3Scanner) Stop(_ context.Context) error {
	fmt.Println("DbOp stop .........")
	return nil
}

// Stopped checks if the Web3Scanner has been stopped.
//
// It returns true if the scanner is stopped, false otherwise. This method
// relies on an atomic operation to safely retrieve the stopped state.
func (ws *Web3Scanner) Stopped() bool {
	return ws.stopped.Load()
}
