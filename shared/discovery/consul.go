package discovery

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
	consulapi "github.com/hashicorp/consul/api"
	// Required for consul:// resolver to work with gRPC.
	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConsulRegistry represents a Consul based service registry.
type ConsulRegistry struct {
	config      *consulRegistryConfig
	client      *consulapi.Client
	healthCheck *healthCheckConfig
	logger      *zerolog.Logger
}

// NewConsulRegistry create a new ConsulRegistry instance with default configuration.
func NewConsulRegistry(logger *zerolog.Logger) (*ConsulRegistry, error) {
	cfg := newConsulRegistryConfig(logger)
	consulAPICfg := consulapi.DefaultConfig()
	consulAPICfg.Address = cfg.Address

	client, err := consulapi.NewClient(consulAPICfg)
	if err != nil {
		return nil, err
	}

	healthCheckCfg := newDefaultHealthCheckConfig()

	return &ConsulRegistry{
		config:      cfg,
		client:      client,
		healthCheck: healthCheckCfg,
		logger:      logger,
	}, nil
}

// Register adds a new service instance in Consul and starts a gRPC health check.
func (r *ConsulRegistry) Register(instanceID, serviceName, serviceAddress string) error {
	parts := strings.Split(serviceAddress, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid service address host:port format: %s", serviceAddress)
	}

	host := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}

	registration := &consulapi.AgentServiceRegistration{
		ID:      instanceID,
		Name:    serviceName,
		Address: host,
		Port:    port,
		Tags:    []string{"grpc"},
		Check: &consulapi.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%s:%d", host, port),
			Interval:                       r.healthCheck.Interval,
			Timeout:                        r.healthCheck.Timeout,
			DeregisterCriticalServiceAfter: r.healthCheck.DeregisterAfter,
		},
	}

	if err := r.client.Agent().ServiceRegister(registration); err != nil {
		return err
	}

	r.logger.Info().
		Str("instanceID", instanceID).
		Str("serviceName", serviceName).
		Msg("Successfully registered service")

	return nil
}

// Deregister removes a service instance from Consul.
func (r *ConsulRegistry) Deregister(instanceID, serviceName string) error {
	if err := r.client.Agent().ServiceDeregister(instanceID); err != nil {
		return err
	}

	r.logger.Info().
		Str("instanceID", instanceID).
		Str("serviceName", serviceName).
		Msg("Successfully deregistered service")

	return nil
}

// Connect establishes a gRPC connection to a service via Consul.
func (r *ConsulRegistry) Connect(serviceName string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("consul://%s/%s?tag=grpc&healthy=true", r.config.Address, serviceName),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`
			{
				"loadBalancingPolicy": "round_robin"
			}
		`),
	)
	if err != nil {
		return nil, err
	}

	r.logger.Info().
		Str("serviceName", serviceName).
		Msg("Successfully connected to service via Consul")

	return conn, nil
}

// healthCheckConfig represents the configuration for the gRPC health check.
type healthCheckConfig struct {
	Interval        string
	Timeout         string
	DeregisterAfter string
}

// newDefaultHealthCheckConfig creates a new healthCheckConfig instance with default values.
func newDefaultHealthCheckConfig() *healthCheckConfig {
	return &healthCheckConfig{
		Interval:        "30s",
		Timeout:         "5s",
		DeregisterAfter: "1m",
	}
}

// consulRegistryConfig represents the configuration for the Consul service registry.
type consulRegistryConfig struct {
	Address string `env:"CONSUL_ADDRESS"`
}

// newConsulRegistryConfig creates a new consulRegistryConfig instance from environment variables.
func newConsulRegistryConfig(logger *zerolog.Logger) *consulRegistryConfig {
	cfg, err := env.ParseAs[consulRegistryConfig]()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse environment variables")
	}

	return &cfg
}
