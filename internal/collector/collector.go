package collector

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/danielgo/goldmane-prometheus/internal/config"
	pb "github.com/danielgo/goldmane-prometheus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Collector handles the collection of flow data from Goldmane API
type Collector struct {
	cfg     *config.Config
	metrics *Metrics
	client  pb.FlowsClient
	conn    *grpc.ClientConn
}

// NewCollector creates a new Collector instance
func NewCollector(cfg *config.Config) (*Collector, error) {
	metrics := NewMetrics()

	return &Collector{
		cfg:     cfg,
		metrics: metrics,
	}, nil
}

// Connect establishes a connection to the Goldmane API
func (c *Collector) Connect(ctx context.Context) error {
	var opts []grpc.DialOption

	if c.cfg.TLSEnabled {
		// Load client cert if provided
		var creds credentials.TransportCredentials
		if c.cfg.TLSCertPath != "" && c.cfg.TLSKeyPath != "" {
			cert, err := tls.LoadX509KeyPair(c.cfg.TLSCertPath, c.cfg.TLSKeyPath)
			if err != nil {
				return fmt.Errorf("failed to load client cert: %w", err)
			}

			// Load CA cert if provided
			var certPool *x509.CertPool
			if c.cfg.TLSCAPath != "" {
				caCert, err := os.ReadFile(c.cfg.TLSCAPath)
				if err != nil {
					return fmt.Errorf("failed to read CA cert: %w", err)
				}
				certPool = x509.NewCertPool()
				if !certPool.AppendCertsFromPEM(caCert) {
					return fmt.Errorf("failed to append CA cert")
				}
			}

			creds = credentials.NewTLS(&tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      certPool,
			})
		} else {
			creds = credentials.NewTLS(&tls.Config{})
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(c.cfg.GoldmaneAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to Goldmane API: %w", err)
	}

	c.conn = conn
	c.client = pb.NewFlowsClient(conn)

	log.Printf("Connected to Goldmane API at %s", c.cfg.GoldmaneAddr)
	return nil
}

// Close closes the connection to the Goldmane API
func (c *Collector) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Start begins collecting flow data on the configured interval
func (c *Collector) Start(ctx context.Context) error {
	ticker := time.NewTicker(c.cfg.PollInterval)
	defer ticker.Stop()

	log.Printf("Starting flow collection with interval: %s", c.cfg.PollInterval)

	// Collect immediately on start
	if err := c.collectFlows(ctx); err != nil {
		log.Printf("Error collecting flows: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping flow collection")
			return ctx.Err()
		case <-ticker.C:
			if err := c.collectFlows(ctx); err != nil {
				log.Printf("Error collecting flows: %v", err)
			}
		}
	}
}

// collectFlows retrieves flow data from the Goldmane API and updates metrics
func (c *Collector) collectFlows(ctx context.Context) error {
	// Stream flows from the last poll interval
	req := &pb.FlowStreamRequest{
		StartTimeGte:        -int64(c.cfg.PollInterval.Seconds()),
		AggregationInterval: 15,
		Filter:              &pb.Filter{},
	}

	stream, err := c.client.Stream(ctx, req)
	if err != nil {
		c.metrics.APIRequests.WithLabelValues("error_stream").Inc()
		return fmt.Errorf("failed to create stream: %w", err)
	}

	flowCount := 0
	for {
		flowResult, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.metrics.APIRequests.WithLabelValues("error_recv").Inc()
			return fmt.Errorf("error receiving flow: %w", err)
		}

		c.processFlow(flowResult)
		flowCount++
	}

	c.metrics.APIRequests.WithLabelValues("success").Inc()
	c.metrics.APILastSuccessTime.SetToCurrentTime()
	c.metrics.APIFlowsProcessed.Add(float64(flowCount))

	log.Printf("Processed %d flows", flowCount)
	return nil
}

// processFlow processes a single flow and updates the appropriate metrics
func (c *Collector) processFlow(flowResult *pb.FlowResult) {
	if flowResult == nil || flowResult.Flow == nil || flowResult.Flow.Key == nil {
		return
	}

	flow := flowResult.Flow
	key := flow.Key

	// Extract label values
	labels := prometheus.Labels{
		"reporter":      c.getReporter(key.Reporter),
		"protocol":      key.Proto,
		"src_namespace": key.SourceNamespace,
		"src_pod":       key.SourceName,
		"src_port":      "0", // Source port is not directly available in FlowKey
		"dst_namespace": key.DestNamespace,
		"dst_object":    key.DestName,
		"dst_port":      strconv.FormatInt(key.DestPort, 10),
	}

	// Increment the appropriate counter based on action
	switch key.Action {
	case pb.Action_Allow:
		c.metrics.FlowAllow.With(labels).Add(float64(flow.NumConnectionsStarted))
	case pb.Action_Deny:
		c.metrics.FlowDeny.With(labels).Add(float64(flow.NumConnectionsStarted))
	}
}

// getReporter converts the Reporter enum to a string
func (c *Collector) getReporter(reporter pb.Reporter) string {
	switch reporter {
	case pb.Reporter_Src:
		return "src"
	case pb.Reporter_Dst:
		return "dst"
	default:
		return "unspecified"
	}
}
