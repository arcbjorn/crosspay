package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type TimeSeriesDB struct {
	db *sql.DB
}

type MetricPoint struct {
	Timestamp time.Time   `json:"timestamp"`
	Metric    string      `json:"metric"`
	Value     float64     `json:"value"`
	Tags      map[string]string `json:"tags"`
}

type QueryOptions struct {
	Start      time.Time
	End        time.Time
	Interval   time.Duration
	Aggregation string // avg, sum, min, max
	Tags       map[string]string
}

func NewTimeSeriesDB(connectionString string) (*TimeSeriesDB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	tsdb := &TimeSeriesDB{db: db}
	
	if err := tsdb.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return tsdb, nil
}

func (ts *TimeSeriesDB) createTables() error {
	createMetricsTable := `
	CREATE TABLE IF NOT EXISTS metrics (
		id BIGSERIAL PRIMARY KEY,
		timestamp TIMESTAMPTZ NOT NULL,
		metric_name VARCHAR(255) NOT NULL,
		value DOUBLE PRECISION NOT NULL,
		tags JSONB,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp);
	CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(metric_name);
	CREATE INDEX IF NOT EXISTS idx_metrics_name_timestamp ON metrics(metric_name, timestamp);
	CREATE INDEX IF NOT EXISTS idx_metrics_tags ON metrics USING GIN(tags);

	-- Create hypertable for TimescaleDB (if available)
	-- This will fail gracefully if TimescaleDB is not installed
	DO $$ 
	BEGIN
		PERFORM create_hypertable('metrics', 'timestamp', if_not_exists => TRUE);
	EXCEPTION 
		WHEN OTHERS THEN NULL;
	END $$;
	`

	_, err := ts.db.Exec(createMetricsTable)
	return err
}

func (ts *TimeSeriesDB) WritePoint(ctx context.Context, point MetricPoint) error {
	query := `
		INSERT INTO metrics (timestamp, metric_name, value, tags)
		VALUES ($1, $2, $3, $4)
	`

	var tags interface{}
	if len(point.Tags) > 0 {
		tagsJSON, err := json.Marshal(point.Tags)
		if err != nil {
			return fmt.Errorf("failed to marshal tags: %w", err)
		}
		tags = string(tagsJSON)
	}

	_, err := ts.db.ExecContext(ctx, query, point.Timestamp, point.Metric, point.Value, tags)
	return err
}

func (ts *TimeSeriesDB) WriteBatch(ctx context.Context, points []MetricPoint) error {
	if len(points) == 0 {
		return nil
	}

	tx, err := ts.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics (timestamp, metric_name, value, tags)
		VALUES ($1, $2, $3, $4)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, point := range points {
		var tags interface{}
		if len(point.Tags) > 0 {
			tagsJSON, err := json.Marshal(point.Tags)
			if err != nil {
				return fmt.Errorf("failed to marshal tags: %w", err)
			}
			tags = string(tagsJSON)
		}

		_, err := stmt.ExecContext(ctx, point.Timestamp, point.Metric, point.Value, tags)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	return tx.Commit()
}

func (ts *TimeSeriesDB) Query(ctx context.Context, metric string, opts QueryOptions) ([]MetricPoint, error) {
	baseQuery := `
		SELECT timestamp, metric_name, value, tags
		FROM metrics 
		WHERE metric_name = $1 
		AND timestamp >= $2 
		AND timestamp <= $3
	`

	args := []interface{}{metric, opts.Start, opts.End}
	argIndex := 3

	// Add tag filters
	tagFilters := ""
	for key, value := range opts.Tags {
		argIndex++
		tagFilters += fmt.Sprintf(" AND tags->>'%s' = $%d", key, argIndex)
		args = append(args, value)
	}

	query := baseQuery + tagFilters

	// Add aggregation and grouping
	if opts.Aggregation != "" && opts.Interval > 0 {
		intervalSeconds := int(opts.Interval.Seconds())
		query = fmt.Sprintf(`
			SELECT 
				date_trunc('second', timestamp) + 
				INTERVAL '%d seconds' * (EXTRACT(epoch FROM timestamp)::int / %d) as timestamp,
				metric_name,
				%s(value) as value,
				tags
			FROM (%s) t
			GROUP BY 1, 2, 4
			ORDER BY 1
		`, intervalSeconds, intervalSeconds, opts.Aggregation, query)
	} else {
		query += " ORDER BY timestamp"
	}

	rows, err := ts.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var points []MetricPoint
	for rows.Next() {
		var point MetricPoint
		var tagsJSON sql.NullString

		err := rows.Scan(&point.Timestamp, &point.Metric, &point.Value, &tagsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if tagsJSON.Valid {
			if err := json.Unmarshal([]byte(tagsJSON.String), &point.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
			}
		}

		points = append(points, point)
	}

	return points, rows.Err()
}

func (ts *TimeSeriesDB) GetLatest(ctx context.Context, metric string, tags map[string]string) (*MetricPoint, error) {
	opts := QueryOptions{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
		Tags:  tags,
	}

	points, err := ts.Query(ctx, metric, opts)
	if err != nil {
		return nil, err
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("no data points found")
	}

	// Return the latest point
	latest := points[len(points)-1]
	return &latest, nil
}

func (ts *TimeSeriesDB) DeleteOldData(ctx context.Context, olderThan time.Time) error {
	query := `DELETE FROM metrics WHERE timestamp < $1`
	result, err := ts.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		return fmt.Errorf("failed to delete old data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	fmt.Printf("Deleted %d old metric records\n", rowsAffected)
	return nil
}

func (ts *TimeSeriesDB) GetMetrics(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT metric_name FROM metrics ORDER BY metric_name`
	
	rows, err := ts.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	defer rows.Close()

	var metrics []string
	for rows.Next() {
		var metric string
		if err := rows.Scan(&metric); err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, metric)
	}

	return metrics, rows.Err()
}

func (ts *TimeSeriesDB) GetStats(ctx context.Context) (map[string]interface{}, error) {
	queries := map[string]string{
		"total_points": "SELECT COUNT(*) FROM metrics",
		"earliest_timestamp": "SELECT MIN(timestamp) FROM metrics",
		"latest_timestamp": "SELECT MAX(timestamp) FROM metrics",
		"unique_metrics": "SELECT COUNT(DISTINCT metric_name) FROM metrics",
	}

	stats := make(map[string]interface{})
	
	for name, query := range queries {
		var result interface{}
		err := ts.db.QueryRowContext(ctx, query).Scan(&result)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get stat %s: %w", name, err)
		}
		stats[name] = result
	}

	return stats, nil
}

func (ts *TimeSeriesDB) Close() error {
	return ts.db.Close()
}

// Helper function to create common metric points
func NewValidatorMetricPoint(address string, metric string, value float64) MetricPoint {
	return MetricPoint{
		Timestamp: time.Now(),
		Metric:    fmt.Sprintf("validator.%s", metric),
		Value:     value,
		Tags: map[string]string{
			"validator_address": address,
		},
	}
}

func NewVaultMetricPoint(tranche string, metric string, value float64) MetricPoint {
	return MetricPoint{
		Timestamp: time.Now(),
		Metric:    fmt.Sprintf("vault.%s", metric),
		Value:     value,
		Tags: map[string]string{
			"tranche": tranche,
		},
	}
}

func NewPaymentMetricPoint(metric string, value float64, paymentType string) MetricPoint {
	return MetricPoint{
		Timestamp: time.Now(),
		Metric:    fmt.Sprintf("payment.%s", metric),
		Value:     value,
		Tags: map[string]string{
			"payment_type": paymentType,
		},
	}
}