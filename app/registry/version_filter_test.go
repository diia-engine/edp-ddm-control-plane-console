package registry

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ddm-admin-console/service/codebase"
)

func TestVersionFilterMake(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "empty version",
			version: "",
			wantErr: require.NoError,
		},
		{
			name:    "invalid version",
			version: "invalid",
			wantErr: require.Error,
		},
		{
			name:    "equal",
			version: "==1.3.2",
			wantErr: require.NoError,
		},
		{
			name:    "bigger",
			version: ">1.8.1",
			wantErr: require.NoError,
		},
		{
			name:    "less",
			version: "<1.9.2",
			wantErr: require.NoError,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := MakeVersionFilter(tt.version)
			tt.wantErr(t, err)
		})
	}
}

func TestVersionFilter_CheckCodebase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		filterVersion   string
		codebaseVersion string
		want            bool
	}{
		{
			name:            "empty filter",
			filterVersion:   "",
			codebaseVersion: "1.1.0",
			want:            true,
		},
		{
			name:            "should be perfectly equal",
			filterVersion:   "==1.1.0",
			codebaseVersion: "1.1.0",
			want:            true,
		},
		{
			name:            "should be equal with minor versions",
			filterVersion:   "==1.1.0",
			codebaseVersion: "1.1.0.123.456",
			want:            true,
		},
		{
			name:            "should be equal with rc",
			filterVersion:   "==1.1.0",
			codebaseVersion: "1.1.0-rc123",
			want:            true,
		},
		{
			name:            "should not be equal",
			filterVersion:   "==1.1.0",
			codebaseVersion: "1.2.0",
			want:            false,
		},
		{
			name:            "should be bigger",
			filterVersion:   ">1.1.0",
			codebaseVersion: "1.2.0",
			want:            true,
		},
		{
			name:            "should not be bigger (less)",
			filterVersion:   ">1.1.0",
			codebaseVersion: "1.0.0",
			want:            false,
		},
		{
			name:            "should not be bigger (equal)",
			filterVersion:   ">1.1.0",
			codebaseVersion: "1.1.0.123.456",
			want:            false,
		},
		{
			name:            "should be less",
			filterVersion:   "<1.1.0",
			codebaseVersion: "1.0.0",
			want:            true,
		},
		{
			name:            "should not be less (bigger)",
			filterVersion:   "<1.1.0",
			codebaseVersion: "1.2.0",
			want:            false,
		},
		{
			name:            "should not be less (equal)",
			filterVersion:   "<1.1.0",
			codebaseVersion: "1.1.0.123",
			want:            false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filter, err := MakeVersionFilter(tt.filterVersion)
			require.NoError(t, err)

			ver, err := version.NewVersion(tt.codebaseVersion)
			require.NoError(t, err)

			cb := &codebase.Codebase{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.codebaseVersion,
				},
				Version: ver,
			}

			passesFilter := filter.CheckCodebase(cb)
			require.Equal(t, tt.want, passesFilter)
		})
	}

}
