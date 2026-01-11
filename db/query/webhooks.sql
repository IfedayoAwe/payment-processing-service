-- name: CreateWebhookEvent :one
INSERT INTO webhook_events (id, provider_name, event_type, provider_reference, transaction_id, payload)
VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
RETURNING *;

