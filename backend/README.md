# Backend Go para AntenaManager PRO

Backend inicial para sincronizar presupuestos, albaranes, facturas y precios estándar.

## Estado actual

Incluye una primera base para:

- Despliegue en Railway con Docker.
- Conexión a PostgreSQL usando `DATABASE_URL`.
- Migraciones para usuarios, refresh tokens, precios, documentos, ajustes SMTP, auditoría y logs de correo.
- Endpoints iniciales de autenticación, documentos y precios.
- Middleware de seguridad HTTP, CORS, timeout y autenticación JWT.
- Hashing de contraseñas con Argon2id.
- Refresh token en cookie HttpOnly.

## Ejecutar en local

```bash
cd backend
cp .env.example .env
# Ajusta DATABASE_URL y variables necesarias

go mod download
go run ./cmd/api
```

Comprobar salud:

```bash
curl http://localhost:8080/health
```

## Migraciones

Las migraciones están en:

```txt
backend/migrations/001_init.sql
backend/migrations/002_refresh_tokens.sql
```

Se pueden aplicar desde Railway, `psql` o una herramienta de migraciones como `goose`, `tern` o `migrate`.

## Crear el primer usuario jefe

Hasta que exista un comando de bootstrap automatizado, se puede insertar un usuario jefe generando el hash con una pequeña utilidad temporal que llame a `auth.HashPassword`, o hacerlo desde una tarea interna de administración. No guardes contraseñas en texto plano en SQL.

Tabla objetivo:

```sql
insert into users (name, email, password_hash, role, is_active)
values ('Jefe', 'jefe@example.com', '<argon2id_hash>', 'boss', true);
```

## Autenticación

Login:

```bash
curl -i -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"jefe@example.com","password":"cambia-esto"}'
```

La respuesta devuelve `accessToken` y guarda el refresh token como cookie HttpOnly en `/api/auth`.

Uso de endpoint protegido:

```bash
curl http://localhost:8080/api/documents \
  -H "Authorization: Bearer <accessToken>"
```

Refresh:

```bash
curl -i -X POST http://localhost:8080/api/auth/refresh \
  --cookie "amp_refresh_token=<cookie>"
```

## Variables de entorno Railway

```txt
APP_ENV=production
PORT=8080
DATABASE_URL=<inyectada por Railway PostgreSQL>
JWT_SECRET=<secreto largo aleatorio>
CORS_ALLOWED_ORIGINS=https://tu-dominio.com
BOSS_EMAIL=<correo-del-jefe>
LOCAL_RETENTION_DAYS=60
ACCESS_TOKEN_MINUTES=15
REFRESH_TOKEN_DAYS=30
```

## Siguientes pasos técnicos

1. Añadir comando de bootstrap para generar el primer usuario jefe de forma cómoda.
2. Añadir endpoints `PATCH /api/prices/:id`, `GET /api/documents/:id` y envío de documentos al jefe.
3. Generar PDF o aceptar PDF subido desde el frontend.
4. Crear cola de correo y reintentos.
5. Integrar `frontend/api-client.js` en `index.html` de forma progresiva.
6. Añadir tests de permisos por rol y propiedad de documento.
