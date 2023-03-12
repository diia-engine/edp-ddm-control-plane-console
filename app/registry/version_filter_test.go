package registry

import (
	"ddm-admin-console/service/codebase"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestVersionFilterMake(t *testing.T) {
	versions := []string{">=1.5.2", "==1.3.2", ">1.8.1", "<1.9.2", "<=5.0.0"}
	for _, v := range versions {
		_, err := makeVersionFilter(v)
		assert.NoError(t, err)
	}
}

func TestVersionFilter(t *testing.T) {
	v180, err := version.NewVersion("1.8.0")
	assert.NoError(t, err)

	v192, err := version.NewVersion("1.9.2")
	assert.NoError(t, err)

	v191, err := version.NewVersion("1.9.1")
	assert.NoError(t, err)

	v193, err := version.NewVersion("1.9.3")
	assert.NoError(t, err)

	cbs := []codebase.Codebase{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "v192"},
			Version:    v192,
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "v191"},
			Version:    v191,
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "v180"},
			Version:    v180,
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "v193"},
			Version:    v193,
		},
	}

	f, err := makeVersionFilter("<=1.9.2")
	assert.NoError(t, err)

	filteredCBS, err := f.filterCodebases(cbs)
	assert.NoError(t, err)
	assert.Len(t, filteredCBS, 3)
	assert.Equal(t, filteredCBS[0].Name, "v192")
	assert.Equal(t, filteredCBS[1].Name, "v191")
	assert.Equal(t, filteredCBS[2].Name, "v180")

	f, err = makeVersionFilter("==1.9.3")
	assert.NoError(t, err)

	filteredCBS, err = f.filterCodebases(cbs)
	assert.NoError(t, err)
	assert.Len(t, filteredCBS, 1)
	assert.Equal(t, filteredCBS[0].Name, "v193")

	f, err = makeVersionFilter("")
	assert.NoError(t, err)

	filteredCBS, err = f.filterCodebases(cbs)
	assert.NoError(t, err)
	assert.Len(t, filteredCBS, 4)

}
