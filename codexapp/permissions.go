package codexapp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type normalizedPermissionProfile struct {
	ID         string
	WriteRoots []string
	Config     permissionProfileConfig
}

type permissionProfileConfig struct {
	Filesystem map[string]string       `json:"filesystem"`
	Network    permissionNetworkConfig `json:"network"`
}

type permissionNetworkConfig struct {
	Enabled bool `json:"enabled"`
}

func (p Permissions) empty() bool {
	return len(p.ReadRoots) == 0 && len(p.WriteRoots) == 0 && !p.NetworkEnabled
}

func buildPermissionProfile(permissions Permissions) (normalizedPermissionProfile, error) {
	readRoots, err := normalizeRoots(permissions.ReadRoots)
	if err != nil {
		return normalizedPermissionProfile{}, fmt.Errorf("codexapp: invalid read roots: %w", err)
	}
	writeRoots, err := normalizeRoots(permissions.WriteRoots)
	if err != nil {
		return normalizedPermissionProfile{}, fmt.Errorf("codexapp: invalid write roots: %w", err)
	}
	filesystem := map[string]string{":root": "deny", ":minimal": "read"}
	for _, root := range readRoots {
		filesystem[root] = "read"
	}
	for _, root := range writeRoots {
		filesystem[root] = "write"
	}
	config := permissionProfileConfig{
		Filesystem: filesystem,
		Network:    permissionNetworkConfig{Enabled: permissions.NetworkEnabled},
	}
	canonical := struct {
		ReadRoots      []string `json:"readRoots"`
		WriteRoots     []string `json:"writeRoots"`
		NetworkEnabled bool     `json:"networkEnabled"`
	}{readRoots, writeRoots, permissions.NetworkEnabled}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return normalizedPermissionProfile{}, fmt.Errorf("encode permission profile: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return normalizedPermissionProfile{
		ID:         "codexapp-" + hex.EncodeToString(sum[:8]),
		WriteRoots: writeRoots,
		Config:     config,
	}, nil
}

func normalizeRoots(roots []string) ([]string, error) {
	unique := make(map[string]struct{}, len(roots))
	for _, raw := range roots {
		root := filepath.Clean(strings.TrimSpace(raw))
		if raw == "" || root == "." {
			return nil, fmt.Errorf("root must not be blank")
		}
		if !filepath.IsAbs(root) {
			return nil, fmt.Errorf("root %q must be absolute", raw)
		}
		volumeRoot := filepath.Clean(filepath.VolumeName(root) + string(filepath.Separator))
		if root == volumeRoot {
			return nil, fmt.Errorf("filesystem root %q is not allowed", root)
		}
		unique[root] = struct{}{}
	}
	normalized := make([]string, 0, len(unique))
	for root := range unique {
		normalized = append(normalized, root)
	}
	sort.Strings(normalized)
	return normalized, nil
}
