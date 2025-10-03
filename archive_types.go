package bmwcardata

type ChargingSessionArchive struct {
	ChargingBlocks                 []ChargingBlock          `json:"chargingBlocks,omitempty"`
	ChargingCostInformation        *ChargingCostInformation `json:"chargingCostInformation,omitempty"`
	ChargingLocation               *ChargingLocation        `json:"chargingLocation,omitempty"`
	DisplayedSoc                   int                      `json:"displayedSoc,omitempty"`
	DisplayedStartSoc              int                      `json:"displayedStartSoc,omitempty"`
	EndTime                        int64                    `json:"endTime,omitempty"`
	EnergyConsumedFromPowerGridKwh float64                  `json:"energyConsumedFromPowerGridKwh,omitempty"`
	IsPreconditioningActivated     bool                     `json:"isPreconditioningActivated,omitempty"`
	Mileage                        int64                    `json:"mileage,omitempty"`
	MileageUnits                   string                   `json:"mileageUnits,omitempty"`
	PublicChargingPoint            *PublicChargingPoint     `json:"publicChargingPoint,omitempty"`
	StartTime                      int64                    `json:"startTime,omitempty"`
	TimeZone                       string                   `json:"timeZone,omitempty"`
	TotalChargingDurationSec       int64                    `json:"totalChargingDurationSec,omitempty"`
	BusinessErrors                 []BusinessError          `json:"businessErrors,omitempty"`
}

// ChargingBlock is currently unspecified in the provided samples (arrays are empty).
// It is defined to allow forward compatibility if the archive later includes block details.
type ChargingBlock struct {
	IsPreconditioningActivated     bool    `json:"isPreconditioningActivated,omitempty"`
	AverageChargingPowerW          float64 `json:"averagePowerGridKw,omitempty"`
	EnergyConsumedFromPowerGridKwh float64 `json:"energyConsumedFromPowerGridKwh,omitempty"`
	StartTime                      int64   `json:"startTime,omitempty"`
	EndTime                        int64   `json:"endTime,omitempty"`
}

type ChargingLocation struct {
	FormattedAddress    string  `json:"formattedAddress,omitempty"`
	MapMatchedLatitude  float64 `json:"mapMatchedLatitude,omitempty"`
	MapMatchedLongitude float64 `json:"mapMatchedLongitude,omitempty"`
	Municipality        string  `json:"municipality,omitempty"`
	StreetAddress       string  `json:"streetAddress,omitempty"`
}

type ChargingCostInformation struct {
	CalculatedChargingCost float64 `json:"calculatedChargingCost,omitempty"`
	CalculatedSavings      float64 `json:"calculatedSavings,omitempty"`
	Currency               string  `json:"currency,omitempty"`
}

type PublicChargingPoint struct {
	PotentialChargingPointMatches []PotentialChargingPointMatch `json:"potentialChargingPointMatches,omitempty"`
}

type PotentialChargingPointMatch struct {
	City          string `json:"city,omitempty"`
	PostalCode    string `json:"postalCode,omitempty"`
	ProviderName  string `json:"providerName,omitempty"`
	StreetAddress string `json:"streetAddress,omitempty"`
}

type BusinessError struct {
	CreationTime string `json:"creationTime,omitempty"`
	Hint         string `json:"hint,omitempty"`
}

type SmartMaintenanceArchive struct {
	Errors       []SmartMaintenanceError `json:"errors,omitempty"`
	PassengerCar *TyresPassengerCar      `json:"passengerCar,omitempty"`
}

// SmartMaintenanceError is intentionally empty to be forward-compatible with
// potential error objects present in the archive. Unknown fields are ignored.
type SmartMaintenanceError struct{}

// TyresPassengerCar groups mounted and unmounted tyres sections.
type TyresPassengerCar struct {
	MountedTyres   *TyreSet `json:"mountedTyres,omitempty"`
	UnmountedTyres *TyreSet `json:"unmountedTyres,omitempty"`
}

// TyreSet represents a set of tyres and an aggregated quality status.
type TyreSet struct {
	AggregatedQualityStatus *QualityStatus `json:"aggregatedQualityStatus,omitempty"`
	FrontLeft               *Tyre          `json:"frontLeft,omitempty"`
	FrontRight              *Tyre          `json:"frontRight,omitempty"`
	RearLeft                *Tyre          `json:"rearLeft,omitempty"`
	RearRight               *Tyre          `json:"rearRight,omitempty"`
	Label                   string         `json:"label,omitempty"`
}

