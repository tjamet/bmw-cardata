// Package cardataapi is a small, hand-written HTTP client for the BMW CarData
// customer API.
//
// It is written directly against the official OpenAPI specification published by
// BMW at
// https://bmw-cardata.bmwgroup.com/customer/public/assets/swagger/swagger-customer-api-v1.json
// (mirrored for reference in ../swagger/customer-api.json). The client
// deliberately avoids code generation: the surface is small enough to maintain
// by hand and much easier to read and test than generated boilerplate.
//
// The client exposes one method per API operation. Each method returns the raw
// *http.Response so that callers (the top-level bmwcardata package) stay in
// control of status-code handling and body decoding.
package cardataapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Defines values for ChargingSessionDtoMileageUnits.
const (
	MileageUnitsKM ChargingSessionDtoMileageUnits = "MileageUnits.KM"
	MileageUnitsMI ChargingSessionDtoMileageUnits = "MileageUnits.MI"
)

// Defines values for ContainerDetailsDtoState.
const (
	ContainerDetailsDtoStateACTIVE  ContainerDetailsDtoState = "ACTIVE"
	ContainerDetailsDtoStateDELETED ContainerDetailsDtoState = "DELETED"
)

// Defines values for ContainerDtoState.
const (
	ContainerDtoStateACTIVE  ContainerDtoState = "ACTIVE"
	ContainerDtoStateDELETED ContainerDtoState = "DELETED"
)

// Defines values for TyreDefectDtoStatusColor.
const (
	TyreDefectDtoStatusColorGREEN  TyreDefectDtoStatusColor = "GREEN"
	TyreDefectDtoStatusColorGREY   TyreDefectDtoStatusColor = "GREY"
	TyreDefectDtoStatusColorRED    TyreDefectDtoStatusColor = "RED"
	TyreDefectDtoStatusColorYELLOW TyreDefectDtoStatusColor = "YELLOW"
)

// Defines values for TyreProductionDateDtoStatusColor.
const (
	TyreProductionDateDtoStatusColorGREEN  TyreProductionDateDtoStatusColor = "GREEN"
	TyreProductionDateDtoStatusColorGREY   TyreProductionDateDtoStatusColor = "GREY"
	TyreProductionDateDtoStatusColorRED    TyreProductionDateDtoStatusColor = "RED"
	TyreProductionDateDtoStatusColorYELLOW TyreProductionDateDtoStatusColor = "YELLOW"
)

// Defines values for TyreWearDtoStatusColor.
const (
	TyreWearDtoStatusColorGREEN  TyreWearDtoStatusColor = "GREEN"
	TyreWearDtoStatusColorGREY   TyreWearDtoStatusColor = "GREY"
	TyreWearDtoStatusColorRED    TyreWearDtoStatusColor = "RED"
	TyreWearDtoStatusColorYELLOW TyreWearDtoStatusColor = "YELLOW"
)

// Defines values for TyreWearDtoUnit.
const (
	TyreWearDtoUnitKILOMETER TyreWearDtoUnit = "KILOMETER"
	TyreWearDtoUnitMILE      TyreWearDtoUnit = "MILE"
)

// Defines values for VehicleDtoBrand.
const (
	VehicleDtoBrandBMW         VehicleDtoBrand = "BMW"
	VehicleDtoBrandBMWMotorrad VehicleDtoBrand = "BMW Motorrad"
	VehicleDtoBrandBMWi        VehicleDtoBrand = "BMWi"
	VehicleDtoBrandMINI        VehicleDtoBrand = "MINI"
	VehicleDtoBrandRollsRoyce  VehicleDtoBrand = "Rolls-Royce"
	VehicleDtoBrandToyotaSupra VehicleDtoBrand = "Toyota-Supra"
)

