package epplib

// EPP result codes as described in https://datatracker.ietf.org/doc/html/rfc5730#section-3
const (
	StatusSuccess       = 1000
	StatusActionPending = 1001
	StatusNoMessage     = 1300
	StatusAckToDequeue  = 1301
	StatusEndingSession = 1500

	StatusUnknownCommand     = 2000
	StatusCommandSyntaxError = 2001
	StatusCommandUseError    = 2002
	StatusMissingParameter   = 2003
	StatusValueRangeError    = 2004
	StatusValueSyntaxError   = 2005

	StatusUnimplementedProtocolVersion = 2100
	StatusUnimplementedCommand         = 2101
	StatusUnimplementedOption          = 2102
	StatusUnimplementedExtension       = 2103
	StatusBillingFailure               = 2104
	StatusNotEligibleForRenewal        = 2105
	StatusNotEligibleForTransfer       = 2106

	StatusAuthenticationError             = 2200
	StatusAuthorizationError              = 2201
	StatusInvalidAuthorizationInformation = 2202

	StatusObjectPendingTransfer               = 2300
	StatusObjectNotPendingTransfer            = 2301
	StatusObjectExists                        = 2302
	StatusObjectDoesNotExist                  = 2303
	StatusObjectStatusProhibitsOperation      = 2304
	StatusObjectAssociationProhibitsOperation = 2305
	StatusParameterPolicyError                = 2306
	StatusUnimplementedObjectService          = 2307
	StatusDataManagementPolicyViolation       = 2308

	StatusCommandFailed                         = 2400
	StatusCommandFailedClosingConnection        = 2500
	StatusAuthenticationErrorClosingConnection  = 2501
	StatusSessionLimitExceededClosingConnection = 2502
)

// EPP result codes strings as described in https://datatracker.ietf.org/doc/html/rfc5730#section-3
var codeText = map[int]string{
	StatusSuccess:       "Command completed successfully",
	StatusActionPending: "Command completed successfully; action pending",
	StatusNoMessage:     "Command completed successfully; no messages",
	StatusAckToDequeue:  "Command completed successfully; ack to dequeue",
	StatusEndingSession: "Command completed successfully; ending session",

	StatusUnknownCommand:     "Unknown command",
	StatusCommandSyntaxError: "Command syntax error",
	StatusCommandUseError:    "Command use error",
	StatusMissingParameter:   "Required parameter missing",
	StatusValueRangeError:    "Parameter value range error",
	StatusValueSyntaxError:   "Parameter value syntax error",

	StatusUnimplementedProtocolVersion: "Unimplemented protocol version",
	StatusUnimplementedCommand:         "Unimplemented command",
	StatusUnimplementedOption:          "Unimplemented option",
	StatusUnimplementedExtension:       "Unimplemented extension",
	StatusBillingFailure:               "Billing failure",
	StatusNotEligibleForRenewal:        "Object is not eligible for renewal",
	StatusNotEligibleForTransfer:       "Object is not eligible for transfer",

	StatusAuthenticationError:                 "Authentication error",
	StatusAuthorizationError:                  "Authorization error",
	StatusInvalidAuthorizationInformation:     "Invalid authorization information",
	StatusObjectPendingTransfer:               "Object pending transfer",
	StatusObjectNotPendingTransfer:            "Object not pending transfer",
	StatusObjectExists:                        "Object exists",
	StatusObjectDoesNotExist:                  "Object does not exist",
	StatusObjectStatusProhibitsOperation:      "Object status prohibits operation",
	StatusObjectAssociationProhibitsOperation: "Object association prohibits operation",
	StatusParameterPolicyError:                "Parameter value policy error",
	StatusUnimplementedObjectService:          "Unimplemented object service",
	StatusDataManagementPolicyViolation:       "Data management policy violation",

	StatusCommandFailed:                         "Command failed",
	StatusCommandFailedClosingConnection:        "Command failed; server closing connection",
	StatusAuthenticationErrorClosingConnection:  "Authentication error; server closing connection",
	StatusSessionLimitExceededClosingConnection: "Session limit exceeded; server closing connection",
}

// StatusText returns the EPP status text for the status code.
func StatusText(code int) string {
	return codeText[code]
}
