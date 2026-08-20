package main

import (
	"log"
	"net/http"
	"os"

	"github.com/wyw14/cry-046/internal/application"
	"github.com/wyw14/cry-046/internal/config"
	"github.com/wyw14/cry-046/internal/platform/clock"
	"github.com/wyw14/cry-046/internal/platform/ids"
	"github.com/wyw14/cry-046/internal/platform/notify"
	"github.com/wyw14/cry-046/internal/repository/memory"
	"github.com/wyw14/cry-046/internal/service"
	httpapi "github.com/wyw14/cry-046/internal/transport/http"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logg, _ := zap.NewProduction()
	defer logg.Sync()
	store := memory.NewStore()
	clk := clock.System{}
	seq := &ids.Sequence{}
	inbox := notify.New()
	writer := service.LocalPackageWriter{Root: cfg.StorageRoot}
	projects := service.NewProjectService(store, store, clk, seq)
	assets := service.NewAssetService(store, store, store, clk, seq)
	palettes := service.NewPaletteService(store, store, store, store, clk, seq)
	deliveries := service.NewDeliveryService(store, store, store, store, writer, store, inbox, clk, seq)
	reviews := service.NewReviewService(store, store, clk, seq)
	search := service.NewSearchService(store, store)
	workspace := service.NewWorkspace(store, clk, seq)
	srv := httpapi.NewServer(projects, assets, palettes, deliveries, reviews, search, workspace, logg)
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: srv.Router()}
	log.Printf("palette server listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logg.Error("server stopped")
		os.Exit(1)
	}
}

var _ application.Clock = clock.System{}