// Defines values for VehicleMappingDtoMappingType.
const (
	VehicleMappingDtoMappingTypePRIMARY   VehicleMappingDtoMappingType = "PRIMARY"
	VehicleMappingDtoMappingTypeSECONDARY VehicleMappingDtoMappingType = "SECONDARY"
)

// AggregatedQualityStatusDto defines model for AggregatedQualityStatusDto.
type AggregatedQualityStatusDto struct {
	Label         *string `json:"label,omitempty"`
	QualityStatus *string `json:"qualityStatus,omitempty"`
	Value         *string `json:"value,omitempty"`
}

// BusinessErrorDto defines model for BusinessErrorDto.
type BusinessErrorDto struct {
	CreationTime *time.Time `json:"creationTime,omitempty"`
	Hint         *string    `json:"hint,omitempty"`
}

// ChargingBlockDto defines model for ChargingBlockDto.
type ChargingBlockDto struct {
	AveragePowerGridKw *float64 `json:"averagePowerGridKw,omitempty"`
	EndTime            int64    `json:"endTime"`
	StartTime          int64    `json:"startTime"`
}

// ChargingHistoryResponseDto defines model for ChargingHistoryResponseDto.
type ChargingHistoryResponseDto struct {
	Data      []ChargingSessionDto `json:"data"`
	NextToken *string              `json:"next_token,omitempty"`
}

// ChargingLocationDto defines model for ChargingLocationDto.
type ChargingLocationDto struct {
	FormattedAddress    string   `json:"formattedAddress"`
	MapMatchedLatitude  *float32 `json:"mapMatchedLatitude,omitempty"`
	MapMatchedLongitude *float32 `json:"mapMatchedLongitude,omitempty"`
	Municipality        string   `json:"municipality"`
	StreetAddress       string   `json:"streetAddress"`
}

// ChargingSessionDto defines model for ChargingSessionDto.
type ChargingSessionDto struct {
	BusinessErrors                 *[]BusinessErrorDto            `json:"businessErrors,omitempty"`
	ChargingBlocks                 *[]ChargingBlockDto            `json:"chargingBlocks,omitempty"`
	ChargingLocation               *ChargingLocationDto           `json:"chargingLocation,omitempty"`
	DisplayedSoc                   int32                          `json:"displayedSoc"`
	DisplayedStartSoc              int32                          `json:"displayedStartSoc"`
	EndTime                        int64                          `json:"endTime"`
	EnergyConsumedFromPowerGridKwh *float64                       `json:"energyConsumedFromPowerGridKwh,omitempty"`
	EnergyDecreaseHvbKwh           *float64                       `json:"energyDecreaseHvbKwh,omitempty"`
	EnergyDischargedKwh            *float64                       `json:"energyDischargedKwh,omitempty"`
	IsPreconditioningActivated     bool                           `json:"isPreconditioningActivated"`
	Mileage                        int32                          `json:"mileage"`
	MileageUnits                   ChargingSessionDtoMileageUnits `json:"mileageUnits"`
	PublicChargingPoint            *PublicChargingPointDto        `json:"publicChargingPoint,omitempty"`
	StartTime                      int64                          `json:"startTime"`
	TimeZone                       string                         `json:"timeZone"`
	TotalChargingDurationSec       int32                          `json:"totalChargingDurationSec"`
}

// ChargingSessionDtoMileageUnits defines model for ChargingSessionDto.MileageUnits.
type ChargingSessionDtoMileageUnits string

// ContainerDetailsDto defines model for ContainerDetailsDto.
type ContainerDetailsDto struct {
	ContainerId          *string                   `json:"containerId,omitempty"`
	Created              *time.Time                `json:"created,omitempty"`
	Name                 *string                   `json:"name,omitempty"`
	Purpose              *string                   `json:"purpose,omitempty"`
	State                *ContainerDetailsDtoState `json:"state,omitempty"`
	TechnicalDescriptors *[]string                 `json:"technicalDescriptors,omitempty"`
}

