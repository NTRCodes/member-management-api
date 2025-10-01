package members

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestMemberRepository_GetMember_UBCLogic(t *testing.T) {
	// Create a fake repository that captures what query would be built
	repo := &testMemberRepository{
		capturedQueries: make([]testQuery, 0),
	}

	// Test 8-digit number (should add "U" prefix)
	_, _ = repo.GetMember(context.Background(), "12345678", false)

	// Check that it tried to query with "U12345678"
	if len(repo.capturedQueries) != 1 {
		t.Fatalf("Expected 1 query, got %d", len(repo.capturedQueries))
	}

	query := repo.capturedQueries[0]
	if query.column != "ubcid" {
		t.Errorf("Expected column 'ubcid', got '%s'", query.column)
	}
	if query.value != "U12345678" {
		t.Errorf("Expected value 'U12345678', got '%s'", query.value)
	}
}

func TestMemberRepository_GetMember_SSNLogic(t *testing.T) {
	repo := &testMemberRepository{
		capturedQueries: make([]testQuery, 0),
	}

	// Test 9-digit SSN (should use "ssn" column)
	_, _ = repo.GetMember(context.Background(), "123456789", false)

	if len(repo.capturedQueries) != 1 {
		t.Fatalf("Expected 1 query, got %d", len(repo.capturedQueries))
	}

	query := repo.capturedQueries[0]
	if query.column != "ssn" {
		t.Errorf("Expected column 'ssn', got '%s'", query.column)
	}
	if query.value != "123456789" {
		t.Errorf("Expected value '123456789', got '%s'", query.value)
	}
}

// Test helper structs
type testQuery struct {
	column string
	value  string
}

type testMemberRepository struct {
	capturedQueries []testQuery
}

func (t *testMemberRepository) GetMember(ctx context.Context, memNum string, fullSSN bool) (*Member, error) {
	// Copy the logic from the real repository, but capture instead of querying
	var column string

	if len(memNum) == 8 {
		memNum = "U" + memNum
		column = "ubcid"
	} else if len(memNum) == 9 && strings.HasPrefix(memNum, "U") {
		column = "ubcid"
	} else if len(memNum) == 9 {
		column = "ssn"
	} else {
		return nil, fmt.Errorf("invalid member format %s", memNum)
	}

	// Capture what would be queried
	t.capturedQueries = append(t.capturedQueries, testQuery{
		column: column,
		value:  memNum,
	})

	// Return fake member for testing
	return &Member{UBCID: memNum}, nil
}

// Test for ValidateMemberNumber - I've set up the structure, you implement the assertions
func TestValidateMemberNumber(t *testing.T) {
	testCases := []struct {
		name             string
		input            string
		shouldError      bool
		expectedLocation string // "length", "format", "input", or ""
		expectedReason   string // Part of the error message to check for
	}{
		{
			name:        "valid 8-digit UBC",
			input:       "12345678",
			shouldError: false,
		},
		{
			name:        "valid 9-digit UBC with U",
			input:       "U87654321",
			shouldError: false,
		},
		{
			name:             "invalid length - too short",
			input:            "1234567",
			shouldError:      true,
			expectedLocation: "length",
			expectedReason:   "must be 8 or 9",
		},
		{
			name:             "invalid length - too long",
			input:            "1234567890",
			shouldError:      true,
			expectedLocation: "length",
			expectedReason:   "must be 8 or 9",
		},
		{
			name:             "invalid format - 9 digits without U",
			input:            "123456789",
			shouldError:      true,
			expectedLocation: "format",
			expectedReason:   "9-digit numbers must start with 'U'",
		},
		{
			name:             "invalid input - non-numeric",
			input:            "invalid8",
			shouldError:      true,
			expectedLocation: "input",
			expectedReason:   "non-numeric",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckValidMemberNumber(tc.input)
			// YOUR TASK: Implement these test assertions
			// Check if error expectation matches
			if tc.shouldError && err == nil {
				t.Errorf("Expected error for input %v; got nil", tc.input)
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Expected no error for input '%v' got: '%v'", tc.input, err)
			}

			//If we expected an error, check the details
			if tc.shouldError && err != nil {
				if !IsInvalidMemberNumber(err) {
					t.Errorf("Expected an InvalidMemberNumber error; got %v", err)
				} else {
					var customErr InvalidMemberNumberError
					if errors.As(err, &customErr) {
						if customErr.Location != tc.expectedLocation {
							t.Errorf("Expected location '%v', got '%s'", tc.expectedLocation, customErr.Location)
						}
						if !strings.Contains(customErr.Reason, tc.expectedReason) {
							t.Errorf("Expected reason to contain '%v'; got '%v'", tc.expectedReason, customErr.Reason)
						}
					}
				}
			}
		})
	}
}
