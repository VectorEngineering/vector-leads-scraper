package database

import (
	"errors"
)

var (
	// ErrOrganizationDoesNotExist is returned when an organization is not found in the database
	ErrOrganizationDoesNotExist = errors.New("organization does not exist")

	// ErrTenantDoesNotExist is returned when a tenant is not found in the database
	ErrTenantDoesNotExist = errors.New("tenant does not exist")
) 