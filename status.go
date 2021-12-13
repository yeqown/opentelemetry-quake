package tracing

import "go.opentelemetry.io/otel/codes"

type Code = codes.Code

const (
	// Unset is the default status code.
	Unset Code = 0
	// Error indicates the operation contains an error.
	Error Code = 1
	// OK indicates operation has been validated by an Application developers
	// or Operator to have completed successfully, or contain no error.
	OK Code = 2
)