// ContainerDetailsDtoState defines model for ContainerDetailsDto.State.
type ContainerDetailsDtoState string

// ContainerDto defines model for ContainerDto.
type ContainerDto struct {
	ContainerId *string            `json:"containerId,omitempty"`
	Created     *time.Time         `json:"created,omitempty"`
	Name        *string            `json:"name,omitempty"`
	Purpose     *string            `json:"purpose,omitempty"`
	State       *ContainerDtoState `json:"state,omitempty"`
}

// ContainerDtoState defines model for ContainerDto.State.
type ContainerDtoState string

// ContainerListDto defines model for ContainerListDto.
type ContainerListDto struct {
	Containers *[]ContainerDto `json:"containers,omitempty"`
}

// CreateContainerRequest is the container creation request payload.
type CreateContainerRequest struct {
	Name                 *string   `json:"name,omitempty"`
	Purpose              *string   `json:"purpose,omitempty"`
	TechnicalDescriptors *[]string `json:"technicalDescriptors,omitempty"`
}

// DimensionDto defines model for DimensionDto.
type DimensionDto struct {
	AspectRatio      *int32  `json:"aspectRatio,omitempty"`
	ConstructionType *string `json:"constructionType,omitempty"`
	Label            *string `json:"label,omitempty"`
	LoadIndex        *int32  `json:"loadIndex,omitempty"`
	RimDiameter      *int32  `json:"rimDiameter,omitempty"`
	SectionWidth     *int32  `json:"sectionWidth,omitempty"`
	SpeedRating      *string `json:"speedRating,omitempty"`
	Value            *string `json:"value,omitempty"`
}

// ExVeTelematicDataResponseDto defines model for ExVeTelematicDataResponseDto.
type ExVeTelematicDataResponseDto struct {
	TelematicData *map[string]TelematicDataEntryDto `json:"telematicData,omitempty"`
}

// LocationBasedChargingSettingsDataDto defines model for LocationBasedChargingSettingsDataDto.
type LocationBasedChargingSettingsDataDto struct {
	AcCurrentLimit                *float64                              `json:"acCurrentLimit,omitempty"`
	AcCurrentLimitFlag            *string                               `json:"acCurrentLimitFlag,omitempty"`
	AcousticLimit                 *string                               `json:"acousticLimit,omitempty"`
	ChargingMode                  *string                               `json:"chargingMode,omitempty"`
	ChargingPlug                  *string                               `json:"chargingPlug,omitempty"`
	ChargingTimeWindows           *[]LocationBasedChargingTimeWindowDto `json:"chargingTimeWindows,omitempty"`
	ClusterLocationId             *string                               `json:"clusterLocationId,omitempty"`
	FlapLock                      *string                               `json:"flapLock,omitempty"`
	Id                            *string                               `json:"id,omitempty"`
	LastUpdated                   *string                               `json:"lastUpdated,omitempty"`
	LastVisit                     *string                               `json:"lastVisit,omitempty"`
	Latitude                      *float64                              `json:"latitude,omitempty"`
	Longitude                     *float64                              `json:"longitude,omitempty"`
	OptimizedChargingPreference   *string                               `json:"optimizedChargingPreference,omitempty"`
	StartChargingTimePeriodHour   *int32                                `json:"startChargingTimePeriodHour,omitempty"`
	StartChargingTimePeriodMinute *int32                                `json:"startChargingTimePeriodMinute,omitempty"`
	StopChargingTimePeriodHour    *int32                                `json:"stopChargingTimePeriodHour,omitempty"`
	StopChargingTimePeriodMinute  *int32                                `json:"stopChargingTimePeriodMinute,omitempty"`
	VehicleIdWithGcid             *string                               `json:"vehicleIdWithGcid,omitempty"`
	Visits                        *int32                                `json:"visits,omitempty"`
}

