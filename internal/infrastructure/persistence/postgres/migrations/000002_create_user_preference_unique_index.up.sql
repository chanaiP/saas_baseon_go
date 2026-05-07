CREATE UNIQUE INDEX IF NOT EXISTS idx_user_preference_key
ON public.user_preference (user_id, pref_key);
