package database

import (
	"context"
	"fmt"
	"time"

	"DP_Maintenance/pkg/config"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func ConnectMemgraph(cfg *config.Config) (neo4j.DriverWithContext, error) {
	driver, err := neo4j.NewDriverWithContext(
		cfg.MemgraphURI,
		neo4j.BasicAuth(cfg.MemgraphUser, cfg.MemgraphPass, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Memgraph driver: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify Memgraph connectivity: %w", err)
	}

	return driver, nil
}

const (
	CreateConstraintsQuery = `
	CREATE CONSTRAINT dataset_urn_unique IF NOT EXISTS
	FOR (d:Dataset) REQUIRE d.urn IS UNIQUE;

	CREATE INDEX dataset_name_index IF NOT EXISTS
	FOR (d:Dataset) ON (d.name);

	CREATE INDEX column_id_index IF NOT EXISTS
	FOR (c:Column) ON (c.id);

	CREATE INDEX lineage_source_index IF NOT EXISTS
	FOR ()-[r:DERIVED_FROM]->() ON (r.source_id);

	CREATE INDEX lineage_target_index IF NOT EXISTS
	FOR ()-[r:DERIVED_FROM]->() ON (r.target_id);
	`
)

func SetupMemgraphConstraints(driver neo4j.DriverWithContext) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := neo4j.ExecuteQuery(ctx, driver, CreateConstraintsQuery, nil, neo4j.EagerResultTransformer)
	if err != nil {
		return fmt.Errorf("failed to setup Memgraph constraints: %w", err)
	}

	fmt.Println("Memgraph constraints setup complete")
	return nil
}