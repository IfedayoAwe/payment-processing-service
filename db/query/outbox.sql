-- name: CreateOutboxEntry :one
INSERT INTO outbox (id, job_type, payload)
VALUES (gen_random_uuid()::text, $1, $2)
RETURNING id, job_type, payload, processed, processed_at, retry_count, created_at;

-- name: GetUnprocessedOutboxEntries :many
SELECT id, job_type, payload, processed, processed_at, retry_count, created_at
FROM outbox
WHERE processed = FALSE
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: MarkOutboxEntryProcessed :exec
UPDATE outbox
SET processed = TRUE, processed_at = NOW()
WHERE id = $1;

-- name: IncrementOutboxRetryCount :exec
UPDATE outbox
SET retry_count = retry_count + 1
WHERE id = $1;
