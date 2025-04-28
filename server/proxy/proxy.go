/*
This is a SOCKS proxy using go. It logs the total number of ingoing and outgoing bytes
for each user (1 user = 1 IP address) and every 5 minutes this information is logged to
a txt file in the format [IP, bytes]\n[IP, bytes]\n[IP, bytes]
*/

package proxy

// Import necessary packages
import (
	"context" // For managing context-based data
	"database/sql" // For SQL database interaction
	"fmt" // For formatted output
	"log" // For logging
	"net" // For network-related functionality
	"strings" // For string manipulation
	"sync" // For managing concurrent access to shared resources
	"time" // For time-related operations

	"server/database/operations" // Custom package for database operations

	"github.com/armon/go-socks5" // Go package to implement a SOCKS5 proxy server
	"github.com/libp2p/go-libp2p/core/host" // Libp2p package for network host operations
)

var paymentInformation = make(map[string]int64) // Map to store payment data by IP address (in bytes)
var mutex sync.Mutex // Mutex to ensure thread-safe access to paymentInformation map

// Define the trafficInterceptor struct to intercept network traffic for each connection
type trafficInterceptor struct {
	conn     net.Conn // Underlying network connection
	clientIP string // Client IP address
	read     int64 // Total bytes received by the client
	written  int64 // Total bytes sent by the client
}

// Read method to intercept incoming traffic and log bytes received
func (t *trafficInterceptor) Read(b []byte) (n int, err error) {
	n, err = t.conn.Read(b) // Read data from the underlying connection
	if err == nil {
		t.read += int64(n) // Track the number of bytes read
	}
	return
}

// Write method to intercept outgoing traffic and log bytes sent
func (t *trafficInterceptor) Write(b []byte) (n int, err error) {
	n, err = t.conn.Write(b) // Write data to the underlying connection
	if err == nil {
		t.written += int64(n) // Track the number of bytes written
	}
	return
}

// Close method to log final data transfer statistics and update payment info
func (t *trafficInterceptor) Close() error {
	log.Printf("Final bytes received: %d", t.read) // Log bytes received
	log.Printf("Final bytes sent: %d", t.written) // Log bytes sent
	log.Printf("IP Of the bytes above: %s", t.clientIP) // Log the client IP address

	// Update the payment information for the client (client IP and total bytes)
	updatePaymentInfo(strings.Split(t.clientIP, ":")[0], t.read+t.written)

	return t.conn.Close() // Close the underlying network connection
}

// LocalAddr returns the local address of the connection (unused in this example)
func (t *trafficInterceptor) LocalAddr() net.Addr {
	return t.conn.LocalAddr()
}

// RemoteAddr returns the remote address of the connection
func (t *trafficInterceptor) RemoteAddr() net.Addr {
	return t.conn.RemoteAddr()
}

// SetDeadline sets the read/write deadline for the connection (unused in this example)
func (t *trafficInterceptor) SetDeadline(deadline time.Time) error {
	return t.conn.SetDeadline(deadline)
}

// SetReadDeadline sets the read deadline for the connection
func (t *trafficInterceptor) SetReadDeadline(deadline time.Time) error {
	return t.conn.SetReadDeadline(deadline)
}

// SetWriteDeadline sets the write deadline for the connection
func (t *trafficInterceptor) SetWriteDeadline(deadline time.Time) error {
	return t.conn.SetWriteDeadline(deadline)
}

// GetBytesSent returns the total bytes sent by the client
func (t *trafficInterceptor) GetBytesSent() int64 {
	return t.written
}

// GetBytesReceived returns the total bytes received by the client
func (t *trafficInterceptor) GetBytesReceived() int64 {
	return t.read
}

// Define clientAddressRuleset to handle client address-specific rules for SOCKS5 proxy
type clientAddressRuleset struct {
	socks5.RuleSet
}

// Function to update payment information for the client
func updatePaymentInfo(key string, value int64) {
	mutex.Lock() // Lock the mutex to ensure safe access to the shared resource
	defer mutex.Unlock()

	log.Printf("Key: %s value: %d\n", key, value)

	// If the client IP already exists in the map, add the new bytes to the existing total
	if currentBytes, exists := paymentInformation[key]; exists {
		paymentInformation[key] = currentBytes + value
	} else {
		paymentInformation[key] = value // If not, create a new entry with the given bytes
	}
}

// Allow function for handling connection requests for the SOCKS5 proxy
func (r *clientAddressRuleset) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	if req.RemoteAddr != nil {
		clientIP := req.RemoteAddr.String() // Get the client's IP address
		log.Printf("Client IP: %s", clientIP) // Log the client IP address
		return context.WithValue(ctx, "clientIP", clientIP), true // Add the client IP to context and allow the connection
	}

	return ctx, true // Allow the connection if no IP address is found
}

// Custom dial function to intercept traffic and wrap the connection
func customDial(ctx context.Context, network, addr string) (net.Conn, error) {
	conn, err := net.Dial(network, addr) // Dial the network connection
	if err != nil {
		return nil, err // Return an error if unable to dial
	}

	clientIP, _ := ctx.Value("clientIP").(string) // Retrieve the client IP from context

	// Return a wrapped connection with the trafficInterceptor to monitor traffic
	return &trafficInterceptor{conn: conn, clientIP: clientIP}, nil
}

// Main Proxy function that sets up and runs the SOCKS5 proxy server
func Proxy(node host.Host, db *sql.DB) {
	dial := customDial // Define the custom dial function to intercept traffic
	conf := &socks5.Config{Dial: dial, Rules: &clientAddressRuleset{}} // Set up SOCKS5 config with custom dial and rules
	server, err := socks5.New(conf) // Create a new SOCKS5 server
	if err != nil {
		panic(err) // Panic if server creation fails
	}

	// Start a goroutine that logs traffic information to the database every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Set up ticker to run every 30 seconds
		for range ticker.C {
			mutex.Lock() // Lock the mutex to safely access the shared resource

			// Loop through the payment information map and log to the database
			for key, value := range paymentInformation {
				log.Printf("%s : %d", key, value) // Log the IP and its transferred data
				operations.AddProxyLogs(db, key, value, time.Now().Unix()) // Insert data into the database
			}

			// Clear the map after logging
			for key := range paymentInformation {
				delete(paymentInformation, key)
			}

			mutex.Unlock() // Unlock the mutex
		}
	}()

	fmt.Println("Proxy is running on http://localhost:8000.") // Log message indicating the proxy is running

	// Start the SOCKS5 proxy server on localhost port 8000
	if err := server.ListenAndServe("tcp", "0.0.0.0:8000"); err != nil {
		panic(err) // Panic if the server fails to start
	}
}
