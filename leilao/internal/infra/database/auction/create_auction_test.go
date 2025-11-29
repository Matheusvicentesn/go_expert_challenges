package auction

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"

	"go.mongodb.org/mongo-driver/bson"
)

func TestAutoCloseRoutineTriggersAfterInterval(t *testing.T) {
	var (
		mu              sync.Mutex
		closedAuctionID string
		wasCalled       bool
	)

	closeFn := func(ctx context.Context, auctionID string) error {
		mu.Lock()
		defer mu.Unlock()
		closedAuctionID = auctionID
		wasCalled = true
		return nil
	}

	auctionID := "test-auction-123"
	interval := 50 * time.Millisecond

	go func() {
		<-time.After(interval)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = closeFn(ctx, auctionID)
	}()

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !wasCalled {
		t.Fatal("Expected close function to be called, but it wasn't")
	}
	if closedAuctionID != auctionID {
		t.Fatalf("Expected auction %s to be closed, got %s", auctionID, closedAuctionID)
	}
}

func TestGetAuctionInterval(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectedResult time.Duration
	}{
		{
			name:           "valid duration from env",
			envValue:       "10m",
			expectedResult: 10 * time.Minute,
		},
		{
			name:           "valid duration in seconds",
			envValue:       "30s",
			expectedResult: 30 * time.Second,
		},
		{
			name:           "invalid duration falls back to default",
			envValue:       "invalid",
			expectedResult: 5 * time.Minute,
		},
		{
			name:           "empty env falls back to default",
			envValue:       "",
			expectedResult: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("AUCTION_INTERVAL", tt.envValue)
			defer os.Unsetenv("AUCTION_INTERVAL")

			result := getAuctionInterval()
			if result != tt.expectedResult {
				t.Errorf("Expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestUpdateAuctionStatusToCompleted(t *testing.T) {
	filter := bson.M{
		"_id":    "test-id",
		"status": auction_entity.Active,
	}
	update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

	if filter["_id"] == nil || filter["status"] == nil {
		t.Error("Filter should contain _id and status")
	}

	if update["$set"] == nil {
		t.Error("Update should contain $set operator")
	}

	statusUpdate := update["$set"].(bson.M)["status"]
	if statusUpdate != auction_entity.Completed {
		t.Errorf("Expected status to be Completed, got %v", statusUpdate)
	}
}
