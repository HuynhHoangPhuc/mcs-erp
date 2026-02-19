CREATE TABLE IF NOT EXISTS public.users_lookup (
    email VARCHAR(255) PRIMARY KEY,
    tenant_schema VARCHAR(63) NOT NULL REFERENCES public.tenants(schema_name)
);

CREATE INDEX IF NOT EXISTS idx_users_lookup_email ON public.users_lookup(email);
