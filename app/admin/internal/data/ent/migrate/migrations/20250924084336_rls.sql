-- createTenantFunc creates a helper function `app_current_tenant()` to fetch current tenant_id from session.
CREATE OR REPLACE FUNCTION app_current_tenant() RETURNS BIGINT AS $$
BEGIN
	RETURN current_setting('app.current_tenant')::BIGINT;
EXCEPTION WHEN others THEN
	RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- Enable Row Level Security (RLS) for all tables in the public schema.
-- Enable RLS for tenants table.
-- SELECT：普通租户用户只能看到自己的租户
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenants_select_policy ON tenants
    FOR SELECT
    USING (id = app_current_tenant());
-- INSERT：普通租户用户不能插入租户
-- DELETE：普通租户用户不能删除租户
-- UPDATE：普通租户用户只能修改自己的租户
CREATE POLICY tenants_update_policy ON tenants
    FOR UPDATE
    USING (id = app_current_tenant())
    WITH CHECK (id = app_current_tenant());


-- Enable RLS for user_tenants table.
ALTER TABLE user_tenants ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_tenants_select ON user_tenants 
	FOR SELECT 
	USING (tenant_id = app_current_tenant());
CREATE POLICY user_tenants_insert ON user_tenants
	FOR INSERT
	WITH CHECK (tenant_id = app_current_tenant());
CREATE POLICY user_tenants_update ON user_tenants
	FOR UPDATE
	USING (tenant_id = app_current_tenant())
	WITH CHECK (tenant_id = app_current_tenant());
CREATE POLICY user_tenants_delete ON user_tenants
	FOR DELETE
	USING (tenant_id = app_current_tenant());

-- Enable RLS for departments table.
ALTER TABLE departments ENABLE ROW LEVEL SECURITY;
CREATE POLICY dept_select ON departments
	FOR SELECT
	USING (tenant_id = app_current_tenant());
CREATE POLICY dept_insert ON departments
	FOR INSERT
	WITH CHECK (tenant_id = app_current_tenant());
CREATE POLICY dept_update ON departments
	FOR UPDATE
	USING (tenant_id = app_current_tenant())
	WITH CHECK (tenant_id = app_current_tenant());
CREATE POLICY dept_delete ON departments
	FOR DELETE
	USING (tenant_id = app_current_tenant());

-- Enable RLS for users table.
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY users_select ON users
    FOR SELECT
    USING (
		EXISTS (
			SELECT 1 FROM user_tenants ut
			WHERE ut.user_id = users.id
				AND ut.tenant_id = app_current_tenant()
		)
	);
CREATE POLICY users_insert ON users
	FOR INSERT
	WITH CHECK (
		EXISTS (
			SELECT 1 FROM user_tenants ut
			WHERE ut.user_id = users.id
				AND ut.tenant_id = app_current_tenant()
		)
	);
CREATE POLICY users_update ON users
	FOR UPDATE
	USING (
		EXISTS (
			SELECT 1 FROM user_tenants ut
			WHERE ut.user_id = users.id
			  AND ut.tenant_id = app_current_tenant()
		)
	)
	WITH CHECK (
		EXISTS (
			SELECT 1 FROM user_tenants ut
			WHERE ut.user_id = users.id
			  AND ut.tenant_id = app_current_tenant()
		)
	);
CREATE POLICY users_delete ON users
	FOR DELETE
	USING (
		EXISTS (
			SELECT 1 FROM user_tenants ut
			WHERE ut.user_id = users.id
			  AND ut.tenant_id = app_current_tenant()
		)
	);

-- function: ensure user_departments.tenant_id matches departments.tenant_id
CREATE OR REPLACE FUNCTION enforce_user_departments_tenant()
RETURNS TRIGGER AS $$
BEGIN
  -- insert or update, enforce tenant_id to match departments
  SELECT d.tenant_id INTO NEW.tenant_id
  FROM departments d
  WHERE d.id = NEW.department_id;

  IF NEW.tenant_id IS NULL THEN
    RAISE EXCEPTION 'Invalid department_id %, no tenant found', NEW.department_id;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- trigger: before insert or update on user_departments
DROP TRIGGER IF EXISTS trg_user_departments_tenant ON user_departments;
CREATE TRIGGER trg_user_departments_tenant
BEFORE INSERT OR UPDATE ON user_departments
FOR EACH ROW
EXECUTE FUNCTION enforce_user_departments_tenant();

-- Enable RLS for user_departments table.
ALTER TABLE user_departments ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_departments_select ON user_departments
    FOR SELECT
    USING (tenant_id = app_current_tenant());
CREATE POLICY user_departments_insert ON user_departments
    FOR INSERT
    WITH CHECK (tenant_id = app_current_tenant());
CREATE POLICY user_departments_update ON user_departments
    FOR UPDATE
    USING (tenant_id = app_current_tenant())
    WITH CHECK (tenant_id = app_current_tenant());
CREATE POLICY user_departments_delete ON user_departments
    FOR DELETE
    USING (tenant_id = app_current_tenant());