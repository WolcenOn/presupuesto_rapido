# Plan de backend, sincronización y seguridad

## Objetivo

Convertir la app actual, que funciona como herramienta rápida local, en una solución sincronizada con backend Go, PostgreSQL y despliegue en Railway sin perder la capacidad offline del empleado.

## Roles

### Jefe

- Ver todos los presupuestos, albaranes y facturas.
- Modificar precios estándar.
- Crear, activar o desactivar empleados.
- Ver logs de envío y auditoría.
- Recibir por correo albaranes y facturas.

### Empleado

- Crear presupuestos, albaranes y facturas.
- Ver solamente sus propios documentos.
- Mantener copia local en su dispositivo.
- Sincronizar documentos cuando haya conexión.
- Recibir avisos para borrar documentos locales antiguos ya sincronizados.

## Sincronización

1. El frontend guarda el documento localmente con `pendingSync=true`.
2. Cuando haya conexión, envía `POST /api/documents`.
3. El backend usa `(employee_id, ref)` como clave de idempotencia.
4. Si el documento ya existe, se actualiza sin duplicarlo.
5. Si es albarán o factura, se encola generación/envío de copia al jefe.
6. El frontend marca el documento como sincronizado.

## Retención local

- No borrar documentos no sincronizados.
- Avisar al empleado cuando existan documentos sincronizados con antigüedad superior a `LOCAL_RETENTION_DAYS`.
- Permitir borrado manual asistido de documentos antiguos sincronizados.

## Seguridad mínima

- HTTPS obligatorio en producción.
- CORS limitado al dominio real.
- Contraseñas con Argon2id.
- Access token corto y refresh token en cookie HttpOnly/Secure/SameSite.
- RBAC en backend, no solo en frontend.
- Restricción por propiedad: empleados solo acceden a `documents.employee_id = user.id`.
- Auditoría de cambios en precios, usuarios y documentos sensibles.
- Cifrado de credenciales SMTP de empleados si se almacenan.
- No registrar datos completos de clientes en logs.

## Despliegue Railway

Servicios previstos:

- `backend`: Dockerfile en `backend/Dockerfile`.
- `postgres`: plugin PostgreSQL de Railway.
- Variables de entorno configuradas desde Railway.

## Pendiente antes de producción

- Sustituir autenticación temporal `X-Dev-*` por JWT/cookies.
- Añadir migrador automático o paso documentado de migraciones.
- Añadir tests de permisos.
- Añadir política de backups.
- Añadir proveedor de email o SMTP cifrado por empleado.
