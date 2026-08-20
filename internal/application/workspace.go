package application

import (
	"context"
	"github.com/wyw14/cry-046/internal/domain"
)

type WorkspaceService interface {
	ToggleFavorite(context.Context, string, string) (bool, error)
	RecordVisit(context.Context, string, string) error
	ListFavorites(context.Context, string) ([]domain.Favorite, error)
	ListRecent(context.Context, string, int) ([]domain.RecentVisit, error)
	ListTodos(context.Context, string, bool) ([]domain.Todo, error)
	CompleteTodo(context.Context, string, string) error
}
