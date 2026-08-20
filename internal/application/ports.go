package application

import (
	"context"
	"time"

	"github.com/wyw14/cry-046/internal/domain"
)

type ProjectRepository interface {
	CreateProject(context.Context, domain.Project) error
	GetProject(context.Context, string) (domain.Project, error)
	UpdateProject(context.Context, domain.Project, int) error
	ListProjects(context.Context, string, int, int) ([]domain.Project, int, error)
}

type AssetRepository interface {
	CreateAsset(context.Context, domain.Asset) error
	GetAsset(context.Context, string) (domain.Asset, error)
	ListAssets(context.Context, string) ([]domain.Asset, error)
	UpdateAsset(context.Context, domain.Asset, int) error
}

type PaletteRepository interface {
	CreatePalette(context.Context, domain.Palette) error
	GetPalette(context.Context, string) (domain.Palette, error)
	UpdatePalette(context.Context, domain.Palette, int) error
	ListPalettes(context.Context, string) ([]domain.Palette, error)
}

type DeliveryRepository interface {
	CreateDelivery(context.Context, domain.DeliveryRequest) error
	GetDelivery(context.Context, string) (domain.DeliveryRequest, error)
	UpdateDelivery(context.Context, domain.DeliveryRequest, int) error
	ListDeliveries(context.Context, string) ([]domain.DeliveryRequest, error)
}

type AuditRepository interface {
	Append(context.Context, domain.AuditEvent) error
	List(context.Context, string, int) ([]domain.AuditEvent, error)
}
type Clock interface{ Now() time.Time }
type IDGenerator interface{ NewID(prefix string) string }
type Notifier interface {
	Notify(context.Context, string, string) error
}
type DeliveryWriter interface {
	WritePackage(context.Context, domain.Palette, []domain.Asset, string) (string, error)
}
