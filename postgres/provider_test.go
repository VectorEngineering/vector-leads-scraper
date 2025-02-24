package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Vector/vector-leads-scraper/gmaps"
	"github.com/gosom/scrapemate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProvider(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Test with default batch size
	p := NewProvider(db).(*provider)
	assert.Equal(t, batchSize, p.batchSize)
	assert.Equal(t, 2*batchSize, cap(p.jobc))

	// Test with custom batch size
	customSize := 20
	p = NewProvider(db, WithBatchSize(customSize)).(*provider)
	assert.Equal(t, customSize, p.batchSize)
	assert.Equal(t, 2*customSize, cap(p.jobc))

	// Test with invalid batch size (should default to batchSize)
	p = NewProvider(db, WithBatchSize(-1)).(*provider)
	assert.Equal(t, batchSize, p.batchSize)
}

func TestPush(t *testing.T) {
	tests := []struct {
		name     string
		mockSetup func(mock sqlmock.Sqlmock)
		job      scrapemate.IJob
		wantErr  bool
	}{
		{
			name: "Push GmapJob",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO gmaps_jobs").
					WithArgs(
						"test-id", 
						int64(1), // Changed from float64 to int64
						"search", 
						sqlmock.AnyArg(), 
						sqlmock.AnyArg(), 
						statusNew,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			job: &gmaps.GmapJob{
				Job: scrapemate.Job{
					ID:       "test-id",
					Priority: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "Push PlaceJob",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO gmaps_jobs").
					WithArgs(
						"place-id", 
						int64(2), // Changed from float64 to int64
						"place", 
						sqlmock.AnyArg(), 
						sqlmock.AnyArg(), 
						statusNew,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			job: &gmaps.PlaceJob{
				Job: scrapemate.Job{
					ID:       "place-id",
					Priority: 2,
				},
			},
			wantErr: false,
		},
		{
			name: "Push EmailExtractJob",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO gmaps_jobs").
					WithArgs(
						"email-id", 
						int64(3), // Changed from float64 to int64
						"email", 
						sqlmock.AnyArg(), 
						sqlmock.AnyArg(), 
						statusNew,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			job: &gmaps.EmailExtractJob{
				Job: scrapemate.Job{
					ID:       "email-id",
					Priority: 3,
				},
			},
			wantErr: false,
		},
		{
			name: "Push with DB error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO gmaps_jobs").
					WithArgs(
						"error-id", 
						int64(1), // Changed from float64 to int64
						"search", 
						sqlmock.AnyArg(), 
						sqlmock.AnyArg(), 
						statusNew,
					).
					WillReturnError(sql.ErrConnDone)
			},
			job: &gmaps.GmapJob{
				Job: scrapemate.Job{
					ID:       "error-id",
					Priority: 1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			p := NewProvider(db)
			ctx := context.Background()

			err = p.Push(ctx, tt.job)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestJobs(t *testing.T) {
	t.Run("FetchJobs returns jobs", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Set up the mock query with precise text matching and expected args
		mock.ExpectQuery(`WITH updated AS \(.*\)`).
			WithArgs(statusQueued, statusNew, batchSize).
			WillReturnRows(sqlmock.NewRows([]string{"payload_type", "payload"}).
				AddRow("search", encodeGmapJob(t, "job1")).
				AddRow("place", encodePlaceJob(t, "job2")))

		p := NewProvider(db, WithBatchSize(batchSize)).(*provider)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		jobChan, errChan := p.Jobs(ctx)

		// Read from channels
		var jobs []scrapemate.IJob
		for i := 0; i < 2; i++ {
			select {
			case job := <-jobChan:
				jobs = append(jobs, job)
			case err := <-errChan:
				t.Fatalf("Unexpected error: %v", err)
			case <-time.After(1 * time.Second):
				t.Fatal("Timeout waiting for jobs")
			}
		}

		// Verify we received both jobs
		assert.Len(t, jobs, 2)
		assert.Equal(t, "job1", jobs[0].GetID())
		assert.Equal(t, "job2", jobs[1].GetID())

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("FetchJobs handles query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(`WITH updated AS \(.*\)`).
			WithArgs(statusQueued, statusNew, batchSize).
			WillReturnError(sql.ErrConnDone)

		p := NewProvider(db, WithBatchSize(batchSize)).(*provider)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, errChan := p.Jobs(ctx)

		// Should receive an error
		select {
		case err := <-errChan:
			assert.Error(t, err)
			assert.Equal(t, sql.ErrConnDone, err)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for error")
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Jobs fetches once with multiple calls", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// The query should only be executed once even with multiple calls to Jobs()
		mock.ExpectQuery(`WITH updated AS \(.*\)`).
			WithArgs(statusQueued, statusNew, batchSize).
			WillReturnRows(sqlmock.NewRows([]string{"payload_type", "payload"}).
				AddRow("search", encodeGmapJob(t, "job1")))

		p := NewProvider(db, WithBatchSize(batchSize))
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Call Jobs twice
		p.Jobs(ctx)
		p.Jobs(ctx)

		// Sleep briefly to let goroutines execute
		time.Sleep(100 * time.Millisecond)

		// Verify that only one query was executed
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDecodeJob(t *testing.T) {
	t.Run("Decode GmapJob", func(t *testing.T) {
		payload := encodeGmapJob(t, "test-gmap")
		job, err := decodeJob("search", payload)
		require.NoError(t, err)
		
		gmapJob, ok := job.(*gmaps.GmapJob)
		assert.True(t, ok)
		assert.Equal(t, "test-gmap", gmapJob.GetID())
	})

	t.Run("Decode PlaceJob", func(t *testing.T) {
		payload := encodePlaceJob(t, "test-place")
		job, err := decodeJob("place", payload)
		require.NoError(t, err)
		
		placeJob, ok := job.(*gmaps.PlaceJob)
		assert.True(t, ok)
		assert.Equal(t, "test-place", placeJob.GetID())
	})

	t.Run("Decode EmailExtractJob", func(t *testing.T) {
		payload := encodeEmailJob(t, "test-email")
		job, err := decodeJob("email", payload)
		require.NoError(t, err)
		
		emailJob, ok := job.(*gmaps.EmailExtractJob)
		assert.True(t, ok)
		assert.Equal(t, "test-email", emailJob.GetID())
	})

	t.Run("Invalid payload type", func(t *testing.T) {
		payload := encodeGmapJob(t, "invalid")
		job, err := decodeJob("invalid", payload)
		assert.Error(t, err)
		assert.Nil(t, job)
	})

	t.Run("Corrupt payload", func(t *testing.T) {
		job, err := decodeJob("search", []byte("corrupt"))
		assert.Error(t, err)
		assert.Nil(t, job)
	})
}

// Helper functions to encode jobs
func encodeGmapJob(t *testing.T, id string) []byte {
	job := &gmaps.GmapJob{
		Job: scrapemate.Job{
			ID: id,
		},
	}
	return encodeJob(t, job)
}

func encodePlaceJob(t *testing.T, id string) []byte {
	job := &gmaps.PlaceJob{
		Job: scrapemate.Job{
			ID: id,
		},
	}
	return encodeJob(t, job)
}

func encodeEmailJob(t *testing.T, id string) []byte {
	job := &gmaps.EmailExtractJob{
		Job: scrapemate.Job{
			ID: id,
		},
	}
	return encodeJob(t, job)
}

// Encode a job directly using the same method as provider.Push
func encodeJob(t *testing.T, job scrapemate.IJob) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	var err error
	switch j := job.(type) {
	case *gmaps.GmapJob:
		err = enc.Encode(j)
	case *gmaps.PlaceJob:
		err = enc.Encode(j)
	case *gmaps.EmailExtractJob:
		err = enc.Encode(j)
	default:
		t.Fatalf("invalid job type %T", job)
	}
	
	require.NoError(t, err, "Failed to encode job")
	return buf.Bytes()
}

// Mock invalid job type for testing
type invalidJob struct{}

func (j *invalidJob) GetID() string          { return "invalid" }
func (j *invalidJob) GetURL() string         { return "" }
func (j *invalidJob) GetPriority() float64   { return 0 }
func (j *invalidJob) SetPriority(p float64)  {}
func (j *invalidJob) Unmarshall([]byte) error { return nil }