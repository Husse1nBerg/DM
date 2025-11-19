package utils

import (
	"errors"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	// ErrInvalidMarinaID is returned when the provided marinaId is not a valid UUID.
	ErrInvalidMarinaID = errors.New("invalid marina ID format")
	// ErrMissingEntityID is returned when the entityId form value is missing.
	ErrMissingEntityID = errors.New("entity ID is required")
	// ErrMissingFile is returned when the multipart file is missing.
	ErrMissingFile = errors.New("file is required")
)

// ParseUploadParams parses common multipart upload params: marinaId, entityId and file.
// Returns the parsed marinaID, entityID, file and file header.
func ParseUploadParams(c echo.Context) (uuid.UUID, string, multipart.File, *multipart.FileHeader, error) {
	marinaIDStr := c.FormValue("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return uuid.Nil, "", nil, nil, ErrInvalidMarinaID
	}

	entityID := c.FormValue("entityId")
	if entityID == "" {
		return uuid.Nil, "", nil, nil, ErrMissingEntityID
	}

	file, header, err := c.Request().FormFile("file")
	if err != nil {
		return uuid.Nil, "", nil, nil, ErrMissingFile
	}

	return marinaID, entityID, file, header, nil
}
