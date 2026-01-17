package db

import (
	blackholedex "blackholego"
	"math/big"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestMySQLRecorder_RecordReport(t *testing.T) {
	// Create a mock database
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer sqlDB.Close()

	// Create GORM DB with mock
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create gorm DB: %v", err)
	}

	// Set up expectations
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `asset_snapshots`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Create recorder without auto-migration for testing
	recorder := &MySQLRecorder{db: gormDB}

	// Create test snapshot
	snapshot := blackholedex.CurrentAssetSnapshot{
		Timestamp:    time.Now(),
		CurrentState: blackholedex.ActiveMonitoring,
		TotalValue:   big.NewInt(1000000),
		AmountWavax:  big.NewInt(500000),
		AmountUsdc:   big.NewInt(300000),
		AmountBlack:  big.NewInt(150000),
		AmountAvax:   big.NewInt(50000),
	}

	// Test RecordReport
	err = recorder.RecordReport(snapshot)
	if err != nil {
		t.Errorf("RecordReport failed: %v", err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestBigIntToString(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected string
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: "0",
		},
		{
			name:     "zero value",
			input:    big.NewInt(0),
			expected: "0",
		},
		{
			name:     "positive value",
			input:    big.NewInt(123456789),
			expected: "123456789",
		},
		{
			name:     "large value",
			input:    new(big.Int).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}),
			expected: "18446744073709551615",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bigIntToString(tt.input)
			if result != tt.expected {
				t.Errorf("bigIntToString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAssetSnapshotRecord_TableName(t *testing.T) {
	record := AssetSnapshotRecord{}
	expected := "asset_snapshots"
	if record.TableName() != expected {
		t.Errorf("TableName() = %v, want %v", record.TableName(), expected)
	}
}

// Integration test example (requires actual MySQL instance)
// Uncomment and configure DSN to run
/*
func TestMySQLRecorder_Integration(t *testing.T) {
	// Configure your test database DSN
	dsn := "testuser:testpass@tcp(localhost:3306)/blackhole_test?charset=utf8mb4&parseTime=True&loc=Local"

	recorder, err := NewMySQLRecorder(dsn)
	if err != nil {
		t.Fatalf("failed to create recorder: %v", err)
	}
	defer recorder.Close()

	// Create test snapshot
	snapshot := blackholedex.CurrentAssetSnapshot{
		Timestamp:    time.Now(),
		CurrentState: blackholedex.Initializing,
		TotalValue:   big.NewInt(1000000),
		AmountWavax:  big.NewInt(500000),
		AmountUsdc:   big.NewInt(300000),
		AmountBlack:  big.NewInt(150000),
		AmountAvax:   big.NewInt(50000),
	}

	// Test RecordReport
	err = recorder.RecordReport(snapshot)
	if err != nil {
		t.Errorf("RecordReport failed: %v", err)
	}

	// Test GetLatestSnapshot
	latest, err := recorder.GetLatestSnapshot()
	if err != nil {
		t.Errorf("GetLatestSnapshot failed: %v", err)
	}
	if latest == nil {
		t.Error("expected latest snapshot to be non-nil")
	}

	// Test CountSnapshots
	count, err := recorder.CountSnapshots()
	if err != nil {
		t.Errorf("CountSnapshots failed: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one snapshot")
	}
}
*/
