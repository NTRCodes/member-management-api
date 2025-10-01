package members

import (
	"NTRCodes/member-api/internal/database"
	"NTRCodes/member-api/internal/privacy"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// UpdateMemberRequest represents the request body for updating a member
type UpdateMemberRequest struct {
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	Email       *string `json:"email,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Address     *string `json:"address,omitempty"`
	City        *string `json:"city,omitempty"`
	State       *string `json:"state,omitempty"`
	ZipCode     *string `json:"zip_code,omitempty"`
}

// InvalidMemberNumberError represents validation errors for member numbers
type InvalidMemberNumberError struct {
	Input    string
	Reason   string
	Location string // Where the validation failed: "input", "format", "length"
}

func (e InvalidMemberNumberError) Error() string {
	return fmt.Sprintf("invalid member number '%s' at %s: %s", e.Input, e.Location, e.Reason)
}

// IsInvalidMemberNumber checks if error is InvalidMemberNumberError
func IsInvalidMemberNumber(err error) bool {
	var invalidMemberNumberError InvalidMemberNumberError
	ok := errors.As(err, &invalidMemberNumberError)
	return ok
}

// isNumeric checks if a string contains only numeric characters
func isNumeric(input string) bool {
	_, err := strconv.Atoi(input)
	return err == nil
}

// CheckValidMemberNumber validates UBC member number format
func CheckValidMemberNumber(input string) error {
	if len(input) < 8 || len(input) > 9 {
		return InvalidMemberNumberError{
			Input:    input,
			Reason:   "Input length must be 8 or 9 characters long",
			Location: "length",
		}
	}

	if len(input) == 8 {
		if !isNumeric(input) {
			return InvalidMemberNumberError{
				Input:    input,
				Reason:   "Input is non-numeric",
				Location: "input",
			}
		}
	}

	if len(input) == 9 {
		if input[0] == 'U' {
			// UBC ID with U prefix
			if !isNumeric(input[1:]) {
				return InvalidMemberNumberError{
					Input:    input,
					Reason:   "9-digit numbers starting with 'U' must be followed by numbers",
					Location: "input",
				}
			}
		} else {
			// SSN - must be all numeric
			if !isNumeric(input) {
				return InvalidMemberNumberError{
					Input:    input,
					Reason:   "9-digit SSN must be all numeric",
					Location: "input",
				}
			}
		}
	}

	return nil
}

type MemberRepository interface {
	GetMember(ctx context.Context, memNum string, fullSSN bool) (*Member, error)
	CreateMember(ctx context.Context, member *Member) error
	UpdateMember(ctx context.Context, id string, updateReq UpdateMemberRequest) error
	DeleteMember(ctx context.Context, id string) error

	// Batch operations for high-performance file processing
	BatchUpsertMembers(ctx context.Context, members []*Member) (*BatchUpsertResult, error)
	BatchCreateMembers(ctx context.Context, members []*Member) (*BatchCreateResult, error)
}

// BatchUpsertResult contains the results of a batch upsert operation
type BatchUpsertResult struct {
	TotalProcessed int
	Created        int
	Updated        int
	Failed         int
	Errors         []BatchError
}

// BatchCreateResult contains the results of a batch create operation
type BatchCreateResult struct {
	TotalProcessed int
	Created        int
	Failed         int
	Errors         []BatchError
}

// BatchError represents an error for a specific record in a batch operation
type BatchError struct {
	Index   int    // Index in the original slice
	MemberID string // UBCID or SSN that failed
	Error   string // Error message
}

type memberRepository struct {
	db *database.DB // Your database connection
}

func NewMemberRepository(db *database.DB) MemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) GetMember(ctx context.Context, memNum string, fullSSN bool) (*Member, error) {
	var query string
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
	query = fmt.Sprintf("SELECT ubcid, ssn, first_name, last_name, address_1, address_2, city, state, zip, phone1,"+
		" phone2, email, local, status, class, dob, gender, init_date, delegate, lang, veteran,"+
		" stamp FROM members WHERE %s = $1", column)
	//query = fmt.Sprintf("Select * from members where %s = $1", column)

	//Debug:
	//fmt.Printf("DEBUG: Query: %s, Value: %s\n", query, memNum)

	row := r.db.QueryRowContext(ctx, query, memNum)

	var member Member

	err := row.Scan(
		&member.UBCID,
		&member.SSN,
		&member.FirstName,
		&member.LastName,
		&member.Address,
		&member.Address2,
		&member.City,
		&member.State,
		&member.Zip,
		&member.Phone,
		&member.Phone2,
		&member.Email,
		&member.Local,
		&member.Status,
		&member.Class,
		&member.DOB,
		&member.Gender,
		&member.InitDate,
		&member.Delegate,
		&member.Language,
		&member.Veteran,
		&member.LastUpdate,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if fullSSN == false && *member.SSN != "" {
		*member.SSN = privacy.LastFourSSN(*member.SSN)
	}

	return &member, nil
}

func (r *memberRepository) CreateMember(ctx context.Context, member *Member) error {
	query := `INSERT INTO members (ubcid, ssn, first_name, last_name, address_1, city, state, zip,
		phone1, email, local, status, class, dob, gender, init_date, delegate, lang, veteran, stamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`

	_, err := r.db.ExecContext(ctx, query,
		member.UBCID,
		member.SSN,
		member.FirstName,
		member.LastName,
		member.Address,
		member.City,
		member.State,
		member.Zip,
		member.Phone,
		member.Email,
		member.Local,
		member.Status,
		member.Class,
		member.DOB,
		member.Gender,
		member.InitDate,
		member.Delegate,
		member.Language,
		member.Veteran,
		member.LastUpdate,
	)

	if err != nil {
		return fmt.Errorf("failed to create member: %w", err)
	}

	return nil
}

func (r *memberRepository) UpdateMember(ctx context.Context, id string, updateReq UpdateMemberRequest) error {
	// Simple approach: update only the fields that are provided
	setParts := []string{}
	args := []interface{}{id}
	argIndex := 2

	if updateReq.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, *updateReq.FirstName)
		argIndex++
	}
	if updateReq.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, *updateReq.LastName)
		argIndex++
	}
	if updateReq.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *updateReq.Email)
		argIndex++
	}
	if updateReq.PhoneNumber != nil {
		setParts = append(setParts, fmt.Sprintf("phone1 = $%d", argIndex))
		args = append(args, *updateReq.PhoneNumber)
		argIndex++
	}
	if updateReq.Address != nil {
		setParts = append(setParts, fmt.Sprintf("address_1 = $%d", argIndex))
		args = append(args, *updateReq.Address)
		argIndex++
	}
	if updateReq.City != nil {
		setParts = append(setParts, fmt.Sprintf("city = $%d", argIndex))
		args = append(args, *updateReq.City)
		argIndex++
	}
	if updateReq.State != nil {
		setParts = append(setParts, fmt.Sprintf("state = $%d", argIndex))
		args = append(args, *updateReq.State)
		argIndex++
	}
	if updateReq.ZipCode != nil {
		setParts = append(setParts, fmt.Sprintf("zip = $%d", argIndex))
		args = append(args, *updateReq.ZipCode)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE members SET %s WHERE ubcid = $1 OR ssn = $1",
		strings.Join(setParts, ", "))

	result, err := r.db.ExecContext(ctx, query, args...)

	if err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

func (r *memberRepository) DeleteMember(ctx context.Context, id string) error {
	query := "DELETE FROM members WHERE ubcid = $1 OR ssn = $1"
	result, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

// BatchUpsertMembers performs high-performance batch upsert operations
// This is the key method that will solve the 99% bottleneck in the Python system
func (r *memberRepository) BatchUpsertMembers(ctx context.Context, members []*Member) (*BatchUpsertResult, error) {
	if len(members) == 0 {
		return &BatchUpsertResult{}, nil
	}

	result := &BatchUpsertResult{
		TotalProcessed: len(members),
		Errors:         make([]BatchError, 0),
	}

	// Use a transaction for atomicity and performance
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() succeeds

	// PostgreSQL UPSERT using ON CONFLICT
	upsertQuery := `
		INSERT INTO members (ubcid, ssn, first_name, last_name, address_1, city, state, zip,
			phone1, email, local, status, class, dob, gender, init_date, delegate, lang, veteran, stamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		ON CONFLICT (ubcid)
		DO UPDATE SET
			ssn = EXCLUDED.ssn,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			address_1 = EXCLUDED.address_1,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			zip = EXCLUDED.zip,
			phone1 = EXCLUDED.phone1,
			email = EXCLUDED.email,
			local = EXCLUDED.local,
			status = EXCLUDED.status,
			class = EXCLUDED.class,
			dob = EXCLUDED.dob,
			gender = EXCLUDED.gender,
			delegate = EXCLUDED.delegate,
			lang = EXCLUDED.lang,
			veteran = EXCLUDED.veteran,
			stamp = EXCLUDED.stamp
		RETURNING (xmax = 0) AS was_inserted`

	// Prepare the statement for reuse (major performance boost)
	stmt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare upsert statement: %w", err)
	}
	defer stmt.Close()

	// Process each member
	for i, member := range members {
		var wasInserted bool
		err := stmt.QueryRowContext(ctx,
			member.UBCID,
			member.SSN,
			member.FirstName,
			member.LastName,
			member.Address,
			member.City,
			member.State,
			member.Zip,
			member.Phone,
			member.Email,
			member.Local,
			member.Status,
			member.Class,
			member.DOB,
			member.Gender,
			member.InitDate,
			member.Delegate,
			member.Language,
			member.Veteran,
			member.LastUpdate,
		).Scan(&wasInserted)

		if err != nil {
			result.Failed++
			memberID := "unknown"
			if member.UBCID != "" {
				memberID = member.UBCID
			}
			result.Errors = append(result.Errors, BatchError{
				Index:    i,
				MemberID: memberID,
				Error:    err.Error(),
			})
			continue
		}

		if wasInserted {
			result.Created++
		} else {
			result.Updated++
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// BatchCreateMembers performs high-performance batch create operations
func (r *memberRepository) BatchCreateMembers(ctx context.Context, members []*Member) (*BatchCreateResult, error) {
	if len(members) == 0 {
		return &BatchCreateResult{}, nil
	}

	result := &BatchCreateResult{
		TotalProcessed: len(members),
		Errors:         make([]BatchError, 0),
	}

	// Use a transaction for atomicity and performance
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the insert statement
	insertQuery := `INSERT INTO members (ubcid, ssn, first_name, last_name, address_1, city, state, zip,
		phone1, email, local, status, class, dob, gender, init_date, delegate, lang, veteran, stamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Process each member
	for i, member := range members {
		_, err := stmt.ExecContext(ctx,
			member.UBCID,
			member.SSN,
			member.FirstName,
			member.LastName,
			member.Address,
			member.City,
			member.State,
			member.Zip,
			member.Phone,
			member.Email,
			member.Local,
			member.Status,
			member.Class,
			member.DOB,
			member.Gender,
			member.InitDate,
			member.Delegate,
			member.Language,
			member.Veteran,
			member.LastUpdate,
		)

		if err != nil {
			result.Failed++
			memberID := "unknown"
			if member.UBCID != "" {
				memberID = member.UBCID
			}
			result.Errors = append(result.Errors, BatchError{
				Index:    i,
				MemberID: memberID,
				Error:    err.Error(),
			})
			continue
		}

		result.Created++
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}
