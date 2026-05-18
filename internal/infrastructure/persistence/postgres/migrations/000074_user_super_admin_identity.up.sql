ALTER TABLE public.app_user
  ADD COLUMN IF NOT EXISTS is_tenant_admin boolean NOT NULL DEFAULT false;

UPDATE public.app_user u
SET is_tenant_admin = false
FROM public.tenant t
WHERE u.tenant_id = t.id
  AND t.is_platform_tenant = true;

UPDATE public.app_user u
SET is_tenant_admin = true
FROM public.tenant t
WHERE u.tenant_id = t.id
  AND t.is_platform_tenant = false
  AND u.deleted_at IS NULL
  AND u.id = (
    SELECT MIN(first_user.id)
    FROM public.app_user first_user
    WHERE first_user.tenant_id = u.tenant_id
      AND first_user.deleted_at IS NULL
  );

UPDATE public.app_user u
SET is_tenant_admin = true
FROM public.user_role ur
JOIN public.role r ON r.id = ur.role_id
JOIN public.tenant t ON t.id = r.tenant_id
WHERE ur.user_id = u.id
  AND r.code = 'admin'
  AND r.deleted_at IS NULL
  AND t.is_platform_tenant = false
  AND u.deleted_at IS NULL;
