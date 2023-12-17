package handler

import "github.com/ockendenjo/osm-pt-validator/pkg/validation"

type CheckRelationEvent struct {
	RelationID int64             `json:"relationID"`
	Config     validation.Config `json:"config"`
}
