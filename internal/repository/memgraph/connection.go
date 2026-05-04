package memgraph

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Connection struct {
	driver neo4j.DriverWithContext
}

func NewConnection(uri, username, password string) (*Connection, error) {
	driver, err := neo4j.NewDriverWithContext(
		uri,
		neo4j.BasicAuth(username, password, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify connectivity: %w", err)
	}

	return &Connection{driver: driver}, nil
}

func NewConnectionFromDriver(driver neo4j.DriverWithContext) *Connection {
	return &Connection{driver: driver}
}

func (c *Connection) Driver() neo4j.DriverWithContext {
	return c.driver
}

func (c *Connection) Close(ctx context.Context) error {
	return c.driver.Close(ctx)
}

func (c *Connection) IsReady(ctx context.Context) error {
	return c.driver.VerifyConnectivity(ctx)
}