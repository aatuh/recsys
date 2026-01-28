package validation

import (
	"errors"
	"strings"

	"recsys/src/specs/types"
)

func ValidateCreateFoo(dto *types.CreateFooDTO) error {
	if dto == nil {
		return errors.New("payload required")
	}
	if strings.TrimSpace(dto.OrgID) == "" {
		return errors.New("org_id is required")
	}
	if strings.TrimSpace(dto.Namespace) == "" {
		return errors.New("namespace is required")
	}
	if strings.TrimSpace(dto.Name) == "" {
		return errors.New("name is required")
	}
	return nil
}

func ValidateUpdateFoo(dto *types.UpdateFooDTO) error {
	if dto == nil {
		return errors.New("payload required")
	}
	if dto.Name == nil {
		return errors.New("name is required")
	}
	if strings.TrimSpace(*dto.Name) == "" {
		return errors.New("name cannot be empty")
	}
	return nil
}
