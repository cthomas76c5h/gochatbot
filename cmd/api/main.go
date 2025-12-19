package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"

	"gochatbot/internal/httpapi"
	"gochatbot/internal/repo"
	"gochatbot/internal/service"
)

func main() {
	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	tenantRepo := repo.NewTenantRepo(conn)
	tenantSvc := service.NewTenantService(tenantRepo)

	templateRepo := repo.NewTemplateRepo(conn)
	templateSvc := service.NewTemplateService(templateRepo)

	s := httpapi.New(httpapi.Deps{
		TenantSvc:   tenantSvc,
		TemplateSvc: templateSvc,
	})

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, s))
}
