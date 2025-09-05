-- createTenantFunc creates a helper function `app_current_tenant()` to fetch current tenant_id from session.
CREATE FUNCTION app_current_tenant() RETURNS BIGINT AS $$
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
CREATE POLICY ut_select ON user_tenants 
	FOR SELECT 
	USING (tenant_id = app_current_tenant());
CREATE POLICY ut_insert ON user_tenants
	FOR INSERT
	WITH CHECK (tenant_id = app_current_tenant());
CREATE POLICY ut_update ON user_tenants
	FOR UPDATE
	USING (tenant_id = app_current_tenant())
	WITH CHECK (tenant_id = app_current_tenant());
CREATE POLICY ut_delete ON user_tenants
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
			WHERE ut.user_id = NEW.id
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
			WHERE ut.user_id = NEW.id
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