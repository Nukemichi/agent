package system

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

// BinaryManager implements domain.BinaryManager.
type BinaryManager struct{}

// NewBinaryManager creates a new BinaryManager.
func NewBinaryManager() *BinaryManager {
	return &BinaryManager{}
}

// EnsureBinary downloads the binary from url to targetPath if not already present.
func (b *BinaryManager) EnsureBinary(ctx context.Context, url, targetPath string) error {
	if _, err := os.Stat(targetPath); err == nil {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download binary: unexpected status %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "agent-binary-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	if err := os.Chmod(tmpName, 0755); err != nil {
		return fmt.Errorf("chmod binary: %w", err)
	}

	if err := os.Rename(tmpName, targetPath); err != nil {
		return fmt.Errorf("move binary to target: %w", err)
	}

	return nil
}
