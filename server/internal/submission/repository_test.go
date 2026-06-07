package submission

import (
	"strings"
	"testing"
)

func TestMarkRejectedOnlyUpdatesPendingSubmissions(t *testing.T) {
	query := markRejectedQuery()

	if !strings.Contains(query, "AND status = $5") {
		t.Fatalf("expected reject update to guard pending status, got:\n%s", query)
	}
}
