package bmwcardata

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"
)

type AdaptiveNavigationArchive struct {
	Places      []NavigationPlaces      `json:"places,omitempty"`
	Routes      []NavigationRoutes      `json:"routes,omitempty"`
	Transitions []NavigationTransitions `json:"transitions,omitempty"`
}

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

type NavigationPlaces struct {
	ClusterHistory                   []NavigationClusterHistory `json:"clusterHistory,omitempty"`
	DepartureToUnknownStatistics     VisitStatistics            `json:"departureToUnknownStatistics,omitempty"`
	Place                            Place                      `json:"place,omitempty"`
	VisitScoreDistributionStatistics VisitStatistics            `json:"visitScoreDistributionStatistics,omitempty"`
}

type NavigationTransitions struct {
	Category            string          `json:"category,omitempty"`
	Created             Time            `json:"created,omitempty"`
	DateTimeCreated     Time            `json:"dateTimeCreated,omitempty"`
	DateTimeUpdated     Time            `json:"dateTimeUpdated,omitempty"`
	DepartureStatistics VisitStatistics `json:"departureStatistics,omitempty"`
	DestinationID       string          `json:"destinationId,omitempty"`
	ID                  string          `json:"id,omitempty"`
	IsBlacklisted       bool            `json:"isBlacklisted,omitempty"`
	Metadata            []Metadata      `json:"metadata,omitempty"`
	Modified            Time            `json:"modified,omitempty"`
	OriginID            string          `json:"originId,omitempty"`
}

type Place struct {
	Created         Time        `json:"created,omitempty"`
	DateTimeCreated Time        `json:"dateTimeCreated,omitempty"`
	DateTimeUpdated Time        `json:"dateTimeUpdated,omitempty"`
	Center          Coordinates `json:"center,omitempty"`
	ID              string      `json:"id,omitempty"`
	IsBlacklisted   bool        `json:"isBlacklisted,omitempty"`
	LearnedLabel    Label       `json:"learnedLabel,omitempty"`
	//Metadata []any `json:"metadata,omitempty"`
	Modified       Time    `json:"modified,omitempty"`
	Radius         float64 `json:"radius,omitempty"`
	RelevanceScore float64 `json:"relevanceScore,omitempty"`
	UserEdited     bool    `json:"userEdited,omitempty"`
}

// Time is a wrapper around Time to allow detecting the format used
// in the BMW CarData archive. There has been 7 different formats
// detected so far.
// It is specially designed to re-serialize the time in the same format
// as the original data to help ensure there is no data loss when parsing
type Time struct {
	time.Time
	format string
	parsed bool
}

func (t *Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var data string
	if err := d.DecodeElement(&data, &start); err != nil {
		return err
	}
	return t.parseAndDetectFormat(data)
}

func (t *Time) parseAndDetectFormat(data string) error {
	for _, format := range []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.999999Z",
		"2006-01-02T15:04:05.999Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"02.01.2006 15:04:05 MST",
	} {
		parsed, err := time.Parse(format, string(data))
		if err == nil {
			t.Time = parsed
			t.format = format
			t.parsed = true
			return nil
		}
	}
	return fmt.Errorf("invalid time format: %s", string(data))
}

func (t *Time) UnmarshalJSON(data []byte) error {
	isNumeric := true
	for _, c := range data {
		if c < '0' || c > '9' {
			isNumeric = false
			break
		}
	}
	if isNumeric {
		parsed, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return err
		}
		t.Time = time.Unix(parsed, 0)
		t.format = "unix"
		t.parsed = true
		return nil
	}
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid time format: %s", string(data))
	}
	data = data[1 : len(data)-1]
	return t.parseAndDetectFormat(string(data))
}

