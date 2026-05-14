ALTER TABLE ai_provider_accounts
  DROP COLUMN IF EXISTS maintainer_contact,
  DROP COLUMN IF EXISTS maintainer,
  DROP COLUMN IF EXISTS login_account,
  DROP COLUMN IF EXISTS login_method;

