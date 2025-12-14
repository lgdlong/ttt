package dto

import "time"

// ============================================================
// Standardized API Response Structure
// ============================================================
// Reference: Backend Development Skill - API Design Best Practices
//
// Design Goals:
// 1. CLARITY: Response clearly indicates success/failure/partial success
// 2. CONSISTENCY: All endpoints use same response structure
// 3. ACTIONABILITY: Errors provide enough detail for client to act
// 4. DEBUGGABILITY: Include request_id and metadata for tracing
//
// Response Categories:
// - SUCCESS: Operation completed successfully
// - CREATED: Resource created successfully (HTTP 201)
// - MERGED: Operation succeeded but resource was merged with existing (tag deduplication)
// - ERROR: Operation failed with clear reason
// ============================================================

// ResponseStatus represents the outcome of an API request
type ResponseStatus string

const (
	// Success statuses
	StatusSuccess ResponseStatus = "success" // Operation completed successfully
	StatusCreated ResponseStatus = "created" // New resource created (HTTP 201)
	StatusMerged  ResponseStatus = "merged"  // Resource auto-merged with existing (tag deduplication)

	// Error statuses
	StatusError    ResponseStatus = "error"     // Generic error
	StatusNotFound ResponseStatus = "not_found" // Resource not found (HTTP 404)
	StatusConflict ResponseStatus = "conflict"  // Business logic conflict (HTTP 409)
	StatusInvalid  ResponseStatus = "invalid"   // Invalid input (HTTP 400)
)

// APIResponse is a standardized wrapper for all API responses
type APIResponse struct {
	Status    ResponseStatus `json:"status"`               // success/created/merged/error/conflict
	Message   string         `json:"message"`              // Human-readable message
	Data      interface{}    `json:"data,omitempty"`       // Response payload (nil for errors)
	Error     *ErrorDetail   `json:"error,omitempty"`      // Error details (nil for success)
	Metadata  *Metadata      `json:"metadata,omitempty"`   // Additional context
	Timestamp time.Time      `json:"timestamp"`            // Response timestamp (ISO 8601)
	RequestID string         `json:"request_id,omitempty"` // For distributed tracing
}

// ErrorDetail provides structured error information
type ErrorDetail struct {
	Code    string      `json:"code"`              // Machine-readable error code (e.g., "TAG_DUPLICATE")
	Message string      `json:"message"`           // Human-readable error message
	Field   *string     `json:"field,omitempty"`   // Field name for validation errors
	Details interface{} `json:"details,omitempty"` // Additional context (e.g., suggestions)
}

// Metadata contains additional response context
type Metadata struct {
	// Tag operation metadata
	IsNewResource   *bool    `json:"is_new_resource,omitempty"`  // true if resource was created
	AutoMerged      *bool    `json:"auto_merged,omitempty"`      // true if auto-merged with existing
	MergedInto      *string  `json:"merged_into,omitempty"`      // ID of resource merged into
	OriginalInput   *string  `json:"original_input,omitempty"`   // User's original input
	CanonicalName   *string  `json:"canonical_name,omitempty"`   // Canonical/normalized name
	TranslationUsed *bool    `json:"translation_used,omitempty"` // true if Translation Layer was used
	TranslatedFrom  *string  `json:"translated_from,omitempty"`  // Original language input
	TranslatedTo    *string  `json:"translated_to,omitempty"`    // Translated English term
	MatchedVia      *string  `json:"matched_via,omitempty"`      // How match was found: "exact"|"translation"|"vector"|"new"
	SimilarityScore *float64 `json:"similarity_score,omitempty"` // Vector similarity score (0.0-1.0)
	LayerHit        *int     `json:"layer_hit,omitempty"`        // Which resolution layer matched (1/1.5/2/3/4)

	// Performance metadata
	ProcessingTimeMs *int64 `json:"processing_time_ms,omitempty"` // Request processing time

	// Pagination metadata (for list endpoints)
	Pagination *PaginationMetadata `json:"pagination,omitempty"`
}

// ============================================================
// Tag-Specific Response Types
// ============================================================

// TagResolveResponse represents the result of tag resolution (CreateCanonicalTag)
type TagResolveResponse struct {
	ID           string  `json:"id"`                      // Canonical tag ID
	Name         string  `json:"name"`                    // Display name (canonical)
	MatchedAlias *string `json:"matched_alias,omitempty"` // Original user input (if different)
}

