package bmwcardata

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/tjamet/bmw-cardata/cardataapi"
)

type DescriptorMatcher interface {
	Match(container Descriptor) bool
}

type DescriptorMatcherFunc func(container Descriptor) bool

func (f DescriptorMatcherFunc) Match(container Descriptor) bool {
	return f(container)
}

func MatchID(id string) DescriptorMatcher {
	return DescriptorMatcherFunc(func(container Descriptor) bool {
		return container.ID == id
	})
}

func MatchBrand(brand Brand) DescriptorMatcher {
	return DescriptorMatcherFunc(func(container Descriptor) bool {
		return slices.ContainsFunc(container.Brand, func(b Brand) bool {
			return b == brand
		})
	})
}

func MatchVehicleType(vehicleType VehicleType) DescriptorMatcher {
	return DescriptorMatcherFunc(func(container Descriptor) bool {
		return slices.ContainsFunc(container.VehicleTypes, func(v VehicleType) bool {
			return v == vehicleType
		})
	})
}

func MatchCategory(category string) DescriptorMatcher {
	return DescriptorMatcherFunc(func(container Descriptor) bool {
		return container.Category == category
	})
}

func MatchAll(matchers ...DescriptorMatcher) DescriptorMatcher {
	return DescriptorMatcherFunc(func(container Descriptor) bool {
		for _, matcher := range matchers {
			if !matcher.Match(container) {
				return false
			}
		}
		return true
	})
}

func MatchStreamable(streamable bool) DescriptorMatcher {
	return DescriptorMatcherFunc(func(container Descriptor) bool {
		return container.Streamable == streamable
	})
}

func MatchDataTypes(dataTypes ...string) DescriptorMatcher {
	return DescriptorMatcherFunc(func(container Descriptor) bool {
		return slices.ContainsFunc(dataTypes, func(dataType string) bool {
			return container.DataType == dataType
		})
	})
}

func MatchAny(matchers ...DescriptorMatcher) DescriptorMatcher {
	return DescriptorMatcherFunc(func(container Descriptor) bool {
		for _, matcher := range matchers {
			if matcher.Match(container) {
				return true
			}
		}
		return false
	})
}

func MatchNone(matchers ...DescriptorMatcher) DescriptorMatcher {
	return DescriptorMatcherFunc(func(container Descriptor) bool {
		for _, matcher := range matchers {
			if matcher.Match(container) {
				return false
			}
		}
		return true
	})
}

func FindDescriptors(matcher DescriptorMatcher) []Descriptor {
	r := []Descriptor{}
	for _, descriptor := range allDescriptors {
		if matcher.Match(descriptor) {
			r = append(r, descriptor)
		}
	}
	return r
}

// ListContainers lists all the containers that are available in the BMW CarData API
// See https://bmw-cardata.bmwgroup.com/customer/public/api-specification#operations-Containers-listContainers
func (c *Client) ListContainers(ctx context.Context) (*cardataapi.ContainerListDto, error) {
	params := &cardataapi.ListContainersParams{XVersion: "v1"}
	resp, err := c.carDataAPI.ListContainers(ctx, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		data := cardataapi.ContainerListDto{}
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

// GetContainerDetails gets the details for a given container ID
// It allows to retrieve all the technical data included in a container.
// See https://bmw-cardata.bmwgroup.com/customer/public/api-specification#operations-Containers-getContainerDetails
func (c *Client) GetContainerDetails(ctx context.Context, containerID string) (*cardataapi.ContainerDetailsDto, error) {
	resp, err := c.carDataAPI.GetContainerDetails(ctx, containerID, &cardataapi.GetContainerDetailsParams{XVersion: "v1"})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		data := cardataapi.ContainerDetailsDto{}
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

// CreateContainer creates a new container to pack many technical descriptors.
// See https://bmw-cardata.bmwgroup.com/customer/public/api-specification#operations-Containers-createContainer
func (c *Client) CreateContainer(ctx context.Context, name, purpose string, containers []Descriptor) (*cardataapi.CreateContainerResponse, error) {
	opts := &cardataapi.CreateContainerJSONRequestBody{}
	opts.Name = &name
	opts.Purpose = &purpose
	descriptors := make([]string, len(containers))
	for i, container := range containers {
		descriptors[i] = container.ID
	}
	opts.TechnicalDescriptors = &descriptors
	resp, err := c.carDataAPI.CreateContainer(ctx, *opts, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("X-Version", "v1")
		return nil
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		data := cardataapi.CreateContainerResponse{}
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

// DeleteContainer deletes a container
// See https://bmw-cardata.bmwgroup.com/customer/public/api-specification#operations-Containers-deleteContainer
// BUG(tjamet): DeleteContainer is not working. It always returns a 400 error and needs to be investigated and fixed.
func (c *Client) DeleteContainer(ctx context.Context, containerID string) (*cardataapi.DeleteContainerResponse, error) {
	resp, err := c.carDataAPI.DeleteContainer(ctx, containerID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNoContent:
		// No body on deletion success
		return &cardataapi.DeleteContainerResponse{}, nil
	default:
		data := cardataapi.CarDataError{}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, err
		}
		return nil, &data
	}
}