// QualityStatus captures a labeled quality status with an optional value.
type QualityStatus struct {
	Label         string `json:"label,omitempty"`
	QualityStatus string `json:"qualityStatus,omitempty"`
	Value         string `json:"value,omitempty"`
}

// Tyre describes a single tyre entry with multiple labeled attributes.
type Tyre struct {
	Dimension          *TyreDimension       `json:"dimension,omitempty"`
	Label              string               `json:"label,omitempty"`
	MountingDate       *TyreMountingDate    `json:"mountingDate,omitempty"`
	OptimizedForOem    *TyreOptimizedForOem `json:"optimizedForOem,omitempty"`
	PartNumber         *TyrePartNumber      `json:"partNumber,omitempty"`
	QualityStatus      *QualityStatus       `json:"qualityStatus,omitempty"`
	RunFlat            *TyreRunFlat         `json:"runFlat,omitempty"`
	Season             *TyreSeason          `json:"season,omitempty"`
	Tread              *TyreTread           `json:"tread,omitempty"`
	TyreDefect         *TyreDefect          `json:"tyreDefect,omitempty"`
	TyreProductionDate *TyreProductionDate  `json:"tyreProductionDate,omitempty"`
	TyreWear           *TyreWear            `json:"tyreWear,omitempty"`
}

// TyreDimension models the tyre dimension details.
type TyreDimension struct {
	AspectRatio      int    `json:"aspectRatio,omitempty"`
	ConstructionType string `json:"constructionType,omitempty"`
	Label            string `json:"label,omitempty"`
	LoadIndex        int    `json:"loadIndex,omitempty"`
	RimDiameter      int    `json:"rimDiameter,omitempty"`
	SectionWidth     int    `json:"sectionWidth,omitempty"`
	SpeedRating      string `json:"speedRating,omitempty"`
	Value            string `json:"value,omitempty"`
}

// TyreMountingDate holds the mounting date value with labels.
type TyreMountingDate struct {
	Label        string `json:"label,omitempty"`
	MountingDate string `json:"mountingDate,omitempty"`
	Value        string `json:"value,omitempty"`
}

// TyreOptimizedForOem indicates OEM optimization.
type TyreOptimizedForOem struct {
	Label           string `json:"label,omitempty"`
	OptimizedForOem string `json:"optimizedForOem,omitempty"`
	Value           string `json:"value,omitempty"`
}

// TyrePartNumber captures the tyre part number.
type TyrePartNumber struct {
	Label      string `json:"label,omitempty"`
	PartNumber string `json:"partNumber,omitempty"`
	Value      string `json:"value,omitempty"`
}

// TyreRunFlat indicates whether the tyre is run-flat.
type TyreRunFlat struct {
	Label   string `json:"label,omitempty"`
	RunFlat bool   `json:"runFlat,omitempty"`
	Value   string `json:"value,omitempty"`
}

// TyreSeason specifies the tyre season (e.g., SUMMER/WINTER).
type TyreSeason struct {
	Label  string `json:"label,omitempty"`
	Season string `json:"season,omitempty"`
	Value  string `json:"value,omitempty"`
}

