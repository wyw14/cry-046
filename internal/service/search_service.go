package service

import (
	"context"
	"sort"
	"strings"

	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/domain"
)

type SearchService struct {
	projects application.ProjectRepository
	palettes application.PaletteRepository
}

func NewSearchService(p application.ProjectRepository, ps application.PaletteRepository) *SearchService {
	return &SearchService{projects: p, palettes: ps}
}

type SearchResult struct {
	Projects []domain.Project `json:"projects"`
	Palettes []domain.Palette `json:"palettes"`
}

func (s *SearchService) Search(ctx context.Context, q string, page, size int) (SearchResult, error) {
	ps, _, err := s.projects.ListProjects(ctx, q, page, size)
	if err != nil {
		return SearchResult{}, err
	}
	out := SearchResult{Projects: ps, Palettes: []domain.Palette{}}
	for _, p := range ps {
		pal, err := s.palettes.ListPalettes(ctx, p.ID)
		if err != nil {
			return SearchResult{}, err
		}
		for _, x := range pal {
			if q == "" || strings.Contains(strings.ToLower(x.Name+" "+x.Source), strings.ToLower(q)) {
				out.Palettes = append(out.Palettes, x)
			}
		}
	}
	sort.Slice(out.Palettes, func(i, j int) bool { return out.Palettes[i].CreatedAt.After(out.Palettes[j].CreatedAt) })
	return out, nil
}
