package btc // Define the package name as "btc"

import (
	"bufio"    // For reading output line-by-line
	"errors"   // For returning error values
	"fmt"      // For printing debug output
	"os/exec"  // For starting external processes
	"strings"  // For working with string operations
)

// Start the btcd process.
func startBtcd(net string, miningaddr string, debug bool) (*exec.Cmd, error) {
	netCmd := "" // Initialize an empty network argument
	if net != "mainnet" {
		netCmd = "--" + net // If not mainnet, add a flag like "--testnet" or "--simnet"
	}

	// Set default public node IP and port
	publicNode := "130.245.173.221:8333"
	if net == "testnet" {
		publicNode = "130.245.173.221:18333" // Change port if testnet
	}

	miningaddrCmd := "" // Initialize mining address argument
	if miningaddr != "" {
		miningaddrCmd = "--miningaddr=" + miningaddr // If provided, add mining address flag
	}

	// Create the command to start btcd with the correct flags
	cmd := exec.Command(
		"./btcd/btcd",            // Path to btcd binary
		"-C", "./btc/conf/btcd.conf", // Config file path
		netCmd,                   // Network (mainnet, testnet, etc.)
		"--connect="+publicNode,  // Connect to the specified public node
		miningaddrCmd,            // Optional mining address
	)

	cmd.SysProcAttr = sysProcAttr // Platform-specific process attributes (e.g., set process group)

	// Get a pipe to read btcd's standard output
	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(cmdStdout) // Create a scanner to read the output

	defer func() {
		go func() {
			// Asynchronously scan and optionally print btcd's output
			for scanner.Scan() {
				if debug {
					fmt.Println(scanner.Text())
				}
			}
		}()
	}()

	// Start the btcd process
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	// Read lines from btcd output to detect successful start
	for scanner.Scan() {
		if debug {
			fmt.Println(scanner.Text())
		}

		if net == "mainnet" || net == "testnet" {
			// For mainnet/testnet, wait until synchronization starts
			if strings.Contains(scanner.Text(), "Syncing to block height") {
				return cmd, nil
			}
		} else {
			// For regtest/simnet, wait until RPC server is ready
			if strings.Contains(scanner.Text(), "RPC server listening") {
				return cmd, nil
			}
		}
	}

	// If we reach here, btcd did not start correctly
	return nil, errors.New("failed to start btcd")
}

// Start the btcwallet process.
func startBtcwallet(net string, debug bool) (*exec.Cmd, error) {
	netCmd := "" // Initialize network argument
	if net != "mainnet" {
		netCmd = "--" + net // Add network flag if not mainnet
	}

	// Create the command to start btcwallet
	cmd := exec.Command(
		"./btcwallet/btcwallet",   // Path to btcwallet binary
		"-C", "./btc/conf/btcwallet.conf", // Config file
		netCmd,                    // Network flag
	)

	cmd.SysProcAttr = sysProcAttr // Set system-specific attributes

	// Get a pipe to read btcwallet's standard output
	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(cmdStdout) // Create a scanner to read the output

	defer func() {
		go func() {
			// Asynchronously scan and optionally print btcwallet's output
			for scanner.Scan() {
				if debug {
					fmt.Println(scanner.Text())
				}
			}
		}()
	}()

	// Start the btcwallet process
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	// Flags to confirm both RPC connection and wallet opening
	rpc := false
	wallet := false

	// Read btcwallet output and look for key startup messages
	for scanner.Scan() {
		if debug {
			fmt.Println(scanner.Text())
		}

		if strings.Contains(scanner.Text(), "Established connection to RPC server") {
			rpc = true // RPC server connection is ready
		} else if strings.Contains(scanner.Text(), "Opened wallet") {
			wallet = true // Wallet is successfully opened
		}

		// When both RPC connection and wallet are ready, consider startup successful
		if rpc && wallet {
			return cmd, nil
		}
	}

	// If we reach here, btcwallet did not start correctly
	return nil, errors.New("failed to start btcwallet")
}
