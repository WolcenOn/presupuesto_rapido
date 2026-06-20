# Backend Go para AntenaManager PRO

Backend inicial para sincronizar presupuestos, albaranes, facturas y precios estándar.

## Estado actual

Incluye una primera base para:

- Despliegue en Railway con Docker.
- Conexión a PostgreSQL usando `DATABASE_URL`.
- Migraciones para usuarios, refresh tokens, precios, documentos, ajustes SMTP, auditoría y logs de correo.
- Endpoints de autenticación, documentos, precios y administración de usuarios.
- Middleware de seguridad HTTP, CORS, timeout y autenticación JWT.
- Hashing de credenciales con Argon2id.
- Refresh token en cookie HttpOnly.
- Setup inicial de un solo uso para crear el primer jefe.
- Cola inicial de correo al jefe para albaranes y facturas usando `document_email_logs`.
- Variables de configuración preparadas para un worker de correo futuro.

## Ejecutar en local

```bash
cd backend
cp .env.example .env
go mod download
go run ./cmd/api
```

Comprobar salud:

```bash
curl http://localhost:8080/health
```

## Migraciones

```txt
backend/migrations/001_init.sql
backend/migrations/002_refresh_tokens.sql
```

Se pueden aplicar desde Railway, `psql` o una herramienta de migraciones como `goose`, `tern` o `migrate`.

## Crear el primer usuario jefe

Configura `BOOTSTRAP_SECRET` en Railway o en `.env`. Mientras no exista ningún usuario con rol `boss`, puedes crear el primer jefe con:

```bash
curl -X POST http://localhost:8080/api/setup/boss \
  -H 'Content-Type: application/json' \
  -H 'X-Setup-Token: <BOOTSTRAP_SECRET>' \
  -d '{"name":"Jefe","email":"jefe@example.com","secret":"cambia-esta-clave"}'
```

Después de crear el primer jefe, el endpoint responderá conflicto si se intenta usar otra vez porque ya existe un jefe. Aun así, conviene borrar `BOOTSTRAP_SECRET` de Railway tras completar el alta inicial.

Una vez que el jefe pueda iniciar sesión, podrá crear empleados o más jefes desde la API.

## Autenticación

```txt
POST /api/auth/login
POST /api/auth/refresh
POST /api/auth/logout
GET  /api/me
```

La respuesta de login devuelve `accessToken` y guarda el refresh token como cookie HttpOnly en `/api/auth`.

Los endpoints protegidos usan:

```txt
Authorization: Bearer <accessToken>
```

## Precios estándar

```txt
GET    /api/prices
POST   /api/prices              solo jefe
GET    /api/prices/{id}
PATCH  /api/prices/{id}         solo jefe
DELETE /api/prices/{id}         solo jefe, desactiva el precio
```

## Documentos

```txt
GET    /api/documents
POST   /api/documents
GET    /api/documents/{id}
POST   /api/documents/{id}/send-to-boss
```

Reglas:

- El jefe puede ver todos los documentos.
- Un empleado solo puede ver sus propios documentos.
- Al crear un `albaran` o una `factura`, se crea automáticamente un registro `queued` en `document_email_logs` si `BOSS_EMAIL` está configurado.
- `send-to-boss` permite reencolar manualmente un albarán/factura.

## Administración de usuarios

```txt
GET  /api/admin/users        solo jefe
POST /api/admin/users        solo jefe
```

El campo de creación de credencial de usuario se transforma en hash Argon2id en el servidor antes de guardarse.

## Cliente frontend auxiliar

`frontend/api-client.js` todavía no está conectado al `index.html`, pero ya incluye helpers para:

- Setup: creación del primer jefe.
- Sesión: login, refresh, logout y restauración.
- Precios: listar, crear, consultar, actualizar y desactivar.
- Documentos: sincronizar, listar, consultar y reencolar envío al jefe.
- Usuarios: listar y crear.
- Offline: guardar documentos pendientes, sincronizar pendientes y limpiar documentos sincronizados antiguos.

## Variables de entorno Railway

```txt
APP_ENV=production
PORT=8080
DATABASE_URL=<inyectada por Railway PostgreSQL>
JWT_SECRET=<secreto largo aleatorio>
BOOTSTRAP_SECRET=<token temporal para crear el primer jefe>
CORS_ALLOWED_ORIGINS=https://tu-dominio.com
BOSS_EMAIL=<correo-del-jefe>
LOCAL_RETENTION_DAYS=60
ACCESS_TOKEN_MINUTES=15
REFRESH_TOKEN_DAYS=30
MAIL_WORKER_ENABLED=false
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM_EMAIL=
SMTP_FROM_NAME=AntenaManager PRO
```

## Limitaciones encontradas

El worker que envía correos reales queda pendiente. Esta rama ya deja preparada la cola `document_email_logs`, la configuración SMTP y el reencolado manual, pero todavía no procesa los registros `queued`.

## Siguientes pasos técnicos

1. Implementar worker real de correo para procesar `document_email_logs`.
2. Generar PDF o aceptar PDF subido desde el frontend.
3. Integrar `frontend/api-client.js` en `index.html` de forma progresiva.
4. Añadir tests de permisos por rol y propiedad de documento.
