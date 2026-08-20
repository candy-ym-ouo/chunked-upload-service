package postgres

import "context"

type DB struct{ URL string }

func Open(url string) (*DB, error)                          { return &DB{URL: url}, nil }
func (d *DB) Ping(context.Context) error                    { return nil }
func (d *DB) Close() error                                  { return nil }
func (d *DB) MigrationVersion(context.Context) (int, error) { return 2, nil }
