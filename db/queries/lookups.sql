-- name: ListLifecycles :many
SELECT * FROM lifecycles
ORDER BY id;

-- name: ListTiers :many
SELECT * FROM tiers
ORDER BY id;

-- name: ListServiceTypes :many
SELECT * FROM service_types
ORDER BY id;

-- name: ListRepoProviders :many
SELECT * FROM repo_providers
ORDER BY id;

-- name: ListLanguages :many
SELECT * FROM languages
ORDER BY id;

-- name: ListRepoTemplates :many
SELECT * FROM repo_templates
WHERE is_active = true
ORDER BY id;
