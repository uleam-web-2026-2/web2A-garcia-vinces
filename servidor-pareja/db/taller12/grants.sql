GRANT USAGE ON SCHEMA public TO :"api_user";

-- Permisos de filas limitados a las operaciones que usa cada tabla.
GRANT SELECT, INSERT ON TABLE public.usuarios TO :"api_user";
GRANT SELECT, INSERT, UPDATE ON TABLE public.cuentas TO :"api_user";
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.eventos TO :"api_user";
GRANT SELECT, INSERT, UPDATE ON TABLE public.invitados TO :"api_user";
GRANT SELECT, INSERT ON TABLE public.fotos TO :"api_user";

-- GORM recupera los IDs generados al insertar filas.
GRANT USAGE, SELECT ON SEQUENCE
    public.usuarios_id_seq,
    public.cuentas_id_seq,
    public.eventos_id_seq,
    public.invitados_id_seq,
    public.fotos_id_seq
TO :"api_user";
