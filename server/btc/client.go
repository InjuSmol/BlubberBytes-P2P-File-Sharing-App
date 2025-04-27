package btc // Define the package name as "btc"

import (
	"github.com/btcsuite/btcd/rpcclient" // Import the rpcclient package for RPC communication
)

// Create a new RPC client using websockets.
func createClient(port string, net string) (*rpcclient.Client, error) {
	netParam := net // Start with the provided network name

	// If the network is "testnet", adjust the network parameter to "testnet3" (specific to Bitcoin network naming)
	if net == "testnet" {
		netParam = "testnet3"
	}

	// Create a configuration for the RPC client connection
	connCfg := &rpcclient.ConnConfig{
		Host:       "localhost:" + port, // Set the host (localhost + port)
		Endpoint:   "ws",                 // Use "ws" (WebSocket) as the communication method
		User:       "user",               // Username for RPC authentication
		Pass:       "password",           // Password for RPC authentication
		DisableTLS: true,                 // Disable TLS because we're connecting locally (no encryption)
		Params:     netParam,             // Set the network parameters (like "mainnet", "testnet3", etc.)
	}

	// Create a new RPC client using the above connection configuration
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err // If there's an error, return it
	}

	return client, nil // Return the successfully created client
}

// Create a new RPC client for btcd using websockets.
func createBtcdClient(net string) (*rpcclient.Client, error) {
	return createClient("8334", net) // Call createClient with btcd's port 8334
}

// Create a new RPC client for btcwallet using websockets.
func createBtcwalletClient(net string) (*rpcclient.Client, error) {
	return createClient("8332", net) // Call createClient with btcwallet's port 8332
}

// Shutdown a client properly.
func ShutdownClient(client *rpcclient.Client) {
	client.Shutdown()           // Initiate shutdown of the RPC client
	client.WaitForShutdown()    // Wait until the shutdown process is fully completed
}
