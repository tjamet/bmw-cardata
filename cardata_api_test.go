package bmwcardata

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/tjamet/bmw-cardata/cardataapi"
)

// mockCardataClient implements cardataapi.ClientInterface for tests
type mockCardataClient struct {
	ListContainersFunc                   func(ctx context.Context, params *cardataapi.ListContainersParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
	CreateContainerFunc                  func(ctx context.Context, body cardataapi.CreateContainerJSONRequestBody, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
	CreateContainerWithBodyFunc          func(ctx context.Context, contentType string, body io.Reader, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
	DeleteContainerFunc                  func(ctx context.Context, containerId string, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
	GetContainerDetailsFunc              func(ctx context.Context, containerId string, params *cardataapi.GetContainerDetailsParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
	GetMappingsFunc                      func(ctx context.Context, params *cardataapi.GetMappingsParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
	GetBasicDataFunc                     func(ctx context.Context, vin string, params *cardataapi.GetBasicDataParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
	GetChargingHistoryFunc               func(ctx context.Context, vin string, params *cardataapi.GetChargingHistoryParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
	GetImageFunc                         func(ctx context.Context, vin string, params *cardataapi.GetImageParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
	GetLocationBasedChargingSettingsFunc func(ctx context.Context, vin string, params *cardataapi.GetLocationBasedChargingSettingsParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
	GetSmartMaintenanceTyreDiagnosisFunc func(ctx context.Context, vin string, params *cardataapi.GetSmartMaintenanceTyreDiagnosisParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
	GetTelematicDataFunc                 func(ctx context.Context, vin string, params *cardataapi.GetTelematicDataParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error)
}

func (m *mockCardataClient) ListContainers(ctx context.Context, params *cardataapi.ListContainersParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.ListContainersFunc != nil {
		return m.ListContainersFunc(ctx, params, reqEditors...)
	}
	return nil, nil
}

func (m *mockCardataClient) CreateContainerWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.CreateContainerWithBodyFunc != nil {
		return m.CreateContainerWithBodyFunc(ctx, contentType, body, reqEditors...)
	}
	return nil, nil
}

func (m *mockCardataClient) CreateContainer(ctx context.Context, body cardataapi.CreateContainerJSONRequestBody, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.CreateContainerFunc != nil {
		return m.CreateContainerFunc(ctx, body, reqEditors...)
	}
	return nil, nil
}

func (m *mockCardataClient) DeleteContainer(ctx context.Context, containerId string, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.DeleteContainerFunc != nil {
		return m.DeleteContainerFunc(ctx, containerId, reqEditors...)
	}
	return nil, nil
}

func (m *mockCardataClient) GetContainerDetails(ctx context.Context, containerId string, params *cardataapi.GetContainerDetailsParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.GetContainerDetailsFunc != nil {
		return m.GetContainerDetailsFunc(ctx, containerId, params, reqEditors...)
	}
	return nil, nil
}

func (m *mockCardataClient) GetMappings(ctx context.Context, params *cardataapi.GetMappingsParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.GetMappingsFunc != nil {
		return m.GetMappingsFunc(ctx, params, reqEditors...)
	}
	return nil, nil
}

func (m *mockCardataClient) GetBasicData(ctx context.Context, vin string, params *cardataapi.GetBasicDataParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.GetBasicDataFunc != nil {
		return m.GetBasicDataFunc(ctx, vin, params, reqEditors...)
	}
	return nil, nil
}

func (m *mockCardataClient) GetChargingHistory(ctx context.Context, vin string, params *cardataapi.GetChargingHistoryParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.GetChargingHistoryFunc != nil {
		return m.GetChargingHistoryFunc(ctx, vin, params, reqEditors...)
	}
	return nil, nil
}

func (m *mockCardataClient) GetImage(ctx context.Context, vin string, params *cardataapi.GetImageParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.GetImageFunc != nil {
		return m.GetImageFunc(ctx, vin, params, reqEditors...)
	}
	return nil, nil
}

func (m *mockCardataClient) GetLocationBasedChargingSettings(ctx context.Context, vin string, params *cardataapi.GetLocationBasedChargingSettingsParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.GetLocationBasedChargingSettingsFunc != nil {
		return m.GetLocationBasedChargingSettingsFunc(ctx, vin, params, reqEditors...)
	}
	return nil, nil
}

func (m *mockCardataClient) GetSmartMaintenanceTyreDiagnosis(ctx context.Context, vin string, params *cardataapi.GetSmartMaintenanceTyreDiagnosisParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.GetSmartMaintenanceTyreDiagnosisFunc != nil {
		return m.GetSmartMaintenanceTyreDiagnosisFunc(ctx, vin, params, reqEditors...)
	}
	return nil, nil
}

func (m *mockCardataClient) GetTelematicData(ctx context.Context, vin string, params *cardataapi.GetTelematicDataParams, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
	if m.GetTelematicDataFunc != nil {
		return m.GetTelematicDataFunc(ctx, vin, params, reqEditors...)
	}
	return nil, nil
}

func jsonResponse(status int, v interface{}, headers map[string]string) *http.Response {
	data, _ := json.Marshal(v)
	resp := &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(data)),
		Header:     make(http.Header),
	}
	for k, val := range headers {
		resp.Header.Set(k, val)
	}
	return resp
}

func bytesResponse(status int, data []byte, headers map[string]string) *http.Response {
	resp := &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(data)),
		Header:     make(http.Header),
	}
	for k, val := range headers {
		resp.Header.Set(k, val)
	}
	return resp
}

func TestGetBasicData_Success(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetBasicDataFunc: func(ctx context.Context, vin string, params *cardataapi.GetBasicDataParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			if params == nil || params.XVersion != "v1" {
				t.Fatalf("expected XVersion v1, got %#v", params)
			}
			vinVal := "VIN123"
			return jsonResponse(http.StatusOK, cardataapi.VehicleDto{Vin: &vinVal}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	data, err := c.GetBasicData(ctx, "VIN123")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if data == nil || data.Vin == nil || *data.Vin != "VIN123" {
		t.Fatalf("unexpected data: %#v", data)
	}
}

func TestGetBasicData_Error(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetBasicDataFunc: func(ctx context.Context, vin string, params *cardataapi.GetBasicDataParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			msg := "bad request"
			return jsonResponse(http.StatusBadRequest, cardataapi.CarDataError{ExveErrorMsg: &msg}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.GetBasicData(ctx, "VIN123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*cardataapi.CarDataError); !ok {
		t.Fatalf("expected CarDataError, got %T", err)
	}
}

func TestGetBasicData_DecodeFailureOnSuccess(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetBasicDataFunc: func(ctx context.Context, vin string, params *cardataapi.GetBasicDataParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return bytesResponse(http.StatusOK, []byte("not-json"), nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.GetBasicData(ctx, "VIN123")
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestGetMappings_Success(t *testing.T) {
	ctx := context.Background()
	mapping := cardataapi.VehicleMappingDto{}
	mock := &mockCardataClient{
		GetMappingsFunc: func(ctx context.Context, params *cardataapi.GetMappingsParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			if params == nil || params.XVersion != "v1" {
				t.Fatalf("expected XVersion v1, got %#v", params)
			}
			return jsonResponse(http.StatusOK, []cardataapi.VehicleMappingDto{mapping}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	list, err := c.GetMappings(ctx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 mapping, got %d", len(list))
	}
}

func TestGetMappings_Error(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetMappingsFunc: func(ctx context.Context, params *cardataapi.GetMappingsParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			msg := "bad mappings"
			return jsonResponse(http.StatusForbidden, cardataapi.CarDataError{ExveErrorMsg: &msg}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.GetMappings(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*cardataapi.CarDataError); !ok {
		t.Fatalf("expected CarDataError, got %T", err)
	}
}

func TestGetChargingHistory_WithNextToken(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetChargingHistoryFunc: func(ctx context.Context, vin string, params *cardataapi.GetChargingHistoryParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			if params == nil || params.XVersion != "v1" {
				t.Fatalf("expected XVersion v1, got %#v", params)
			}
			if params.NextToken == nil || *params.NextToken != "next123" {
				t.Fatalf("expected NextToken next123, got %#v", params.NextToken)
			}
			return jsonResponse(http.StatusOK, cardataapi.ChargingHistoryResponseDto{}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	from := time.Now().Add(-time.Hour)
	to := time.Now()
	resp, err := c.GetChargingHistory(ctx, "VIN", from, to, WithChargingHistoryNextToken("next123"))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestGetChargingHistory_ErrorNon200(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetChargingHistoryFunc: func(ctx context.Context, vin string, params *cardataapi.GetChargingHistoryParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			msg := "oops"
			return jsonResponse(http.StatusBadRequest, cardataapi.CarDataError{ExveErrorMsg: &msg}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	from := time.Now().Add(-time.Hour)
	to := time.Now()
	_, err := c.GetChargingHistory(ctx, "VIN", from, to)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*cardataapi.CarDataError); !ok {
		t.Fatalf("expected CarDataError, got %T", err)
	}
}

func TestGetChargingHistory_DecodeFailureOnSuccess(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetChargingHistoryFunc: func(ctx context.Context, vin string, params *cardataapi.GetChargingHistoryParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return bytesResponse(http.StatusOK, []byte("not-json"), nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	from := time.Now().Add(-time.Hour)
	to := time.Now()
	_, err := c.GetChargingHistory(ctx, "VIN", from, to)
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestGetImage_Success(t *testing.T) {
	ctx := context.Background()
	dataBytes := []byte{1, 2, 3}
	mock := &mockCardataClient{
		GetImageFunc: func(ctx context.Context, vin string, params *cardataapi.GetImageParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			if params == nil || params.XVersion != "v1" {
				t.Fatalf("expected XVersion v1, got %#v", params)
			}
			return bytesResponse(http.StatusOK, dataBytes, map[string]string{"Content-Type": "image/jpeg"}), nil
		},
	}
	c := &Client{carDataAPI: mock}
	img, err := c.GetImage(ctx, "VIN")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if img == nil || img.ContentType != "image/jpeg" || len(img.Data) != 3 {
		t.Fatalf("unexpected image: %#v", img)
	}
}

func TestGetImage_ErrorCarData(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetImageFunc: func(ctx context.Context, vin string, params *cardataapi.GetImageParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			msg := "not found"
			return jsonResponse(http.StatusNotFound, cardataapi.CarDataError{ExveErrorMsg: &msg}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.GetImage(ctx, "VIN")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*cardataapi.CarDataError); !ok {
		t.Fatalf("expected CarDataError, got %T", err)
	}
}

func TestGetImage_ErrorDecodeFailure(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetImageFunc: func(ctx context.Context, vin string, params *cardataapi.GetImageParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return bytesResponse(http.StatusBadRequest, []byte("not-json"), nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.GetImage(ctx, "VIN")
	if err == nil {
		t.Fatal("expected decode error on error body, got nil")
	}
}

func TestGetLocationBasedChargingSettings_WithNextToken(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetLocationBasedChargingSettingsFunc: func(ctx context.Context, vin string, params *cardataapi.GetLocationBasedChargingSettingsParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			if params == nil || params.XVersion != "v1" {
				t.Fatalf("expected XVersion v1, got %#v", params)
			}
			if params.NextToken == nil || *params.NextToken != "n2" {
				t.Fatalf("expected NextToken n2, got %#v", params.NextToken)
			}
			return jsonResponse(http.StatusOK, cardataapi.LocationBasedChargingSettingsDto{}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	resp, err := c.GetLocationBasedChargingSettings(ctx, "VIN", WithLocationBasedChargingSettingsNextToken("n2"))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestGetLocationBasedChargingSettings_Error(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetLocationBasedChargingSettingsFunc: func(ctx context.Context, vin string, params *cardataapi.GetLocationBasedChargingSettingsParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			msg := "no perms"
			return jsonResponse(http.StatusForbidden, cardataapi.CarDataError{ExveErrorMsg: &msg}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.GetLocationBasedChargingSettings(ctx, "VIN")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*cardataapi.CarDataError); !ok {
		t.Fatalf("expected CarDataError, got %T", err)
	}
}

func TestGetSmartMaintenanceTyreDiagnosis_Success(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetSmartMaintenanceTyreDiagnosisFunc: func(ctx context.Context, vin string, params *cardataapi.GetSmartMaintenanceTyreDiagnosisParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			if params == nil || params.XVersion != "v1" {
				t.Fatalf("expected XVersion v1, got %#v", params)
			}
			return jsonResponse(http.StatusOK, cardataapi.SmartMaintenanceTyreDiagnosisDto{}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	resp, err := c.GetSmartMaintenanceTyreDiagnosis(ctx, "VIN")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestGetSmartMaintenanceTyreDiagnosis_Error(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetSmartMaintenanceTyreDiagnosisFunc: func(ctx context.Context, vin string, params *cardataapi.GetSmartMaintenanceTyreDiagnosisParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			msg := "denied"
			return jsonResponse(http.StatusUnauthorized, cardataapi.CarDataError{ExveErrorMsg: &msg}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.GetSmartMaintenanceTyreDiagnosis(ctx, "VIN")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*cardataapi.CarDataError); !ok {
		t.Fatalf("expected CarDataError, got %T", err)
	}
}

func TestGetTelematicData_Success(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetTelematicDataFunc: func(ctx context.Context, vin string, params *cardataapi.GetTelematicDataParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			if params == nil || params.XVersion != "v1" || params.ContainerId == "" {
				t.Fatalf("expected XVersion v1 and container id, got %#v", params)
			}
			return jsonResponse(http.StatusOK, cardataapi.ExVeTelematicDataResponseDto{}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	resp, err := c.GetTelematicData(ctx, "VIN", "CID")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestGetTelematicData_Error(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetTelematicDataFunc: func(ctx context.Context, vin string, params *cardataapi.GetTelematicDataParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			msg := "oops"
			return jsonResponse(http.StatusBadRequest, cardataapi.CarDataError{ExveErrorMsg: &msg}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.GetTelematicData(ctx, "VIN", "CID")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*cardataapi.CarDataError); !ok {
		t.Fatalf("expected CarDataError, got %T", err)
	}
}

func TestListContainers_Success(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		ListContainersFunc: func(ctx context.Context, params *cardataapi.ListContainersParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			if params == nil || params.XVersion != "v1" {
				t.Fatalf("expected XVersion v1, got %#v", params)
			}
			return jsonResponse(http.StatusOK, cardataapi.ContainerListDto{}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	resp, err := c.ListContainers(ctx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestListContainers_Error(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		ListContainersFunc: func(ctx context.Context, params *cardataapi.ListContainersParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			msg := "oops"
			return jsonResponse(http.StatusBadRequest, cardataapi.CarDataError{ExveErrorMsg: &msg}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.ListContainers(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*cardataapi.CarDataError); !ok {
		t.Fatalf("expected CarDataError, got %T", err)
	}
}

func TestGetContainerDetails_Success(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetContainerDetailsFunc: func(ctx context.Context, containerId string, params *cardataapi.GetContainerDetailsParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			if params == nil || params.XVersion != "v1" || containerId == "" {
				t.Fatalf("expected XVersion v1 and containerId, got params=%#v id=%s", params, containerId)
			}
			return jsonResponse(http.StatusOK, cardataapi.ContainerDetailsDto{}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	resp, err := c.GetContainerDetails(ctx, "CID")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestGetContainerDetails_Error(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetContainerDetailsFunc: func(ctx context.Context, containerId string, params *cardataapi.GetContainerDetailsParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			msg := "oops"
			return jsonResponse(http.StatusBadRequest, cardataapi.CarDataError{ExveErrorMsg: &msg}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.GetContainerDetails(ctx, "CID")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*cardataapi.CarDataError); !ok {
		t.Fatalf("expected CarDataError, got %T", err)
	}
}

func TestCreateContainer_SetsHeaderAndBody(t *testing.T) {
	ctx := context.Background()
	capturedHeader := ""
	var capturedBody cardataapi.CreateContainerJSONRequestBody
	mock := &mockCardataClient{
		CreateContainerFunc: func(ctx context.Context, body cardataapi.CreateContainerJSONRequestBody, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
			capturedBody = body
			// Simulate request editors application to capture header
			req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
			for _, ed := range reqEditors {
				_ = ed(ctx, req)
			}
			capturedHeader = req.Header.Get("X-Version")
			return jsonResponse(http.StatusOK, cardataapi.CreateContainerResponse{}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	containers := []Descriptor{{ID: "id1"}, {ID: "id2"}}
	resp, err := c.CreateContainer(ctx, "name", "purpose", containers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if capturedHeader != "v1" {
		t.Fatalf("expected X-Version header v1, got %q", capturedHeader)
	}
	if capturedBody.Name == nil || *capturedBody.Name != "name" {
		t.Fatalf("expected name in body, got %#v", capturedBody.Name)
	}
	if capturedBody.Purpose == nil || *capturedBody.Purpose != "purpose" {
		t.Fatalf("expected purpose in body, got %#v", capturedBody.Purpose)
	}
	if capturedBody.TechnicalDescriptors == nil || len(*capturedBody.TechnicalDescriptors) != 2 {
		t.Fatalf("expected 2 technical descriptors, got %#v", capturedBody.TechnicalDescriptors)
	}
}

func TestCreateContainer_ErrorNon200(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		CreateContainerFunc: func(ctx context.Context, body cardataapi.CreateContainerJSONRequestBody, reqEditors ...cardataapi.RequestEditorFn) (*http.Response, error) {
			msg := "bad"
			return jsonResponse(http.StatusBadRequest, cardataapi.CarDataError{ExveErrorMsg: &msg}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.CreateContainer(ctx, "name", "purpose", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*cardataapi.CarDataError); !ok {
		t.Fatalf("expected CarDataError, got %T", err)
	}
}

func TestDeleteContainer_Success(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		DeleteContainerFunc: func(ctx context.Context, containerId string, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			if containerId != "CID" {
				t.Fatalf("expected containerId CID, got %s", containerId)
			}
			return jsonResponse(http.StatusOK, cardataapi.DeleteContainerResponse{}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	resp, err := c.DeleteContainer(ctx, "CID")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestDeleteContainer_Error(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		DeleteContainerFunc: func(ctx context.Context, containerId string, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			msg := "bad"
			return jsonResponse(http.StatusBadRequest, cardataapi.CarDataError{ExveErrorMsg: &msg}, nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.DeleteContainer(ctx, "CID")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*cardataapi.CarDataError); !ok {
		t.Fatalf("expected CarDataError, got %T", err)
	}
}

func TestUnderlyingClientErrorPropagation(t *testing.T) {
	ctx := context.Background()
	underlyingErr := errors.New("network down")
	mock := &mockCardataClient{
		ListContainersFunc: func(ctx context.Context, params *cardataapi.ListContainersParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return nil, underlyingErr
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.ListContainers(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, underlyingErr) {
		t.Fatalf("expected underlying error, got %v", err)
	}
}

func ExampleClient_GetBasicData() {
	client := Must(NewClient(
		WithAuthenticator(
			Must(NewAuthenticator(
				WithClientID(clientID),
				WithPromptURI(func(verificationURI, userCode, verificationURIComplete string) {
					fmt.Printf("Open %s and enter code %s\n", verificationURI, userCode)
					fmt.Printf("Direct link: %s\n", verificationURIComplete)
				}),
			)),
		)),
	)

	ctx := context.Background()
	vin := "WBA00000000000000"
	vehicle, err := client.GetBasicData(ctx, vin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Vehicle: %+v\n", *vehicle.ModelName)
	// Output: Open https://example.com and enter code 123456
	// Direct link: https://example.com?code=123456
	// Vehicle: X7
}

func ExampleClient_GetMappings() {
	client := Must(NewClient(
		WithAuthenticator(
			Must(NewAuthenticator(
				WithClientID(clientID),
				WithPromptURI(func(verificationURI, userCode, verificationURIComplete string) {
					fmt.Printf("Open %s and enter code %s\n", verificationURI, userCode)
					fmt.Printf("Direct link: %s\n", verificationURIComplete)
				}),
			)),
		)),
	)

	ctx := context.Background()
	mappings, err := client.GetMappings(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Mappings: %d\n", len(mappings))
	// Output: Open https://example.com and enter code 123456
	// Direct link: https://example.com?code=123456
	// Mappings: 2
}

func ExampleClient_GetChargingHistory() {
	client := Must(NewClient(
		WithAuthenticator(
			Must(NewAuthenticator(
				WithClientID(clientID),
				WithPromptURI(func(verificationURI, userCode, verificationURIComplete string) {
					fmt.Printf("Open %s and enter code %s\n", verificationURI, userCode)
					fmt.Printf("Direct link: %s\n", verificationURIComplete)
				}),
			)),
		)),
	)

	ctx := context.Background()
	vin := "WBA00000000000000"
	chargingHistory, err := client.GetChargingHistory(ctx, vin, time.Now().Add(-1*24*time.Hour), time.Now())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Charging history: %+v\n", *chargingHistory.Data[0].EnergyConsumedFromPowerGridKwh)
	// Output: Open https://example.com and enter code 123456
	// Direct link: https://example.com?code=123456
	// Charging history: 15.4
}
