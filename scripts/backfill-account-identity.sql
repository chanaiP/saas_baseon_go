-- Backfill account / user_identity for legacy app_user rows that have not been linked yet.
-- Review the result in staging before running in production.

WITH legacy_users AS (
  SELECT
    u.id AS user_id,
    u.account,
    u.phone,
    u.email,
    u.password_hash,
    u.name,
    u.avatar_url
  FROM app_user u
  WHERE u.deleted_at IS NULL
    AND u.account_id IS NULL
    AND COALESCE(NULLIF(u.phone, ''), NULLIF(u.account, '')) IS NOT NULL
),
inserted_accounts AS (
  INSERT INTO account (login_account, phone, email, password_hash, status, created_at, updated_at)
  SELECT
    COALESCE(NULLIF(phone, ''), account),
    NULLIF(phone, ''),
    NULLIF(email, ''),
    password_hash,
    1,
    now(),
    now()
  FROM legacy_users
  ON CONFLICT DO NOTHING
  RETURNING id, login_account, phone, email
),
matched_accounts AS (
  SELECT DISTINCT ON (u.user_id)
    u.user_id,
    a.id AS account_id,
    u.name,
    u.avatar_url,
    u.phone,
    u.email
  FROM legacy_users u
  JOIN account a
    ON a.deleted_at IS NULL
   AND (
      a.login_account = COALESCE(NULLIF(u.phone, ''), u.account)
      OR (u.phone IS NOT NULL AND u.phone <> '' AND a.phone = u.phone)
      OR (u.email IS NOT NULL AND u.email <> '' AND a.email = u.email)
   )
  ORDER BY u.user_id, a.id
),
inserted_identities AS (
  INSERT INTO user_identity (account_id, display_name, avatar_url, phone, email, status, created_at, updated_at)
  SELECT account_id, COALESCE(NULLIF(name, ''), '未命名用户'), avatar_url, NULLIF(phone, ''), NULLIF(email, ''), 1, now(), now()
  FROM matched_accounts
  WHERE NOT EXISTS (
    SELECT 1 FROM user_identity ui
    WHERE ui.account_id = matched_accounts.account_id
      AND ui.deleted_at IS NULL
  )
  RETURNING id, account_id
),
matched_identities AS (
  SELECT DISTINCT ON (m.user_id)
    m.user_id,
    m.account_id,
    ui.id AS identity_user_id
  FROM matched_accounts m
  JOIN user_identity ui ON ui.account_id = m.account_id AND ui.deleted_at IS NULL
  ORDER BY m.user_id, ui.id
)
UPDATE app_user u
SET
  account_id = m.account_id,
  identity_user_id = m.identity_user_id,
  updated_at = now()
FROM matched_identities m
WHERE u.id = m.user_id
  AND u.account_id IS NULL;
