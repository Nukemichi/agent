package usecase

import (
	"context"
	"fmt"
	"os"

	"agent-michi/internal/domain"
	"agent-michi/internal/infrastructure/system"
)

const (
	binaryPath  = "/usr/local/bin/sing-box"
	configPath  = "/etc/sing-box/config.json"
	serviceName = "sing-box"
)

// DeployUseCase orchestrates the full deploy lifecycle.
type DeployUseCase struct {
	binaryMgr domain.BinaryManager
	svcMgr    domain.ServiceManager
}

// NewDeployUseCase creates a new DeployUseCase.
func NewDeployUseCase(binaryMgr domain.BinaryManager, svcMgr domain.ServiceManager) *DeployUseCase {
	return &DeployUseCase{
		binaryMgr: binaryMgr,
		svcMgr:    svcMgr,
	}
}

// Deploy runs the full deploy sequence.
func (d *DeployUseCase) Deploy(ctx context.Context, req domain.DeployRequest) error {
	if err := d.binaryMgr.EnsureBinary(ctx, req.BinaryURL, binaryPath); err != nil {
		return fmt.Errorf("ensure binary: %w", err)
	}

	if err := os.MkdirAll("/etc/sing-box", 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := os.WriteFile(configPath, req.Config, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if err := system.WriteUnitFile(serviceName, binaryPath, configPath); err != nil {
		return fmt.Errorf("write unit file: %w", err)
	}

	if err := system.DaemonReload(ctx); err != nil {
		return fmt.Errorf("daemon reload: %w", err)
	}

	if err := d.svcMgr.RestartService(ctx, serviceName); err != nil {
		return fmt.Errorf("restart service: %w", err)
	}

	return nil
}