// LocationBasedChargingSettingsDto defines model for LocationBasedChargingSettingsDto.
type LocationBasedChargingSettingsDto struct {
	Data      *[]LocationBasedChargingSettingsDataDto `json:"data,omitempty"`
	NextToken *string                                 `json:"nextToken,omitempty"`
}

// LocationBasedChargingTimeWindowDto defines model for LocationBasedChargingTimeWindowDto.
type LocationBasedChargingTimeWindowDto struct {
	StartChargingTimePeriodHour   *int32 `json:"startChargingTimePeriodHour,omitempty"`
	StartChargingTimePeriodMinute *int32 `json:"startChargingTimePeriodMinute,omitempty"`
	StopChargingTimePeriodHour    *int32 `json:"stopChargingTimePeriodHour,omitempty"`
	StopChargingTimePeriodMinute  *int32 `json:"stopChargingTimePeriodMinute,omitempty"`
}

// MountingDateDto defines model for MountingDateDto.
type MountingDateDto struct {
	Label        *string `json:"label,omitempty"`
	MountingDate *string `json:"mountingDate,omitempty"`
	Value        *string `json:"value,omitempty"`
}

// OptimizedForOemDto defines model for OptimizedForOemDto.
type OptimizedForOemDto struct {
	Label           *string `json:"label,omitempty"`
	OptimizedForOem *string `json:"optimizedForOem,omitempty"`
	Value           *string `json:"value,omitempty"`
}

// PartNumberDto defines model for PartNumberDto.
type PartNumberDto struct {
	Label      *string `json:"label,omitempty"`
	PartNumber *string `json:"partNumber,omitempty"`
	Value      *string `json:"value,omitempty"`
}

// PassengerCarDto defines model for PassengerCarDto.
type PassengerCarDto struct {
	MountedTyres   *PassengerCarTyreSetDto `json:"mountedTyres,omitempty"`
	UnmountedTyres *PassengerCarTyreSetDto `json:"unmountedTyres,omitempty"`
}

// PassengerCarTyreSetDto defines model for PassengerCarTyreSetDto.
type PassengerCarTyreSetDto struct {
	AggregatedQualityStatus *AggregatedQualityStatusDto `json:"aggregatedQualityStatus,omitempty"`
	FrontLeft               *TyreDataDto                `json:"frontLeft,omitempty"`
	FrontRight              *TyreDataDto                `json:"frontRight,omitempty"`
	Label                   *string                     `json:"label,omitempty"`
	RearLeft                *TyreDataDto                `json:"rearLeft,omitempty"`
	RearRight               *TyreDataDto                `json:"rearRight,omitempty"`
}

// PublicChargingPointDto defines model for PublicChargingPointDto.
type PublicChargingPointDto struct {
	PotentialChargingPointMatches []PublicChargingPointMatchesDto `json:"potentialChargingPointMatches"`
}

// PublicChargingPointMatchesDto defines model for PublicChargingPointMatchesDto.
type PublicChargingPointMatchesDto struct {
	City          *string `json:"city,omitempty"`
	PostalCode    *string `json:"postalCode,omitempty"`
	ProviderName  *string `json:"providerName,omitempty"`
	StreetAddress *string `json:"streetAddress,omitempty"`
}

// QualityStatusDto defines model for QualityStatusDto.
type QualityStatusDto struct {
	Label         *string `json:"label,omitempty"`
	QualityStatus *string `json:"qualityStatus,omitempty"`
	Value         *string `json:"value,omitempty"`
}

// RunFlatDto defines model for RunFlatDto.
type RunFlatDto struct {
	Label   *string `json:"label,omitempty"`
	RunFlat *bool   `json:"runFlat,omitempty"`
	Value   *string `json:"value,omitempty"`
}

// SeasonDto defines model for SeasonDto.
type SeasonDto struct {
	Label  *string `json:"label,omitempty"`
	Season *string `json:"season,omitempty"`
	Value  *string `json:"value,omitempty"`
}

