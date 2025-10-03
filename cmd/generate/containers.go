package main

import bmwcardata "github.com/tjamet/bmw-cardata"

type GetContainersResponse struct {
	Data struct {
		Items       []bmwcardata.Descriptor     `json:"items"`
		Total       int                         `json:"total"`
		Offset      int                         `json:"offset"`
		PageSize    int                         `json:"pageSize"`
		HasNextPage bool                        `json:"hasNextPage"`
		Categories  map[string]CategoryResponse `json:"categories"`
	} `json:"data"`
	ErrorData   *string `json:"errorData,omitempty"`
	Message     string  `json:"message"`
	ProcessUUID string  `json:"processUuid"`
	Status      string  `json:"status"`
	Success     bool    `json:"success"`
}

type CategoryResponse struct {
	Description string `json:"description"`
	Rank        int    `json:"rank"`
}
