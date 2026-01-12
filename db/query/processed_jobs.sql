-- name: MarkJobProcessed :one
INSERT INTO processed_jobs (job_id, expires_at)
VALUES ($1, $2)
ON CONFLICT (job_id) DO NOTHING
RETURNING job_id, processed_at, expires_at;

-- name: IsJobProcessed :one
SELECT EXISTS(SELECT 1 FROM processed_jobs WHERE job_id = $1) as processed;

-- name: CleanupExpiredJobs :exec
DELETE FROM processed_jobs WHERE expires_at < NOW();
