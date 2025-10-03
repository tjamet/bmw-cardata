package bmwcardata

type Descriptor struct {
	Name         string        `json:"name,omitempty"`
	ID           string        `json:"id,omitempty"`
	Description  string        `json:"description,omitempty"`
	Unit         string        `json:"unit,omitempty"`
	Range        string        `json:"range,omitempty"`
	DataType     string        `json:"datatype,omitempty"`
	Streamable   bool          `json:"streamable,omitempty"`
	VehicleTypes []VehicleType `json:"vehicletypes,omitempty"`
	Brand        []Brand       `json:"brand,omitempty"`
	Category     string        `json:"category,omitempty"`
}

type Category struct {
	Description string       `json:"description,omitempty"`
	Rank        int          `json:"rank,omitempty"`
	Name        string       `json:"name,omitempty"`
	Containers  []Descriptor `json:"containers,omitempty"`
}

type VehicleType string

type Brand string

const (
	VehicleTypeICE  VehicleType = "ICE"
	VehicleTypePHEV VehicleType = "PHEV"
	VehicleTypeBEV  VehicleType = "BEV"
	VehicleTypeMHEV VehicleType = "MHEV"
)

const (
	BrandBMW Brand = "BMW"
)
