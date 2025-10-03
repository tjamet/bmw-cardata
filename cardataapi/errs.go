package cardataapi

import (
	"strings"
)

type CarDataError struct {
	ExveErrorId  *string `json:"exveErrorId,omitempty"`
	ExveErrorMsg *string `json:"exveErrorMsg,omitempty"`
	ExveErrorRef *string `json:"exveErrorRef,omitempty"`
	ExveNote     *string `json:"exveNote,omitempty"`
}

func (e *CarDataError) Error() string {
	builder := strings.Builder{}
	if e.ExveErrorId != nil {
		builder.WriteString(*e.ExveErrorId)
	}
	if e.ExveErrorMsg != nil {
		if len(builder.String()) > 0 {
			builder.WriteString(": ")
		}
		builder.WriteString(*e.ExveErrorMsg)
	}
	if e.ExveErrorRef != nil {
		if len(builder.String()) > 0 {
			builder.WriteString(": ")
		}
		builder.WriteString(*e.ExveErrorRef)
	}
	if e.ExveNote != nil {
		if len(builder.String()) > 0 {
			builder.WriteString(": ")
		}
		builder.WriteString(*e.ExveNote)
	}
	return builder.String()
}
