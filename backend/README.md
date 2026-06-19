# Backend Go para AntenaManager PRO

Backend inicial para sincronizar presupuestos, albaranes, facturas y precios estándar.

## Estado actual

Incluye una primera base para:

- Despliegue en Railway con Docker.
- Conexión a PostgreSQL usando `DATABASE_URL`.
- Migración inicial para usuarios, precios, documentos, ajustes SMTP, auditoría y logs de correo.
- Endpoints iniciales de documentos y precios.
- Middleware de seguridad HTTP, CORS y timeout.
- Hashing de contraseñas con Argon2id.

La autenticación JWT real está pendiente. Mientras se desarrolla, `cmd/api/main.go` usa cabeceras `X-Dev-*` para simular usuarios. No usar ese modo en producción.

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

La migración inicial está en:

```txt
backend/migrations/001_init.sql
```

Se puede aplicar desde Railway, psql o una herramienta de migraciones como `goose`, `tern` o `migrate`.

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

1. Añadir login real con access token corto y refresh token en cookie HttpOnly.
2. Crear comando de bootstrap para generar el primer usuario jefe.
3. Añadir endpoints `PATCH /api/prices/:id`, `GET /api/documents/:id` y envío de documentos al jefe.
4. Generar PDF o aceptar PDF subido desde el frontend.
5. Crear cola de correo y reintentos.
6. Integrar `frontend/api-client.js` en `index.html` de forma progresiva.
