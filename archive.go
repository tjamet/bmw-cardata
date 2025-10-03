package bmwcardata

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"path/filepath"
	"strings"
)

// ZipReader represents a zip file reader
type ZipReader struct {
	reader *zip.ReadCloser
}

// NewZipReader creates a new zip reader from the given file path
func NewZipReader(path string) (*ZipReader, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	return &ZipReader{reader: r}, nil
}

// Close closes the zip reader
func (z *ZipReader) Close() error {
	return z.reader.Close()
}

// Files returns a list of files in the zip archive
func (z *ZipReader) Files() []*zip.File {
	return z.reader.File
}

// ReadArchive reads an archive from a file downloaded from the BMW CarData portal
// It parses the zip file and returns a structured representation of the archive
func ReadArchive(path string) (*Archive, error) {
	zipReader, err := NewZipReader(path)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()
	archiveContent := customerArchiveContent{}
	archiveRelPath := ""
	for _, file := range zipReader.Files() {
		if strings.Contains(file.Name, "KeyList") && strings.HasSuffix(file.Name, ".xml") {
			archiveRelPath = filepath.Dir(file.Name)
			fd, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer fd.Close()
			err = xml.NewDecoder(fd).Decode(&archiveContent)
			if err != nil {
				return nil, err
			}
		}
	}
	archive := Archive{
		Lang:                archiveContent.Lang,
		RequestDate:         archiveContent.RequestDate,
		VIN:                 archiveContent.VIN,
		UnitOfLength:        archiveContent.UnitOfLength,
		BasicVehicleData:    archiveContent.BasicVehicleData,
		CasaContractDetails: archiveContent.CasaContractDetailsDataList,
		TelematicValues:     archiveContent.TelematicValues,
		VehicleImage:        archiveContent.VehicleImage,
	}
	if archiveContent.ChargingHistoryFileName != "" {
		fd, err := zipReader.reader.Open(filepath.Join(archiveRelPath, archiveContent.ChargingHistoryFileName))
		if err != nil {
			return nil, err
		}
		defer fd.Close()
		err = json.NewDecoder(fd).Decode(&archive.ChargingHistory)
		if err != nil {
			return nil, err
		}
	}
	if archiveContent.SmartMaintenanceFileName != "" {
		fd, err := zipReader.reader.Open(filepath.Join(archiveRelPath, archiveContent.SmartMaintenanceFileName))
		if err != nil {
			return nil, err
		}
		defer fd.Close()
		err = json.NewDecoder(fd).Decode(&archive.SmartMaintenance)
	}
	return &archive, nil
}
