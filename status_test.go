package epplib

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusTextExist(t *testing.T) {
	t.Parallel()

	// Make sure that we have a status text for every code.
	codes := []int{
		StatusSuccess,
		StatusActionPending,
		StatusNoMessage,
		StatusAckToDequeue,
		StatusEndingSession,
		StatusUnknownCommand,
		StatusCommandSyntaxError,
		StatusCommandUseError,
		StatusMissingParameter,
		StatusValueRangeError,
		StatusValueSyntaxError,
		StatusUnimplementedProtocolVersion,
		StatusUnimplementedCommand,
		StatusUnimplementedOption,
		StatusUnimplementedExtension,
		StatusBillingFailure,
		StatusNotEligibleForRenewal,
		StatusNotEligibleForTransfer,
		StatusAuthenticationError,
		StatusAuthorizationError,
		StatusInvalidAuthorizationInformation,
		StatusObjectPendingTransfer,
		StatusObjectNotPendingTransfer,
		StatusObjectExists,
		StatusObjectDoesNotExist,
		StatusObjectStatusProhibitsOperation,
		StatusObjectAssociationProhibitsOperation,
		StatusParameterPolicyError,
		StatusUnimplementedObjectService,
		StatusDataManagementPolicyViolation,
		StatusCommandFailed,
		StatusCommandFailedClosingConnection,
		StatusAuthenticationErrorClosingConnection,
		StatusSessionLimitExceededClosingConnection,
	}

	for _, code := range codes {
		assert.NotEqual((t), "", StatusText(code), fmt.Sprintf("Code: %d does not have a status text", code))
	}
}
