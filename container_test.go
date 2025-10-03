package bmwcardata

import (
	"context"
	"net/http"
	"testing"

	"github.com/tjamet/bmw-cardata/cardataapi"
)

// Container API decode-failure corner cases

func TestListContainers_DecodeFailureOnSuccess(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		ListContainersFunc: func(ctx context.Context, params *cardataapi.ListContainersParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return bytesResponse(http.StatusOK, []byte("not-json"), nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.ListContainers(ctx)
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestListContainers_DecodeFailureOnErrorBody(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		ListContainersFunc: func(ctx context.Context, params *cardataapi.ListContainersParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return bytesResponse(http.StatusBadRequest, []byte("not-json"), nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.ListContainers(ctx)
	if err == nil {
		t.Fatal("expected decode error on error body, got nil")
	}
}

func TestGetContainerDetails_DecodeFailureOnSuccess(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetContainerDetailsFunc: func(ctx context.Context, containerId string, params *cardataapi.GetContainerDetailsParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return bytesResponse(http.StatusOK, []byte("not-json"), nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.GetContainerDetails(ctx, "CID")
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestGetContainerDetails_DecodeFailureOnErrorBody(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		GetContainerDetailsFunc: func(ctx context.Context, containerId string, params *cardataapi.GetContainerDetailsParams, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return bytesResponse(http.StatusBadRequest, []byte("not-json"), nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.GetContainerDetails(ctx, "CID")
	if err == nil {
		t.Fatal("expected decode error on error body, got nil")
	}
}

func TestCreateContainer_DecodeFailureOnSuccess(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		CreateContainerFunc: func(ctx context.Context, body cardataapi.CreateContainerJSONRequestBody, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return bytesResponse(http.StatusOK, []byte("not-json"), nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.CreateContainer(ctx, "name", "purpose", nil)
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestCreateContainer_DecodeFailureOnErrorBody(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		CreateContainerFunc: func(ctx context.Context, body cardataapi.CreateContainerJSONRequestBody, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return bytesResponse(http.StatusBadRequest, []byte("not-json"), nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.CreateContainer(ctx, "name", "purpose", nil)
	if err == nil {
		t.Fatal("expected decode error on error body, got nil")
	}
}

func TestDeleteContainer_DecodeFailureOnSuccess(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		DeleteContainerFunc: func(ctx context.Context, containerId string, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return bytesResponse(http.StatusOK, []byte("not-json"), nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.DeleteContainer(ctx, "CID")
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestDeleteContainer_DecodeFailureOnErrorBody(t *testing.T) {
	ctx := context.Background()
	mock := &mockCardataClient{
		DeleteContainerFunc: func(ctx context.Context, containerId string, _ ...cardataapi.RequestEditorFn) (*http.Response, error) {
			return bytesResponse(http.StatusBadRequest, []byte("not-json"), nil), nil
		},
	}
	c := &Client{carDataAPI: mock}
	_, err := c.DeleteContainer(ctx, "CID")
	if err == nil {
		t.Fatal("expected decode error on error body, got nil")
	}
}

// Matcher helpers and descriptor filtering

func TestDescriptorMatchers_Basic(t *testing.T) {
	d := Descriptor{
		ID:           "id1",
		Brand:        []Brand{BrandBMW},
		VehicleTypes: []VehicleType{VehicleTypeBEV},
		Category:     "BASIC_DATA",
	}

	if !MatchID("id1").Match(d) {
		t.Fatal("MatchID should match same id")
	}
	if MatchID("other").Match(d) {
		t.Fatal("MatchID should not match different id")
	}

	if !MatchBrand(BrandBMW).Match(d) {
		t.Fatal("MatchBrand should match BrandBMW")
	}
	if MatchBrand(Brand("OTHER")).Match(d) {
		t.Fatal("MatchBrand should not match OTHER")
	}

	if !MatchVehicleType(VehicleTypeBEV).Match(d) {
		t.Fatal("MatchVehicleType should match BEV")
	}
	if MatchVehicleType(VehicleTypeICE).Match(d) {
		t.Fatal("MatchVehicleType should not match ICE")
	}

	if !MatchCategory("BASIC_DATA").Match(d) {
		t.Fatal("MatchCategory should match BASIC_DATA")
	}
	if MatchCategory("NOPE").Match(d) {
		t.Fatal("MatchCategory should not match NOPE")
	}
}

func TestDescriptorMatchers_Combinators(t *testing.T) {
	d := Descriptor{
		ID:           "id1",
		Brand:        []Brand{BrandBMW},
		VehicleTypes: []VehicleType{VehicleTypeBEV},
		Category:     "BASIC_DATA",
	}

	if !MatchAll(MatchID("id1"), MatchBrand(BrandBMW)).Match(d) {
		t.Fatal("MatchAll should be true when all matchers match")
	}
	if MatchAll(MatchID("id1"), MatchCategory("NOPE")).Match(d) {
		t.Fatal("MatchAll should be false when at least one does not match")
	}
	if !MatchAll().Match(d) {
		t.Fatal("MatchAll with no matchers should be true")
	}

	if !MatchAny(MatchID("nope"), MatchBrand(BrandBMW)).Match(d) {
		t.Fatal("MatchAny should be true if any matcher matches")
	}
	if MatchAny().Match(d) {
		t.Fatal("MatchAny with no matchers should be false")
	}

	if !MatchNone(MatchID("nope"), MatchCategory("NOPE")).Match(d) {
		t.Fatal("MatchNone should be true when none match")
	}
	if MatchNone(MatchCategory("BASIC_DATA")).Match(d) {
		t.Fatal("MatchNone should be false when one matches")
	}
}

func TestFindDescriptors(t *testing.T) {
	// Always-true matcher should return at least one descriptor from the generated catalogue
	results := FindDescriptors(DescriptorMatcherFunc(func(container Descriptor) bool { return true }))
	if len(results) == 0 {
		t.Fatal("expected some descriptors to be returned")
	}

	// Always-false matcher should return an empty slice
	results = FindDescriptors(DescriptorMatcherFunc(func(container Descriptor) bool { return false }))
	if len(results) != 0 {
		t.Fatalf("expected 0 descriptors, got %d", len(results))
	}
}