// SmartMaintenanceErrorDto defines model for SmartMaintenanceErrorDto.
type SmartMaintenanceErrorDto struct {
	EventId *string                 `json:"eventId,omitempty"`
	Message *map[string]interface{} `json:"message,omitempty"`
	Type    *string                 `json:"type,omitempty"`
}

// SmartMaintenanceTyreDiagnosisDto defines model for SmartMaintenanceTyreDiagnosisDto.
type SmartMaintenanceTyreDiagnosisDto struct {
	Errors       *[]SmartMaintenanceErrorDto `json:"errors,omitempty"`
	PassengerCar *PassengerCarDto            `json:"passengerCar,omitempty"`
}

// TelematicDataEntryDto defines model for TelematicDataEntryDto.
type TelematicDataEntryDto struct {
	Timestamp *string `json:"timestamp,omitempty"`
	Unit      *string `json:"unit,omitempty"`
	Value     *string `json:"value,omitempty"`
}

// TreadDto defines model for TreadDto.
type TreadDto struct {
	Carcass      *string `json:"carcass,omitempty"`
	Label        *string `json:"label,omitempty"`
	Manufacturer *string `json:"manufacturer,omitempty"`
	TreadDesign  *string `json:"treadDesign,omitempty"`
	Value        *string `json:"value,omitempty"`
}

// TyreDataDto defines model for TyreDataDto.
type TyreDataDto struct {
	Dimension          *DimensionDto          `json:"dimension,omitempty"`
	Label              *string                `json:"label,omitempty"`
	MountingDate       *MountingDateDto       `json:"mountingDate,omitempty"`
	OptimizedForOem    *OptimizedForOemDto    `json:"optimizedForOem,omitempty"`
	PartNumber         *PartNumberDto         `json:"partNumber,omitempty"`
	QualityStatus      *QualityStatusDto      `json:"qualityStatus,omitempty"`
	RunFlat            *RunFlatDto            `json:"runFlat,omitempty"`
	Season             *SeasonDto             `json:"season,omitempty"`
	Tread              *TreadDto              `json:"tread,omitempty"`
	TyreDefect         *TyreDefectDto         `json:"tyreDefect,omitempty"`
	TyreProductionDate *TyreProductionDateDto `json:"tyreProductionDate,omitempty"`
	TyreWear           *TyreWearDto           `json:"tyreWear,omitempty"`
}

// TyreDefectDto defines model for TyreDefectDto.
type TyreDefectDto struct {
	Label       *string                   `json:"label,omitempty"`
	Status      *string                   `json:"status,omitempty"`
	StatusColor *TyreDefectDtoStatusColor `json:"statusColor,omitempty"`
	Value       *string                   `json:"value,omitempty"`
}

// TyreDefectDtoStatusColor defines model for TyreDefectDto.StatusColor.
type TyreDefectDtoStatusColor string

// TyreProductionDateDto defines model for TyreProductionDateDto.
type TyreProductionDateDto struct {
	Label       *string                           `json:"label,omitempty"`
	StatusColor *TyreProductionDateDtoStatusColor `json:"statusColor,omitempty"`
	Value       *string                           `json:"value,omitempty"`
}

// TyreProductionDateDtoStatusColor defines model for TyreProductionDateDto.StatusColor.
type TyreProductionDateDtoStatusColor string

// TyreWearDto defines model for TyreWearDto.
type TyreWearDto struct {
	DueMileage  *int32                  `json:"dueMileage,omitempty"`
	Label       *string                 `json:"label,omitempty"`
	Status      *string                 `json:"status,omitempty"`
	StatusColor *TyreWearDtoStatusColor `json:"statusColor,omitempty"`
	Unit        *TyreWearDtoUnit        `json:"unit,omitempty"`
	Value       *string                 `json:"value,omitempty"`
}

// TyreWearDtoStatusColor defines model for TyreWearDto.StatusColor.
type TyreWearDtoStatusColor string

