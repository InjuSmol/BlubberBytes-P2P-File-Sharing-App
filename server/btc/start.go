package btc // Package name: btc (this file belongs to the btc package)

import (
	"database/sql" // SQL database package
	"errors"       // Standard error handling package
	"fmt"          // For printing and scanning from stdin
	"os"           // OS-level functions (file system, env, etc.)
	"os/exec"      // For starting and controlling external processes
	"path/filepath" // For building filesystem paths in a portable way

	"server/database/operations" // Your custom package for database operations

	"github.com/btcsuite/btcd/btcutil" // Bitcoin utility functions (path helpers, etc.)
	"github.com/btcsuite/btcd/rpcclient" // Bitcoin RPC client (talk to btcd/btcwallet)
)

// Start starts Bitcoin-related services: btcd and btcwallet,
// ensures the wallet exists, gets the mining address, and returns everything ready.
func Start(net string, db *sql.DB, debug bool) (*exec.Cmd, *exec.Cmd, *rpcclient.Client, *rpcclient.Client, error) {
	pubPassphrase := "public" // Hardcoded public passphrase (for wallet encryption)
	var privPassphrase string // Private passphrase (user input)

	// Prompt user for private passphrase
	fmt.Print("Enter your private passphrase: ")
	_, err := fmt.Scanln(&privPassphrase) // Read input from console
	if err != nil {
		return nil, nil, nil, nil, err // Fail if unable to read
	}

	// Get wallet directory path (based on system, eg. ~/.btcwallet/)
	walletDir := btcutil.AppDataDir("btcwallet", false)

	// (Optional code commented out that would delete wallet.db — probably for dev resets)

	// Check if wallet.db exists at the right path
	if _, err := os.Stat(filepath.Join(walletDir, net+"/wallet.db")); errors.Is(err, os.ErrNotExist) {
		// If wallet.db does not exist, create a new wallet
		err := createWallet(walletDir, net, pubPassphrase, privPassphrase, db)
		if err != nil {
			return nil, nil, nil, nil, err // Fail if wallet creation fails
		}
	}

	// Store wallet passphrases into the database
	err = operations.UpdateWalletPassphrases(db, pubPassphrase, privPassphrase)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Retrieve wallet info (address, etc.) from database
	walletInfo, err := operations.GetWalletInfo(db)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	address := walletInfo.Address // Get the mining/receiving address

	// If no address stored yet, create a new one
	if address == "" {
		// Start temporary btcd and btcwallet instances without mining address
		btcdCmd, btcwalletCmd, btcd, btcwallet, err := startBtc(net, "", debug)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		// Generate a new address and store it in database
		address, err = operations.StoreAddress(btcwallet, db)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		// Gracefully shutdown the temporary clients
		ShutdownClient(btcd)
		ShutdownClient(btcwallet)

		// Stop the btcd and btcwallet processes
		InterruptCmd(btcwalletCmd)
		InterruptCmd(btcdCmd)
	}

	// Start final btcd and btcwallet instances, now with a mining address
	btcdCmd, btcwalletCmd, btcd, btcwallet, err := startBtc(net, address, debug)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Return the process handles and client connections
	return btcdCmd, btcwalletCmd, btcd, btcwallet, nil
}

// startBtc starts btcd (full node) and btcwallet processes and RPC clients.
// If an error happens at any step, it shuts everything down cleanly.
func startBtc(net string, miningaddr string, debug bool) (*exec.Cmd, *exec.Cmd, *rpcclient.Client, *rpcclient.Client, error) {
	// Start btcd process (optionally providing mining address)
	btcdCmd, err := startBtcd(net, miningaddr, debug)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Start btcwallet process
	btcwalletCmd, err := startBtcwallet(net, debug)
	if err != nil {
		InterruptCmd(btcdCmd) // Clean up btcd if btcwallet failed
		return nil, nil, nil, nil, err
	}

	// Create RPC client to communicate with btcd node
	btcd, err := createBtcdClient(net)
	if err != nil {
		InterruptCmd(btcwalletCmd) // Clean up btcwallet
		InterruptCmd(btcdCmd)
		return nil, nil, nil, nil, err
	}

	// Create RPC client to communicate with btcwallet
	btcwallet, err := createBtcwalletClient(net)
	if err != nil {
		ShutdownClient(btcd) // Shut down btcd client
		InterruptCmd(btcwalletCmd)
		InterruptCmd(btcdCmd)
		return nil, nil, nil, nil, err
	}

	// Return process handles and RPC clients
	return btcdCmd, btcwalletCmd, btcd, btcwallet, nil
}
