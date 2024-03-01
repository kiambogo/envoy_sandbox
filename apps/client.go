package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	pb "main/proto"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	serverAddress string
	qps           int
	duration      time.Duration
)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_client_requests_total",
			Help: "Total number of gRPC client requests",
		},
		[]string{"status"},
	)

	latencyHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "grpc_client_request_latency_seconds",
			Help: "Histogram of gRPC client request latencies",
		},
		[]string{"status"},
	)
)

func init() {
	clientCmd.Flags().StringVar(&serverAddress, "address", "localhost:9090", "gRPC server address")
	clientCmd.Flags().IntVar(&qps, "qps", 10, "queries per second")
	clientCmd.Flags().DurationVar(&duration, "duration", 5*time.Minute, "duration of runtime")

	rootCmd.AddCommand(clientCmd)

	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(latencyHistogram)
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Run gRPC client",
	Run: func(cmd *cobra.Command, args []string) {
		runClient()
	},
}

func runClient() {
	log.Println("Starting gRPC client")
	proxyURL := os.Getenv("PROXY_URL")

	targetURL := os.Getenv("TARGET_URL")
	if targetURL == "" {
		log.Fatal("TARGET_URL environment variable is not set")
	}

	log.Printf("Using proxy: %v", proxyURL)

	proxiedTransport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(proxyURL)
		},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	dialer := func(ctx context.Context, address string) (net.Conn, error) {
		log.Printf("Dialing %v", address)
		return proxiedTransport.DialContext(ctx, "tcp", address)
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer), grpc.WithAuthority(targetURL))

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a gRPC client.
	client := pb.NewGreeterClient(conn)

	// Calculate the interval between requests based on the specified queries per second.
	interval := time.Second / time.Duration(qps)

	timer := time.NewTimer(duration)
	ticker := time.NewTicker(interval)

	log.Printf("Generating traffic for %v...", duration)

	//	go func() {
	for {
		select {
		case <-ticker.C:
			generateTraffic(client)
		case _ = <-timer.C:
			log.Println("reached end of duration, exiting")
			return
		}
	}
	//}()

	// Start an HTTP server for Prometheus metrics
	//http.Handle("/metrics", promhttp.Handler())
	//log.Fatal(http.ListenAndServe(":8080", nil))
}

func generateTraffic(client pb.GreeterClient) {
	// Set a timeout for each RPC call.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	ctx = context.WithValue(ctx, "x-request-id", uuid.New().String())
	defer cancel()

	// Record the start time of the request
	startTime := time.Now()

	log.Printf("Sending request with x-request-id: %v", ctx.Value("x-request-id"))

	// Call the SayHello RPC.
	_, err := client.SayHello(ctx, &pb.HelloRequest{})
	if err != nil {
		log.Printf("Error calling SayHello: %v", err)
		statusCode := status.Code(err)
		requestCounter.WithLabelValues(statusCode.String()).Inc()
		// Observe the latency for failed requests
		latencyHistogram.WithLabelValues(statusCode.String()).Observe(time.Since(startTime).Seconds())
		return
	}

	log.Printf("Received response for x-request-id: %v", ctx.Value("x-request-id"))

	// Increment the total successful requests counter.
	requestCounter.WithLabelValues("success").Inc()

	// Observe the latency for successful requests
	latencyHistogram.WithLabelValues("success").Observe(time.Since(startTime).Seconds())
}