// TyreTread holds tread manufacturer/design info.
type TyreTread struct {
	Carcass      string `json:"carcass,omitempty"`
	Label        string `json:"label,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	TreadDesign  string `json:"treadDesign,omitempty"`
	Value        string `json:"value,omitempty"`
}

// TyreDefect currently only exposes a label in provided samples.
type TyreDefect struct {
	Label string `json:"label,omitempty"`
}

// TyreProductionDate holds production date status and value (e.g., week-year code).
type TyreProductionDate struct {
	Label       string `json:"label,omitempty"`
	StatusColor string `json:"statusColor,omitempty"`
	Value       string `json:"value,omitempty"`
}

// TyreWear captures wear/service information.
type TyreWear struct {
	Label       string `json:"label,omitempty"`
	StatusColor string `json:"statusColor,omitempty"`
	Unit        string `json:"unit,omitempty"`
}

type Archive struct {
	VIN                 string                   `json:"vin,omitempty"`
	UnitOfLength        string                   `json:"unitOfLength,omitempty"`
	BasicVehicleData    BasicVehicleData         `json:"basicVehicleData,omitempty"`
	CasaContractDetails []CasaContractDetails    `json:"casaContractDetails,omitempty"`
	TelematicValues     []TelematicValues        `json:"telematicValues,omitempty"`
	VehicleImage        string                   `json:"vehicleImage,omitempty"`
	Lang                string                   `json:"lang,omitempty"`
	RequestDate         string                   `json:"requestDate,omitempty"`
	SmartMaintenance    SmartMaintenanceArchive  `json:"smartMaintenance,omitempty"`
	ChargingHistory     []ChargingSessionArchive `json:"chargingHistory,omitempty"`
}

// Types for parsing the BMW CarData "KeyList" XML (customerArchiveContent)
// via encoding/xml. These structs intentionally model only the parts we
// currently need; unknown elements/attributes will be ignored by the decoder.

// customerArchiveContent is the root element of the KeyList XML.
//
// Example root tag:
// <customerArchiveContent
//
//	chargingHistoryFileName="...json"
//	keyListFileName="...xml"
//	lang="ES"
//	requestDate="01-10-2025"
//	smartMaintenanceFileName="...json"
//	unitOfLength="km"
//	vin="WBY...">
type customerArchiveContent struct {
	ChargingHistoryFileName  string `xml:"chargingHistoryFileName,attr"`
	KeyListFileName          string `xml:"keyListFileName,attr"`
	Lang                     string `xml:"lang,attr"`
	RequestDate              string `xml:"requestDate,attr"`
	SmartMaintenanceFileName string `xml:"smartMaintenanceFileName,attr"`
	UnitOfLength             string `xml:"unitOfLength,attr"`
	VIN                      string `xml:"vin,attr"`

	BasicVehicleData            BasicVehicleData      `xml:"basicVehicleData"`
	CasaContractDetailsDataList []CasaContractDetails `xml:"casaContractDetailsDataList"`

	TelematicValues []TelematicValues `xml:"telematicValueList"`

	// Base64-encoded PNG data in the sample
	VehicleImage string `xml:"vehicleImage"`
}

// BasicVehicleData contains telematic name/value entries and a category.
type BasicVehicleData struct {
	DataCategory    string                `xml:"dataCategory,attr" json:"dataCategory,omitempty"`
	TelematicValues []BasicTelematicValue `xml:"telematicValue" json:"telematicValues,omitempty"`
}

// BasicTelematicValue represents a single name/value pair under basicVehicleData.
type BasicTelematicValue struct {
	Name  string `xml:"name" json:"name,omitempty"`
	Value string `xml:"value" json:"value,omitempty"`
}

type TelematicValues struct {
	DataCategory    string           `xml:"dataCategory,attr" json:"dataCategory,omitempty"`
	TelematicValues []TelematicValue `xml:"telematicValue" json:"telematicValues,omitempty"`
}

// BasicTelematicValue represents a single name/value pair under basicVehicleData.
type TelematicValue struct {
	Name             string `xml:"name" json:"name,omitempty"`
	Value            string `xml:"value" json:"value,omitempty"`
	Unit             string `xml:"unit" json:"unit,omitempty"`
	FetchTimestamp   string `xml:"fetchTimestamp" json:"fetchTimestamp,omitempty"`
	ValueTimestamp   string `xml:"valueTimestamp" json:"valueTimestamp,omitempty"`
	DataCategoryType string `xml:"dataCategoryType" json:"dataCategoryType,omitempty"`
	TelematicKeyName string `xml:"telematicKeyName" json:"telematicKeyName,omitempty"`
}

// CasaContractDetails represents a single casaContractDetailsDataList entry.
type CasaContractDetails struct {
	ContractPeriod ContractPeriod `xml:"contractPeriod" json:"contractPeriod,omitempty"`
	Name           string         `xml:"name" json:"name,omitempty"`
	OfferID        OfferID        `xml:"offerId" json:"offerId,omitempty"`
	Status         string         `xml:"status" json:"status,omitempty"`
}

// ContractPeriod holds optional start/end timestamps.
type ContractPeriod struct {
	Start string `xml:"start,omitempty" json:"start,omitempty"`
	End   string `xml:"end,omitempty" json:"end,omitempty"`
}

// OfferID holds offer identifiers for a contract.
type OfferID struct {
	GlobalID      string `xml:"globalId" json:"globalId,omitempty"`
	MasterOfferID string `xml:"masterOfferId" json:"masterOfferId,omitempty"`
}
