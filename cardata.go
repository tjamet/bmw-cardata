package bmwcardata

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/tjamet/bmw-cardata/cardataapi"
)

// GetBasicData gets the basic data for a given VIN
// See https://bmw-cardata.bmwgroup.com/customer/public/api-specification#operations-Vehicles-getBasicData
func (c *Client) GetBasicData(ctx context.Context, vin string) (*cardataapi.VehicleDto, error) {
	resp, err := c.carDataAPI.GetBasicData(ctx, vin, &cardataapi.GetBasicDataParams{XVersion: "v1"})
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		data := cardataapi.VehicleDto{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	default:
		data := cardataapi.CarDataError{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return nil, &data
	}
}

// GetMappings lists all the existing mappings (i.e. car VINs) that are available in the BMW CarData API
// See https://bmw-cardata.bmwgroup.com/customer/public/api-specification#operations-Vehicles-getMappings
func (c *Client) GetMappings(ctx context.Context) ([]cardataapi.VehicleMappingDto, error) {
	resp, err := c.carDataAPI.GetMappings(ctx, &cardataapi.GetMappingsParams{XVersion: "v1"})
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		data := []cardataapi.VehicleMappingDto{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return data, nil
	default:
		data := cardataapi.CarDataError{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return nil, &data
	}
}

type GetChargingHistoryParamsOption func(*cardataapi.GetChargingHistoryParams)

func WithChargingHistoryNextToken(token string) GetChargingHistoryParamsOption {
	return func(params *cardataapi.GetChargingHistoryParams) {
		params.NextToken = &token
	}
}

// GetChargingHistory gets the charging history for a given VIN
// See https://bmw-cardata.bmwgroup.com/customer/public/api-specification#operations-Vehicles-getChargingHistory
func (c *Client) GetChargingHistory(ctx context.Context, vin string, from, to time.Time, options ...GetChargingHistoryParamsOption) (*cardataapi.ChargingHistoryResponseDto, error) {
	params := &cardataapi.GetChargingHistoryParams{XVersion: "v1", From: from, To: to}
	for _, option := range options {
		option(params)
	}
	resp, err := c.carDataAPI.GetChargingHistory(ctx, vin, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		data := cardataapi.ChargingHistoryResponseDto{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	default:
		data := cardataapi.CarDataError{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return nil, &data
	}
}

type Image struct {
	Data        []byte
	ContentType string
}

// GetImage gets the image for a given VIN
// See https://bmw-cardata.bmwgroup.com/customer/public/api-specification#operations-Vehicles-getImage
func (c *Client) GetImage(ctx context.Context, vin string) (*Image, error) {
	resp, err := c.carDataAPI.GetImage(ctx, vin, &cardataapi.GetImageParams{XVersion: "v1"})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return &Image{Data: data, ContentType: resp.Header.Get("Content-Type")}, nil
	default:
		data := cardataapi.CarDataError{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return nil, &data
	}
}

type GetLocationBasedChargingSettingsParamsOption func(*cardataapi.GetLocationBasedChargingSettingsParams)

func WithLocationBasedChargingSettingsNextToken(token string) GetLocationBasedChargingSettingsParamsOption {
	return func(params *cardataapi.GetLocationBasedChargingSettingsParams) {
		params.NextToken = &token
	}
}

// GetLocationBasedChargingSettings gets the location based charging settings for a given VIN
// See https://bmw-cardata.bmwgroup.com/customer/public/api-specification#operations-Vehicles-getLocationBasedChargingSettings
func (c *Client) GetLocationBasedChargingSettings(ctx context.Context, vin string, options ...GetLocationBasedChargingSettingsParamsOption) (*cardataapi.LocationBasedChargingSettingsDto, error) {
	params := &cardataapi.GetLocationBasedChargingSettingsParams{XVersion: "v1"}
	for _, option := range options {
		option(params)
	}
	resp, err := c.carDataAPI.GetLocationBasedChargingSettings(ctx, vin, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		data := cardataapi.LocationBasedChargingSettingsDto{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	default:
		data := cardataapi.CarDataError{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return nil, &data
	}
}

// GetSmartMaintenanceTyreDiagnosis gets the smart maintenance tyre diagnosis for a given VIN
// See https://bmw-cardata.bmwgroup.com/customer/public/api-specification#operations-Vehicles-getSmartMaintenanceTyreDiagnosis
func (c *Client) GetSmartMaintenanceTyreDiagnosis(ctx context.Context, vin string) (*cardataapi.SmartMaintenanceTyreDiagnosisDto, error) {
	resp, err := c.carDataAPI.GetSmartMaintenanceTyreDiagnosis(ctx, vin, &cardataapi.GetSmartMaintenanceTyreDiagnosisParams{XVersion: "v1"})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		data := cardataapi.SmartMaintenanceTyreDiagnosisDto{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	default:
		data := cardataapi.CarDataError{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return nil, &data
	}
}

// GetTelematicData gets the telematic data for a given VIN and container ID
// See https://bmw-cardata.bmwgroup.com/customer/public/api-specification#operations-Vehicles-getTelematicData
func (c *Client) GetTelematicData(ctx context.Context, vin, containerID string) (*cardataapi.ExVeTelematicDataResponseDto, error) {
	resp, err := c.carDataAPI.GetTelematicData(ctx, vin, &cardataapi.GetTelematicDataParams{XVersion: "v1", ContainerId: containerID})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		data := cardataapi.ExVeTelematicDataResponseDto{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	default:
		data := cardataapi.CarDataError{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return nil, &data
	}
}