func (t Time) MarshalJSON() ([]byte, error) {
	if !t.parsed {
		return []byte(`""`), nil
	}
	if t.format == "unix" {
		return []byte(fmt.Sprintf("%d", t.Time.Unix())), nil
	}
	if t.format == "" {
		return []byte(fmt.Sprintf("\"%s\"", t.Time.Format(time.RFC3339))), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", t.Time.Format(t.format))), nil
}

type Entry struct {
	Score float64  `json:"score,omitempty"`
	Key   EntryKey `json:"key,omitempty"`
}

type EntryKey struct {
	TimeBucketId int `json:"timeBucketId,omitempty"`
	Weekday      int `json:"weekday,omitempty"`
}

type Label struct {
	Label         string  `json:"label,omitempty"`
	SuccessFactor float64 `json:"successFactor,omitempty"`
}

type VisitStatistics struct {
	Created         Time    `json:"created,omitempty"`
	DateTimeCreated Time    `json:"dateTimeCreated,omitempty"`
	DateTimeUpdated Time    `json:"dateTimeUpdated,omitempty"`
	Entries         []Entry `json:"entries,omitempty"`
	Modified        Time    `json:"modified,omitempty"`
	Name            string  `json:"name,omitempty"`
	TimeResolution  int64   `json:"timeResolution,omitempty"`
}

type NavigationRoutes struct {
	Route      Route      `json:"route,omitempty"`
	Statistics Statistics `json:"statistics,omitempty"`
}

type Route struct {
	Created       Time       `json:"created,omitempty"`
	DestinationID string     `json:"destinationId,omitempty"`
	ID            string     `json:"id,omitempty"`
	Metadata      []Metadata `json:"metadata,omitempty"`
	Modified      Time       `json:"modified,omitempty"`
	OriginID      string     `json:"originId,omitempty"`
	Segments      []Segment  `json:"segments,omitempty"`
}

type Statistics struct {
	Created         Time    `json:"created,omitempty"`
	DateTimeCreated Time    `json:"dateTimeCreated,omitempty"`
	DateTimeUpdated Time    `json:"dateTimeUpdated,omitempty"`
	Entries         []Entry `json:"entries,omitempty"`
	Modified        Time    `json:"modified,omitempty"`
	Name            string  `json:"name,omitempty"`
	TimeResolution  int64   `json:"timeResolution,omitempty"`
}

type Segment struct {
	Locations []Location `json:"locations,omitempty"`
	SegmentID int        `json:"segmentId,omitempty"`
}

type Location struct {
	Location Coordinates `json:"location,omitempty"`
}

type Metadata struct {
	Count           int    `json:"count,omitempty"`
	DecayedCount    int    `json:"decayedCount,omitempty"`
	LastUpdatedTime Time   `json:"lastUpdatedTime,omitempty"`
	Name            string `json:"name,omitempty"`
	Value           string `json:"value,omitempty"`
}

type NavigationClusterHistory struct {
	Accuracy    float64     `json:"accuracy,omitempty"`
	Coordinates Coordinates `json:"coordinates,omitempty"`
}

type Coordinates struct {
	Latitude  float64 `json:"lat,omitempty"`
	Longitude float64 `json:"lng,omitempty"`
}

// ChargingBlock is currently unspecified in the provided samples (arrays are empty).
// It is defined to allow forward compatibility if the archive later includes block details.
type ChargingBlock struct {
	IsPreconditioningActivated     bool    `json:"isPreconditioningActivated,omitempty"`
	AverageChargingPowerW          float64 `json:"averagePowerGridKw,omitempty"`
	EnergyConsumedFromPowerGridKwh float64 `json:"energyConsumedFromPowerGridKwh,omitempty"`
	StartTime                      Time    `json:"startTime,omitempty"`
	EndTime                        Time    `json:"endTime,omitempty"`
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
	CreationTime Time   `json:"creationTime,omitempty"`
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
	MountingDate Time   `json:"mountingDate,omitempty"`
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
	VIN                 string                    `json:"vin,omitempty"`
	UnitOfLength        string                    `json:"unitOfLength,omitempty"`
	BasicVehicleData    BasicVehicleData          `json:"basicVehicleData,omitempty"`
	CasaContractDetails []CasaContractDetails     `json:"casaContractDetails,omitempty"`
	TelematicValues     []TelematicValues         `json:"telematicValues,omitempty"`
	VehicleImage        string                    `json:"vehicleImage,omitempty"`
	Lang                string                    `json:"lang,omitempty"`
	RequestDate         string                    `json:"requestDate,omitempty"`
	SmartMaintenance    SmartMaintenanceArchive   `json:"smartMaintenance,omitempty"`
	ChargingHistory     []ChargingSessionArchive  `json:"chargingHistory,omitempty"`
	AdaptiveNavigation  AdaptiveNavigationArchive `json:"adaptiveNavigationArchive,omitempty"`
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
	ChargingHistoryFileName    string `xml:"chargingHistoryFileName,attr"`
	KeyListFileName            string `xml:"keyListFileName,attr"`
	LearningNavigationFileName string `xml:"learningNavigationFileName,attr"`
	SmartMaintenanceFileName   string `xml:"smartMaintenanceFileName,attr"`
	Lang                       string `xml:"lang,attr"`
	RequestDate                string `xml:"requestDate,attr"`
	UnitOfLength               string `xml:"unitOfLength,attr"`
	VIN                        string `xml:"vin,attr"`

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
	FetchTimestamp   Time   `xml:"fetchTimestamp" json:"fetchTimestamp,omitempty"`
	ValueTimestamp   Time   `xml:"valueTimestamp" json:"valueTimestamp,omitempty"`
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
	Start Time `xml:"start,omitempty" json:"start,omitempty"`
	End   Time `xml:"end,omitempty" json:"end,omitempty"`
}

// OfferID holds offer identifiers for a contract.
type OfferID struct {
	GlobalID      string `xml:"globalId" json:"globalId,omitempty"`
	MasterOfferID string `xml:"masterOfferId" json:"masterOfferId,omitempty"`
}
