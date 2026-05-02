package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const unitTemplate = `[Unit]
Description=sing-box service
After=network.target

[Service]
ExecStart=/usr/local/bin/sing-box run -c /etc/sing-box/config.json
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
`

// SystemdManager implements domain.ServiceManager.
type SystemdManager struct{}

// NewSystemdManager creates a new SystemdManager.
func NewSystemdManager() *SystemdManager {
	return &SystemdManager{}
}

// WriteUnitFile writes a systemd unit file for the given service.
func WriteUnitFile(serviceName, binaryPath, configPath string) error {
	unitPath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	if err := os.WriteFile(unitPath, []byte(unitTemplate), 0644); err != nil {
		return fmt.Errorf("write unit file: %w", err)
	}
	return nil
}

// DaemonReload runs systemctl daemon-reload.
func DaemonReload(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "systemctl", "daemon-reload")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("daemon-reload: %w: %s", err, out)
	}
	return nil
}

// RestartService restarts the named systemd service.
func (s *SystemdManager) RestartService(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "systemctl", "restart", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("restart service %s: %w: %s", name, err, out)
	}
	return nil
}

// ServiceStatus returns "active" or "inactive" for the named service.
func (s *SystemdManager) ServiceStatus(ctx context.Context, name string) (string, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "is-active", name)
	out, err := cmd.Output()
	status := strings.TrimSpace(string(out))
	if err != nil {
		// is-active exits non-zero when inactive; still return the status string.
		if status == "" {
			return "inactive", nil
		}
		return status, nil
	}
	return status, nil
}