// TyreWearDtoUnit defines model for TyreWearDto.Unit.
type TyreWearDtoUnit string

// VehicleDto defines model for VehicleDto.
type VehicleDto struct {
	BodyType                  *string          `json:"bodyType,omitempty"`
	Brand                     *VehicleDtoBrand `json:"brand,omitempty"`
	ChargingModes             *[]string        `json:"chargingModes,omitempty"`
	ColourCode                *string          `json:"colourCode,omitempty"`
	ColourCodeRaw             *string          `json:"colourCodeRaw,omitempty"`
	ConstructionDate          *string          `json:"constructionDate,omitempty"`
	CountryCodeISO            *string          `json:"countryCodeISO,omitempty"`
	DriveTrain                *string          `json:"driveTrain,omitempty"`
	Engine                    *string          `json:"engine,omitempty"`
	FullSAList                *string          `json:"fullSAList,omitempty"`
	HasNavi                   *bool            `json:"hasNavi,omitempty"`
	HasSunRoof                *bool            `json:"hasSunRoof,omitempty"`
	HeadUnit                  *string          `json:"headUnit,omitempty"`
	HvsMaxEnergyAbsolute      *string          `json:"hvsMaxEnergyAbsolute,omitempty"`
	IsTelematicsCapable       *bool            `json:"isTelematicsCapable,omitempty"`
	ModelKey                  *string          `json:"modelKey,omitempty"`
	ModelName                 *string          `json:"modelName,omitempty"`
	ModelRange                *string          `json:"modelRange,omitempty"`
	NumberOfDoors             *int32           `json:"numberOfDoors,omitempty"`
	PropulsionType            *string          `json:"propulsionType,omitempty"`
	PuStep                    *string          `json:"puStep,omitempty"`
	ReessNominalCapacityGross *string          `json:"reessNominalCapacityGross,omitempty"`
	Series                    *string          `json:"series,omitempty"`
	SimStatus                 *string          `json:"simStatus,omitempty"`
	Steering                  *string          `json:"steering,omitempty"`
	Vin                       *string          `json:"vin,omitempty"`
}

// VehicleDtoBrand defines model for VehicleDto.Brand.
type VehicleDtoBrand string

// VehicleMappingDto defines model for VehicleMappingDto.
type VehicleMappingDto struct {
	MappedSince *time.Time                    `json:"mappedSince,omitempty"`
	MappingType *VehicleMappingDtoMappingType `json:"mappingType,omitempty"`
	Vin         *string                       `json:"vin,omitempty"`
}

// VehicleMappingDtoMappingType defines model for VehicleMappingDto.MappingType.
type VehicleMappingDtoMappingType string

// The following *Params types carry the parameters (mostly the required
// "x-version" header) for each operation. They mirror the OpenAPI operation
// parameters and keep call sites explicit.

// ListContainersParams defines parameters for ListContainers.
type ListContainersParams struct {
	XVersion string `json:"x-version"`
}

// CreateContainerParams defines parameters for CreateContainer.
type CreateContainerParams struct {
	XVersion string `json:"x-version"`
}

// DeleteContainerParams defines parameters for DeleteContainer.
type DeleteContainerParams struct {
	XVersion string `json:"x-version"`
}

// GetContainerDetailsParams defines parameters for GetContainerDetails.
type GetContainerDetailsParams struct {
	XVersion string `json:"x-version"`
}

// GetMappingsParams defines parameters for GetMappings.
type GetMappingsParams struct {
	XVersion string `json:"x-version"`
}

// GetBasicDataParams defines parameters for GetBasicData.
type GetBasicDataParams struct {
	XVersion string `json:"x-version"`
}

