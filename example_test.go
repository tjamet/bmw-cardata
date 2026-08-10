package bmwcardata

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

// clientID is a placeholder BMW-assigned client ID (a valid UUID) used by the
// runnable documentation examples in this package. The examples exercise the
// real client code paths end to end; TestMain wires package-level fakes — a
// fake HTTP transport serving canned data and a no-op session store — so the
// examples run offline and deterministically.
const clientID = "00000000-0000-0000-0000-000000000000"

// TestMain installs package-level fakes shared by the runnable examples.
//
// Both the auth and CarData clients build their requests with a zero-value
// *http.Client, which uses http.DefaultTransport. Replacing that transport lets
// the examples run against canned responses without any test wiring leaking into
// the example bodies. Replacing the default session store with a no-op store
// forces every example through the full device-code flow (so the PromptURI
// output is deterministic) and keeps the tests from touching the real session
// file on disk.
func TestMain(m *testing.M) {
	http.DefaultTransport = fakeTransport{}
	defaultSessionStore = &NoSessionStore{}
	os.Exit(m.Run())
}

// fakeTransport serves canned BMW auth and CarData responses for the examples.
type fakeTransport struct{}

func (fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, status := fakeResponse(req.Method, req.URL.Path)
	header := http.Header{"Content-Type": []string{"application/json"}}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     header,
		Request:    req,
	}, nil
}

func fakeResponse(method, path string) (body string, status int) {
	switch {
	case method == http.MethodPost && path == "/gcdm/oauth/device/code":
		return `{"device_code":"dev-123","expires_in":600,"interval":1,"user_code":"123456","verification_uri":"https://example.com"}`, http.StatusOK
	case method == http.MethodPost && path == "/gcdm/oauth/token":
		return `{"access_token":"fake-access-token","expires_in":3600,"gcid":"00000000-0000-0000-0000-000000000000","refresh_token":"fake-refresh","scope":"openid cardata:api:read","token_type":"Bearer","id_token":"fake-id-token"}`, http.StatusOK
	case method == http.MethodGet && strings.HasSuffix(path, "/basicData"):
		return `{"vin":"WBA00000000000000","modelName":"X7"}`, http.StatusOK
	case method == http.MethodGet && path == "/customers/vehicles/mappings":
		return `[{"vin":"WBA00000000000000","mappingType":"PRIMARY"},{"vin":"WBA00000000000001","mappingType":"SECONDARY"}]`, http.StatusOK
	case method == http.MethodGet && strings.HasSuffix(path, "/chargingHistory"):
		return `{"data":[{"energyConsumedFromPowerGridKwh":15.4}]}`, http.StatusOK
	case method == http.MethodGet && path == "/customers/containers":
		return `{"containers":[{"containerId":"cid-1","name":"phev-info"}]}`, http.StatusOK
	case method == http.MethodPost && path == "/customers/containers":
		return `{"containerId":"123456","name":"phev-info"}`, http.StatusCreated
	}
	return `{"exveErrorMsg":"unexpected request in example fake: ` + method + ` ` + path + `"}`, http.StatusNotFound
}
