package db

import (
	blackholedex "blackholego"
	"fmt"
	"math/big"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// AssetSnapshotRecord represents the database model for CurrentAssetSnapshot
type AssetSnapshotRecord struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	Timestamp    time.Time `gorm:"index;not null"`
	CurrentState int       `gorm:"not null;comment:Strategy phase as integer"`
	TotalValue   string    `gorm:"type:varchar(78);not null;comment:big.Int as string"`
	AmountWavax  string    `gorm:"type:varchar(78);not null;comment:big.Int as string"`
	AmountUsdc   string    `gorm:"type:varchar(78);not null;comment:big.Int as string"`
	AmountBlack  string    `gorm:"type:varchar(78);not null;comment:big.Int as string"`
	AmountAvax   string    `gorm:"type:varchar(78);not null;comment:big.Int as string"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (AssetSnapshotRecord) TableName() string {
	return "asset_snapshots"
}

// MySQLRecorder implements TransactionRecorder interface using GORM and MySQL
type MySQLRecorder struct {
	db *gorm.DB
}

// NewMySQLRecorder creates a new MySQLRecorder instance
// dsn format: "user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
func NewMySQLRecorder(dsn string) (*MySQLRecorder, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&AssetSnapshotRecord{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return &MySQLRecorder{db: db}, nil
}

// NewMySQLRecorderWithDB creates a new MySQLRecorder with an existing GORM DB instance
func NewMySQLRecorderWithDB(db *gorm.DB) (*MySQLRecorder, error) {
	// Auto migrate the schema
	if err := db.AutoMigrate(&AssetSnapshotRecord{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return &MySQLRecorder{db: db}, nil
}

// RecordReport implements TransactionRecorder interface
func (r *MySQLRecorder) RecordReport(snapshot blackholedex.CurrentAssetSnapshot) error {
	record := AssetSnapshotRecord{
		Timestamp:    snapshot.Timestamp,
		CurrentState: int(snapshot.CurrentState),
		TotalValue:   bigIntToString(snapshot.TotalValue),
		AmountWavax:  bigIntToString(snapshot.AmountWavax),
		AmountUsdc:   bigIntToString(snapshot.AmountUsdc),
		AmountBlack:  bigIntToString(snapshot.AmountBlack),
		AmountAvax:   bigIntToString(snapshot.AmountAvax),
	}

	result := r.db.Create(&record)
	if result.Error != nil {
		return fmt.Errorf("failed to record snapshot: %w", result.Error)
	}

	return nil
}

// GetDB returns the underlying GORM DB instance for advanced queries
func (r *MySQLRecorder) GetDB() *gorm.DB {
	return r.db
}

// Close closes the database connection
func (r *MySQLRecorder) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying DB: %w", err)
	}
	return sqlDB.Close()
}

// bigIntToString safely converts *big.Int to string, handling nil values
func bigIntToString(value *big.Int) string {
	if value == nil {
		return "0"
	}
	return value.String()
}

// GetLatestSnapshot retrieves the most recent snapshot from the database
func (r *MySQLRecorder) GetLatestSnapshot() (*AssetSnapshotRecord, error) {
	var record AssetSnapshotRecord
	result := r.db.Order("timestamp DESC").First(&record)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get latest snapshot: %w", result.Error)
	}
	return &record, nil
}

// GetSnapshotsByTimeRange retrieves snapshots within a time range
func (r *MySQLRecorder) GetSnapshotsByTimeRange(start, end time.Time) ([]AssetSnapshotRecord, error) {
	var records []AssetSnapshotRecord
	result := r.db.Where("timestamp BETWEEN ? AND ?", start, end).
		Order("timestamp ASC").
		Find(&records)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get snapshots by time range: %w", result.Error)
	}
	return records, nil
}

// GetSnapshotsByPhase retrieves all snapshots for a specific strategy phase
func (r *MySQLRecorder) GetSnapshotsByPhase(phase blackholedex.StrategyPhase) ([]AssetSnapshotRecord, error) {
	var records []AssetSnapshotRecord
	result := r.db.Where("current_state = ?", int(phase)).
		Order("timestamp ASC").
		Find(&records)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get snapshots by phase: %w", result.Error)
	}
	return records, nil
}

// CountSnapshots returns the total number of snapshots in the database
func (r *MySQLRecorder) CountSnapshots() (int64, error) {
	var count int64
	result := r.db.Model(&AssetSnapshotRecord{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count snapshots: %w", result.Error)
	}
	return count, nil
}
