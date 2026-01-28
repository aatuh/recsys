package mapper

import (
	"strings"

	"recsys/internal/services/foosvc"
	"recsys/src/specs/types"
)

func CreateFooInput(dto *types.CreateFooDTO) foosvc.CreateInput {
	if dto == nil {
		return foosvc.CreateInput{}
	}
	return foosvc.CreateInput{
		OrgID:     strings.TrimSpace(dto.OrgID),
		Namespace: strings.TrimSpace(dto.Namespace),
		Name:      strings.TrimSpace(dto.Name),
	}
}

func UpdateFooInput(dto *types.UpdateFooDTO, id string) foosvc.UpdateInput {
	out := foosvc.UpdateInput{ID: strings.TrimSpace(id)}
	if dto == nil || dto.Name == nil {
		return out
	}
	name := strings.TrimSpace(*dto.Name)
	out.Name = &name
	return out
}

func FooDTOFromModel(f *foosvc.Foo) types.FooDTO {
	var dto types.FooDTO
	if f == nil {
		return dto
	}
	dto.ID = f.ID
	dto.OrgID = f.OrgID
	dto.Namespace = f.Namespace
	dto.Name = f.Name
	dto.CreatedAt = f.CreatedAt
	dto.UpdatedAt = f.UpdatedAt
	return dto
}
