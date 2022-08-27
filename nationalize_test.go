package nationalize

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldCreateNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
}

func TestShouldGetPredictionForName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Rate-Limit-Limit", "1000")
		w.Header().Set("X-Rate-Limit-Remaining", "728")
		w.Header().Set("X-Rate-Reset", "15281")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"michael","country":[{"country_id":"US","probability":0.08986482266532715},{"country_id":"AU","probability":0.05976757527083082},{"country_id":"NZ","probability":0.04666974820852911}]}`))
	}))
	defer server.Close()

	client := NewClient(WithUrl(server.URL))

	result, rateLimit, err := client.Predict("michael")
	assert.Nil(t, err)
	assert.Len(t, result.Country, 3)
	assert.Equal(t, "michael", result.Name)

	country := result.Country[0]
	assert.Equal(t, "US", country.CountryId)
	assert.Equal(t, 0.08986482266532715, country.Probability)

	assert.Equal(t, "1000", rateLimit.Limit)
	assert.Equal(t, "728", rateLimit.Remaining)
	assert.Equal(t, "15281", rateLimit.Reset)
}

func TestShouldGetErrorWhenUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{ "error": "Invalid API key" }`))
	}))
	defer server.Close()

	client := NewClient(WithUrl(server.URL))
	result, rateLimit, err := client.Predict("michael")

	assert.Nil(t, result)
	assert.NotNil(t, rateLimit)
	assert.NotNil(t, err)
}

func TestShouldGetErrorWhenTooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{ "error": "Request limit reached" }`))
	}))
	defer server.Close()

	client := NewClient(WithUrl(server.URL))
	result, rateLimit, err := client.Predict("michael")

	assert.Nil(t, result)
	assert.NotNil(t, rateLimit)
	assert.NotNil(t, err)
}

func TestShouldGetErrorWhenUnprocessable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{ "error": "Too many requests" }`))
	}))
	defer server.Close()

	client := NewClient(WithUrl(server.URL))
	result, rateLimit, err := client.Predict("michael")

	assert.Nil(t, result)
	assert.NotNil(t, rateLimit)
	assert.NotNil(t, err)
}

func TestShouldOverrideDefaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key", r.URL.Query().Get("apikey"))
		assert.Equal(t, "michael", r.URL.Query().Get("name"))

		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{ "error": "testing" }`))
	}))

	defer server.Close()

	client := NewClient(WithUrl(server.URL), WithClient(&http.Client{}), WithApiKey("test-key"))
	assert.NotNil(t, client)

	_, _, err := client.Predict("michael")
	assert.NotNil(t, err)
}

func TestShouldReturnBatchResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		names := r.URL.Query()["name[]"]
		assert.NotNil(t, names)
		assert.Len(t, names, 3)
		w.Write([]byte(`[{"name":"michael","country":[{"country_id":"US","probability":0.08986482266532715},{"country_id":"AU","probability":0.05976757527083082},{"country_id":"NZ","probability":0.04666974820852911}]},{"name":"matthew","country":[{"country_id":"US","probability":0.23117379187666118},{"country_id":"AU","probability":0.21104716641125307},{"country_id":"NZ","probability":0.16699309208074417}]},{"name":"jane","country":[{"country_id":"LR","probability":0.08708128063678007},{"country_id":"KE","probability":0.0763087780134742},{"country_id":"CN","probability":0.04713115986309489}]}]`))
	}))
	defer server.Close()

	client := NewClient(WithUrl(server.URL))
	result, _, err := client.BatchPredict([]string{"michael", "matthew", "jane"})
	assert.Nil(t, err)

	assert.Len(t, result, 3)
}
