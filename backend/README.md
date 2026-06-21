# Backend Go para AntenaManager PRO

Backend inicial para sincronizar presupuestos, albaranes, facturas y precios estándar.

## Estado actual

Incluye una primera base para:

- Despliegue en Railway con Docker.
- Conexión a PostgreSQL usando `DATABASE_URL`.
- Migraciones para empresas, usuarios, refresh tokens, precios, documentos, ajustes SMTP, auditoría y logs de correo.
- Endpoints de autenticación, empresa, documentos, precios y administración de usuarios.
- Middleware de seguridad HTTP, CORS, timeout y autenticación JWT.
- Hashing de credenciales con Argon2id.
- Refresh token en cookie HttpOnly.
- Setup inicial de empresa + usuario `owner`.
- Cola inicial de correo al jefe para albaranes y facturas usando `document_email_logs`.
- Worker SMTP básico para procesar correos `queued`.

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

## Roles y permisos

```txt
owner    propietario de la empresa
boss     jefe/gestor de la empresa
employee empleado de la empresa
```

Reglas actuales:

- `owner` puede configurar datos fiscales y generales de la empresa.
- `owner` y `boss` pueden crear, consultar y modificar usuarios operativos y precios.
- `employee` puede consultar datos de empresa, consultar precios y crear/sincronizar sus documentos.
- No hay endpoint público de borrado de precios ni documentos; se evita borrar datos de negocio por seguridad y trazabilidad.

## Crear la primera empresa y su owner

Configura `BOOTSTRAP_SECRET` en Railway o en `.env`. Mientras no exista ningún usuario con rol `owner`, puedes crear la primera empresa y su propietario con:

```bash
curl -X POST http://localhost:8080/api/setup/boss \
  -H 'Content-Type: application/json' \
  -H 'X-Setup-Token: <BOOTSTRAP_SECRET>' \
  -d '{"name":"Propietario","email":"owner@example.com","secret":"cambia-esta-clave","companyName":"Empresa Demo","companyTaxId":"B00000000","companyEmail":"admin@example.com","companyPhone":"600000000","companyAddress":"Calle Demo 1"}'
```

Después de crear el `owner`, el endpoint responderá conflicto si se intenta usar otra vez porque ya existe un propietario inicial. Aun así, conviene borrar `BOOTSTRAP_SECRET` de Railway tras completar el alta inicial.

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

## Empresa

```txt
GET   /api/company          owner, boss, employee
PATCH /api/company          solo owner
```

Los datos de empresa se usarán para cabecera/datos fiscales de futuras facturas y documentos.

## Precios estándar

```txt
GET   /api/prices           owner, boss, employee
POST  /api/prices           owner, boss
GET   /api/prices/{id}      owner, boss, employee
PATCH /api/prices/{id}      owner, boss
```

No se expone borrado de precios. Cuando sea necesario, se podrá añadir desactivación controlada y auditada.

## Documentos

```txt
GET  /api/documents
POST /api/documents
GET  /api/documents/{id}
POST /api/documents/{id}/send-to-boss
```

Reglas:

- `owner` y `boss` pueden ver los documentos de la empresa.
- Un `employee` solo puede ver sus propios documentos.
- Al crear un `albaran` o una `factura`, se crea automáticamente un registro `queued` en `document_email_logs` si `BOSS_EMAIL` está configurado.
- `send-to-boss` permite reencolar manualmente un albarán/factura.

## Worker de correo

El worker se activa con:

```txt
MAIL_WORKER_ENABLED=true
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=<usuario smtp>
SMTP_PASSWORD=<clave smtp>
SMTP_FROM_EMAIL=notificaciones@example.com
SMTP_FROM_NAME=AntenaManager PRO
```

Cada 30 segundos busca hasta 10 correos `queued`, envía un correo de texto al jefe y actualiza el log a `sent` o `failed`. Cuando el envío se completa, también marca `documents.sent_to_boss_at`.

De momento el correo incluye datos básicos del documento. El adjunto PDF queda para la siguiente fase.

## Administración de usuarios

```txt
GET  /api/admin/users        owner, boss
POST /api/admin/users        owner, boss
```

El campo de creación de credencial de usuario se transforma en hash Argon2id en el servidor antes de guardarse. La intención de permisos es que `owner` pueda crear `boss` y `employee`, mientras que `boss` solo cree `employee`.

## Cliente frontend auxiliar

`frontend/api-client.js` todavía no está conectado al `index.html`, pero ya incluye helpers para:

- Setup: creación del primer owner y empresa.
- Sesión: login, refresh, logout y restauración.
- Precios: listar, crear, consultar y actualizar.
- Documentos: sincronizar, listar, consultar y reencolar envío al jefe.
- Usuarios: listar y crear.
- Offline: guardar documentos pendientes, sincronizar pendientes y limpiar documentos sincronizados antiguos.

## Variables de entorno Railway

```txt
APP_ENV=production
PORT=8080
DATABASE_URL=<inyectada por Railway PostgreSQL>
JWT_SECRET=<secreto largo aleatorio>
BOOTSTRAP_SECRET=<token temporal para crear el primer owner>
CORS_ALLOWED_ORIGINS=https://tu-dominio.com
BOSS_EMAIL=<correo-del-jefe>
LOCAL_RETENTION_DAYS=60
ACCESS_TOKEN_MINUTES=15
REFRESH_TOKEN_DAYS=30
MAIL_WORKER_ENABLED=true
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM_EMAIL=
SMTP_FROM_NAME=AntenaManager PRO
```

## Limitaciones encontradas

El worker SMTP actual envía un correo de texto, pero todavía no genera ni adjunta PDF. Para producción también conviene añadir control de reintentos con contador y evitar reintentos infinitos.

## Siguientes pasos técnicos

1. Terminar filtros estrictos por `company_id` en login, usuarios, precios y documentos.
2. Implementar invitaciones por correo para empleados.
3. Generar PDF o aceptar PDF subido desde el frontend.
4. Integrar `frontend/api-client.js` en `index.html` de forma progresiva.
5. Añadir tests de permisos por rol y propiedad de documento.
