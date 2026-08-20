package application

import "github.com/wyw14/cry-046/internal/domain"

type CreateProjectInput struct {
	ID, Name, Customer, Scene, Series, OwnerID string
	Tags                                       []string
	Confidentiality                            domain.Confidentiality
}
type CreateAssetInput struct {
	ID, ProjectID, Name, Filename, Mime, CopyrightNote, LicenseHolder, Role string
	Bytes                                                                   int64
	ExpiresAt                                                               *string
}
type ColorInput struct{ Name, Hex, Source, Replacement, AssetID string }
type CreatePaletteInput struct {
	ID, ProjectID, Name, Source, Branch string
	Entries                             []ColorInput
}
type CreateDeliveryInput struct {
	ID, PaletteID, Applicant, Format string
	ExpiresAt                        string
}
type Page struct {
	Items    any `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}