// TagConflictDetail provides details when similar tags are found
type TagConflictDetail struct {
	ExistingTag TagResponse   `json:"existing_tag"`          // The existing similar tag
	Similarity  float64       `json:"similarity"`            // Similarity percentage (0-100)
	Suggestions []TagResponse `json:"suggestions,omitempty"` // Other similar tags
	Message     string        `json:"message"`               // User-friendly message
}

// ============================================================
// Success Response Helpers
// ============================================================

// NewSuccessResponse creates a standardized success response
func NewSuccessResponse(data interface{}, message string, metadata *Metadata) APIResponse {
	return APIResponse{
		Status:    StatusSuccess,
		Message:   message,
		Data:      data,
		Metadata:  metadata,
		Timestamp: time.Now().UTC(),
	}
}

// NewCreatedResponse creates a standardized "created" response (HTTP 201)
func NewCreatedResponse(data interface{}, message string, metadata *Metadata) APIResponse {
	if metadata == nil {
		metadata = &Metadata{}
	}
	isNew := true
	metadata.IsNewResource = &isNew

	return APIResponse{
		Status:    StatusCreated,
		Message:   message,
		Data:      data,
		Metadata:  metadata,
		Timestamp: time.Now().UTC(),
	}
}

// NewMergedResponse creates a standardized "merged" response (HTTP 200)
// Used when resource was auto-merged with existing (e.g., tag deduplication)
func NewMergedResponse(data interface{}, message string, metadata *Metadata) APIResponse {
	if metadata == nil {
		metadata = &Metadata{}
	}
	isNew := false
	autoMerged := true
	metadata.IsNewResource = &isNew
	metadata.AutoMerged = &autoMerged

	return APIResponse{
		Status:    StatusMerged,
		Message:   message,
		Data:      data,
		Metadata:  metadata,
		Timestamp: time.Now().UTC(),
	}
}

// ============================================================
// Error Response Helpers
// ============================================================

// NewErrorResponse creates a standardized error response
func NewErrorResponse(status ResponseStatus, code string, message string, details interface{}) APIResponse {
	return APIResponse{
		Status:  status,
		Message: message,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC(),
	}
}

// NewNotFoundResponse creates a standardized "not found" error (HTTP 404)
func NewNotFoundResponse(resource string, id string) APIResponse {
	return NewErrorResponse(
		StatusNotFound,
		"RESOURCE_NOT_FOUND",
		resource+" not found",
		map[string]string{"resource": resource, "id": id},
	)
}

// NewConflictResponse creates a standardized "conflict" error (HTTP 409)
func NewConflictResponse(code string, message string, details interface{}) APIResponse {
	return NewErrorResponse(StatusConflict, code, message, details)
}

// NewValidationErrorResponse creates a standardized validation error (HTTP 400)
func NewValidationErrorResponse(field string, message string) APIResponse {
	return APIResponse{
		Status:  StatusInvalid,
		Message: "Validation failed",
		Error: &ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Field:   &field,
		},
		Timestamp: time.Now().UTC(),
	}
}

// NewInternalErrorResponse creates a standardized internal error (HTTP 500)
func NewInternalErrorResponse(message string) APIResponse {
	return NewErrorResponse(
		StatusError,
		"INTERNAL_ERROR",
		message,
		nil,
	)
}

// ============================================================
// Response Metadata Builders
// ============================================================

// NewTagMetadata creates metadata for tag operations
func NewTagMetadata(opts TagMetadataOptions) *Metadata {
	return &Metadata{
		IsNewResource:    opts.IsNewResource,
		AutoMerged:       opts.AutoMerged,
		MergedInto:       opts.MergedInto,
		OriginalInput:    opts.OriginalInput,
		CanonicalName:    opts.CanonicalName,
		TranslationUsed:  opts.TranslationUsed,
		TranslatedFrom:   opts.TranslatedFrom,
		TranslatedTo:     opts.TranslatedTo,
		MatchedVia:       opts.MatchedVia,
		SimilarityScore:  opts.SimilarityScore,
		LayerHit:         opts.LayerHit,
		ProcessingTimeMs: opts.ProcessingTimeMs,
	}
}

// TagMetadataOptions contains options for building tag metadata
type TagMetadataOptions struct {
	IsNewResource    *bool
	AutoMerged       *bool
	MergedInto       *string
	OriginalInput    *string
	CanonicalName    *string
	TranslationUsed  *bool
	TranslatedFrom   *string
	TranslatedTo     *string
	MatchedVia       *string
	SimilarityScore  *float64
	LayerHit         *int
	ProcessingTimeMs *int64
}
