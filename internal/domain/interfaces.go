package domain

import "context"

// Deployer handles deployment of a binary and its configuration.
type Deployer interface {
	Deploy(ctx context.Context, req DeployRequest) error
}

// StatsCollector collects system statistics.
type StatsCollector interface {
	Collect(ctx context.Context) (StatsResponse, error)
}

// ServiceManager manages systemd services.
type ServiceManager interface {
	RestartService(ctx context.Context, name string) error
	ServiceStatus(ctx context.Context, name string) (string, error)
}

// BinaryManager handles downloading and placing binaries.
type BinaryManager interface {
	EnsureBinary(ctx context.Context, url, targetPath string) error
}
