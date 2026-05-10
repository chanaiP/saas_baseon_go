CREATE TABLE IF NOT EXISTS file_object (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  file_id VARCHAR(64) NOT NULL,
  created_by BIGINT NOT NULL,
  original_name VARCHAR(255) NOT NULL,
  stored_name VARCHAR(255) NOT NULL,
  storage_path TEXT NOT NULL,
  mime_type VARCHAR(128) NOT NULL,
  file_size BIGINT NOT NULL,
  status BIGINT NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ,
  CONSTRAINT file_object_file_size_nonnegative CHECK (file_size >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_file_object_tenant_file_id ON file_object (tenant_id, file_id);
CREATE INDEX IF NOT EXISTS idx_file_object_created_by ON file_object (created_by);
CREATE INDEX IF NOT EXISTS idx_file_object_status ON file_object (status);
CREATE INDEX IF NOT EXISTS idx_file_object_deleted_at ON file_object (deleted_at);
