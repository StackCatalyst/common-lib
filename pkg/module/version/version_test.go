package version

import (
	"testing"

	"github.com/StackCatalyst/common-lib/pkg/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionParsing(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "valid version",
			version: "1.2.3",
			wantErr: false,
		},
		{
			name:    "valid version with prerelease",
			version: "1.2.3-beta.1",
			wantErr: false,
		},
		{
			name:    "valid version with metadata",
			version: "1.2.3+20240101",
			wantErr: false,
		},
		{
			name:    "invalid version",
			version: "invalid",
			wantErr: true,
		},
		{
			name:    "empty version",
			version: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := manager.Parse(tt.version)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, v)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, v)
				assert.Equal(t, tt.version, v.String())
			}
		})
	}
}

func TestVersionResolution(t *testing.T) {
	manager := NewManager()
	versions := []string{
		"1.0.0",
		"1.1.0",
		"1.2.0",
		"2.0.0",
		"2.1.0",
		"2.1.1",
	}

	tests := []struct {
		name       string
		constraint string
		want       string
		wantErr    bool
	}{
		{
			name:       "exact version",
			constraint: "=2.1.0",
			want:       "2.1.0",
			wantErr:    false,
		},
		{
			name:       "greater than",
			constraint: ">1.0.0",
			want:       "2.1.1",
			wantErr:    false,
		},
		{
			name:       "compatible version",
			constraint: "~2.1.0",
			want:       "2.1.1",
			wantErr:    false,
		},
		{
			name:       "no matching version",
			constraint: ">3.0.0",
			want:       "",
			wantErr:    true,
		},
		{
			name:       "invalid constraint",
			constraint: "invalid",
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.Resolve(tt.constraint, versions)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestVersionComparison(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name    string
		v1      string
		v2      string
		want    int
		wantErr bool
	}{
		{
			name:    "equal versions",
			v1:      "1.0.0",
			v2:      "1.0.0",
			want:    0,
			wantErr: false,
		},
		{
			name:    "greater version",
			v1:      "2.0.0",
			v2:      "1.0.0",
			want:    1,
			wantErr: false,
		},
		{
			name:    "lesser version",
			v1:      "1.0.0",
			v2:      "2.0.0",
			want:    -1,
			wantErr: false,
		},
		{
			name:    "invalid first version",
			v1:      "invalid",
			v2:      "1.0.0",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid second version",
			v1:      "1.0.0",
			v2:      "invalid",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.Compare(tt.v1, tt.v2)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestVersionValidation(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "valid version",
			version: "1.0.0",
			want:    true,
		},
		{
			name:    "valid version with prerelease",
			version: "1.0.0-alpha.1",
			want:    true,
		},
		{
			name:    "valid version with metadata",
			version: "1.0.0+20240101",
			want:    true,
		},
		{
			name:    "invalid version",
			version: "invalid",
			want:    false,
		},
		{
			name:    "empty version",
			version: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := manager.IsValid(tt.version)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVersionConstraints(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name       string
		version    string
		constraint string
		want       bool
		wantErr    bool
	}{
		{
			name:       "exact match",
			version:    "1.0.0",
			constraint: "=1.0.0",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "greater than",
			version:    "2.0.0",
			constraint: ">1.0.0",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "not satisfied",
			version:    "1.0.0",
			constraint: ">2.0.0",
			want:       false,
			wantErr:    false,
		},
		{
			name:       "invalid version",
			version:    "invalid",
			constraint: ">1.0.0",
			want:       false,
			wantErr:    true,
		},
		{
			name:       "invalid constraint",
			version:    "1.0.0",
			constraint: "invalid",
			want:       false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.IsSatisfied(tt.version, tt.constraint)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestModuleVerification(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name    string
		module  *module.Module
		wantErr bool
	}{
		{
			name: "valid module",
			module: &module.Module{
				ID:      "test-module",
				Version: "1.0.0",
				Dependencies: []*module.Dependency{
					{
						Name:     "dep1",
						Version:  "1.0.0",
						Required: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid module version",
			module: &module.Module{
				ID:      "test-module",
				Version: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid dependency version",
			module: &module.Module{
				ID:      "test-module",
				Version: "1.0.0",
				Dependencies: []*module.Dependency{
					{
						Name:     "dep1",
						Version:  "invalid",
						Required: true,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.Verify(tt.module)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVersionMethods(t *testing.T) {
	manager := NewManager()
	v, err := manager.Parse("1.2.3-beta.1+20240101")
	require.NoError(t, err)

	assert.Equal(t, uint64(1), v.Major())
	assert.Equal(t, uint64(2), v.Minor())
	assert.Equal(t, uint64(3), v.Patch())
	assert.Equal(t, "beta.1", v.Prerelease())
	assert.Equal(t, "20240101", v.Metadata())
	assert.True(t, v.IsPrerelease())
	assert.Equal(t, "1.2.3-beta.1+20240101", v.Original())
}
