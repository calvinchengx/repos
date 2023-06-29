package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRepositories(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prepare mock response
		responseJSON := `[
			{"name": "repo1", "clone_url": "https://github.com/org/repo1.git"},
			{"name": "repo2", "clone_url": "https://github.com/org/repo2.git"}
		]`

		// Write mock response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseJSON))
	}))
	defer mockServer.Close()

	// Perform the test
	fmt.Println(mockServer.URL)
	repositories, err := GetRepositories(mockServer.URL, "org", "", "accessToken")
	assert.NoError(t, err)
	assert.NotNil(t, repositories)
	assert.Equal(t, 2, len(repositories))
	assert.Equal(t, "repo1", repositories[0].Name)
	assert.Equal(t, "https://github.com/org/repo1.git", repositories[0].CloneURL)
	assert.Equal(t, "repo2", repositories[1].Name)
	assert.Equal(t, "https://github.com/org/repo2.git", repositories[1].CloneURL)
}
