package httpapi

import (
	"github.com/wyw14/cry-046/internal/platform/clock"
	"github.com/wyw14/cry-046/internal/platform/ids"
	"github.com/wyw14/cry-046/internal/platform/notify"
	"github.com/wyw14/cry-046/internal/repository/memory"
	"github.com/wyw14/cry-046/internal/service"
	"go.uber.org/zap"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	s := memory.NewStore()
	c := clock.System{}
	i := &ids.Sequence{}
	p := service.NewProjectService(s, s, c, i)
	a := service.NewAssetService(s, s, s, c, i)
	pa := service.NewPaletteService(s, s, s, s, c, i)
	d := service.NewDeliveryService(s, s, s, s, service.LocalPackageWriter{Root: "tmp"}, s, notify.New(), c, i)
	r := service.NewReviewService(s, s, c, i)
	q := service.NewSearchService(s, s)
	w := service.NewWorkspace(s, c, i)
	g := NewServer(p, a, pa, d, r, q, w, zap.NewNop()).Router()
	req := httptest.NewRequest("GET", "/healthz", nil)
	res := httptest.NewRecorder()
	g.ServeHTTP(res, req)
	if res.Code != 200 {
		t.Fatalf("status %d", res.Code)
	}
}