// GetChargingHistoryParams defines parameters for GetChargingHistory.
type GetChargingHistoryParams struct {
	NextToken *string   `json:"nextToken,omitempty"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
	XVersion  string    `json:"x-version"`
}

// GetImageParams defines parameters for GetImage.
type GetImageParams struct {
	XVersion string `json:"x-version"`
}

// GetLocationBasedChargingSettingsParams defines parameters for GetLocationBasedChargingSettings.
type GetLocationBasedChargingSettingsParams struct {
	NextToken *string `json:"nextToken,omitempty"`
	XVersion  string  `json:"x-version"`
}

// GetSmartMaintenanceTyreDiagnosisParams defines parameters for GetSmartMaintenanceTyreDiagnosis.
type GetSmartMaintenanceTyreDiagnosisParams struct {
	XVersion string `json:"x-version"`
}

// GetTelematicDataParams defines parameters for GetTelematicData.
type GetTelematicDataParams struct {
	ContainerId string `json:"containerId"`
	XVersion    string `json:"x-version"`
}

// CreateContainerJSONRequestBody defines the JSON body for CreateContainer.
type CreateContainerJSONRequestBody = CreateContainerRequest

// CreateContainerResponse is the body returned by a successful CreateContainer
// call (HTTP 201).
type CreateContainerResponse = ContainerDetailsDto

// DeleteContainerResponse is returned by a successful DeleteContainer call.
// The API replies with an empty body (HTTP 204), so the type carries no fields.
type DeleteContainerResponse struct{}

// RequestEditorFn is invoked right before a request is sent, allowing callers to
// mutate it (for example to inject an Authorization header).
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// HttpRequestDoer performs HTTP requests. The standard *http.Client satisfies it.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientInterface is the set of operations exposed by the CarData customer API.
// Every method returns the raw *http.Response so the caller controls decoding.
type ClientInterface interface {
	ListContainers(ctx context.Context, params *ListContainersParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	CreateContainerWithBody(ctx context.Context, params *CreateContainerParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)
	CreateContainer(ctx context.Context, params *CreateContainerParams, body CreateContainerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
	DeleteContainer(ctx context.Context, containerId string, params *DeleteContainerParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetContainerDetails(ctx context.Context, containerId string, params *GetContainerDetailsParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetMappings(ctx context.Context, params *GetMappingsParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetBasicData(ctx context.Context, vin string, params *GetBasicDataParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetChargingHistory(ctx context.Context, vin string, params *GetChargingHistoryParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetImage(ctx context.Context, vin string, params *GetImageParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetLocationBasedChargingSettings(ctx context.Context, vin string, params *GetLocationBasedChargingSettingsParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetSmartMaintenanceTyreDiagnosis(ctx context.Context, vin string, params *GetSmartMaintenanceTyreDiagnosisParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetTelematicData(ctx context.Context, vin string, params *GetTelematicDataParams, reqEditors ...RequestEditorFn) (*http.Response, error)
}

// Client is a CarData customer API client.
type Client struct {
	// Server is the base URL of the API, e.g. https://api-cardata.bmwgroup.com.
	Server string
	// Client performs the HTTP requests. Defaults to http.DefaultClient's behaviour.
	Client HttpRequestDoer
	// RequestEditors are applied to every request before it is sent.
	RequestEditors []RequestEditorFn
}

var _ ClientInterface = (*Client)(nil)

// ClientOption customises a Client during construction.
type ClientOption func(*Client) error

// NewClient creates a Client for the given server URL with reasonable defaults.
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	c := &Client{Server: server}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	if c.Client == nil {
		c.Client = &http.Client{}
	}
	return c, nil
}

// WithHTTPClient overrides the HTTP doer, which is useful in tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn registers a callback invoked right before every request is
// sent. It is typically used to inject authentication headers.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// do builds and sends a request. path is appended to the server URL, query may
// be nil, and version (when non-empty) is sent as the required x-version header.
func (c *Client) do(ctx context.Context, method, path, version string, query url.Values, body io.Reader, contentType string, reqEditors []RequestEditorFn) (*http.Response, error) {
	u, err := url.Parse(c.Server)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimSuffix(u.Path, "/") + path
	if query != nil {
		u.RawQuery = query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}
	if version != "" {
		req.Header.Set("x-version", version)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for _, edit := range c.RequestEditors {
		if err := edit(ctx, req); err != nil {
			return nil, err
		}
	}
	for _, edit := range reqEditors {
		if err := edit(ctx, req); err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) ListContainers(ctx context.Context, params *ListContainersParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/customers/containers", params.XVersion, nil, nil, "", reqEditors)
}

func (c *Client) CreateContainerWithBody(ctx context.Context, params *CreateContainerParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, "/customers/containers", params.XVersion, nil, body, contentType, reqEditors)
}

func (c *Client) CreateContainer(ctx context.Context, params *CreateContainerParams, body CreateContainerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.CreateContainerWithBody(ctx, params, "application/json", bytes.NewReader(buf), reqEditors...)
}

func (c *Client) DeleteContainer(ctx context.Context, containerId string, params *DeleteContainerParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.do(ctx, http.MethodDelete, "/customers/containers/"+url.PathEscape(containerId), params.XVersion, nil, nil, "", reqEditors)
}

func (c *Client) GetContainerDetails(ctx context.Context, containerId string, params *GetContainerDetailsParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/customers/containers/"+url.PathEscape(containerId), params.XVersion, nil, nil, "", reqEditors)
}

func (c *Client) GetMappings(ctx context.Context, params *GetMappingsParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/customers/vehicles/mappings", params.XVersion, nil, nil, "", reqEditors)
}

func (c *Client) GetBasicData(ctx context.Context, vin string, params *GetBasicDataParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/customers/vehicles/"+url.PathEscape(vin)+"/basicData", params.XVersion, nil, nil, "", reqEditors)
}

func (c *Client) GetChargingHistory(ctx context.Context, vin string, params *GetChargingHistoryParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	query := url.Values{}
	query.Set("from", formatTime(params.From))
	query.Set("to", formatTime(params.To))
	if params.NextToken != nil {
		query.Set("nextToken", *params.NextToken)
	}
	return c.do(ctx, http.MethodGet, "/customers/vehicles/"+url.PathEscape(vin)+"/chargingHistory", params.XVersion, query, nil, "", reqEditors)
}

func (c *Client) GetImage(ctx context.Context, vin string, params *GetImageParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/customers/vehicles/"+url.PathEscape(vin)+"/image", params.XVersion, nil, nil, "", reqEditors)
}

func (c *Client) GetLocationBasedChargingSettings(ctx context.Context, vin string, params *GetLocationBasedChargingSettingsParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	var query url.Values
	if params.NextToken != nil {
		query = url.Values{"nextToken": {*params.NextToken}}
	}
	return c.do(ctx, http.MethodGet, "/customers/vehicles/"+url.PathEscape(vin)+"/locationBasedChargingSettings", params.XVersion, query, nil, "", reqEditors)
}

func (c *Client) GetSmartMaintenanceTyreDiagnosis(ctx context.Context, vin string, params *GetSmartMaintenanceTyreDiagnosisParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/customers/vehicles/"+url.PathEscape(vin)+"/smartMaintenanceTyreDiagnosis", params.XVersion, nil, nil, "", reqEditors)
}

func (c *Client) GetTelematicData(ctx context.Context, vin string, params *GetTelematicDataParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	query := url.Values{"containerId": {params.ContainerId}}
	return c.do(ctx, http.MethodGet, "/customers/vehicles/"+url.PathEscape(vin)+"/telematicData", params.XVersion, query, nil, "", reqEditors)
}

// formatTime renders a time as RFC 3339, the format expected by the API for the
// chargingHistory "from"/"to" query parameters.
func formatTime(t time.Time) string {
	b, _ := t.MarshalText()
	return string(b)
}
