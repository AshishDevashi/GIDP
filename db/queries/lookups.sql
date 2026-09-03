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

-- name: ListRegistryProviders :many
SELECT * FROM registry_providers
ORDER BY id;
