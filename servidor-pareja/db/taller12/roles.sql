CREATE ROLE :"migration_user" LOGIN PASSWORD :'migration_password';
CREATE ROLE :"api_user" LOGIN PASSWORD :'api_password';

ALTER DATABASE :"database_name" OWNER TO :"migration_user";
GRANT CONNECT ON DATABASE :"database_name" TO :"api_user";

-- Solo el migrador crea objetos en public; la API recibe únicamente USAGE.
ALTER SCHEMA public OWNER TO :"migration_user";
