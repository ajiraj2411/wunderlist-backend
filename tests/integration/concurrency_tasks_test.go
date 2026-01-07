package integration

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestConcurrentTaskCreation(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	listID := createListAndGetID(t, access)

	const workers = 100
	var wg sync.WaitGroup
	wg.Add(workers)

	success := int32(0)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()

			payload := fmt.Sprintf(
				`{"title":"task-%d","list_id":"%s"}`,
				i,
				listID,
			)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/tasks",
				bytes.NewBufferString(payload),
			)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+access)

			w := httptest.NewRecorder()
			TestRouter.ServeHTTP(w, req)

			if w.Code == http.StatusCreated {
				atomic.AddInt32(&success, 1)
			}
		}(i)
	}

	wg.Wait()

	if success != workers {
		t.Fatalf("expected %d successful creates, got %d", workers, success)
	}
}
