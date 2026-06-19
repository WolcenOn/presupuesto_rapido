// Cliente API inicial para conectar AntenaManager PRO con el backend Go.
// De momento no se importa desde index.html para no romper la app actual.

const AMP_API_BASE_URL = window.AMP_API_BASE_URL || "http://localhost:8080";
const AMP_PENDING_DOCS_KEY = "amp_pending_documents";
const AMP_SYNCED_DOCS_KEY = "amp_synced_documents";
let ampAccessToken = "";

async function ampApiRequest(path, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {})
  };
  if (ampAccessToken) headers.Authorization = `Bearer ${ampAccessToken}`;

  let res = await fetch(`${AMP_API_BASE_URL}${path}`, {
    credentials: "include",
    ...options,
    headers
  });

  if (res.status === 401 && path !== "/api/auth/login" && path !== "/api/auth/refresh") {
    const refreshed = await ampRefreshSession().catch(() => null);
    if (refreshed?.accessToken) {
      headers.Authorization = `Bearer ${ampAccessToken}`;
      res = await fetch(`${AMP_API_BASE_URL}${path}`, {
        credentials: "include",
        ...options,
        headers
      });
    }
  }

  const text = await res.text();
  const data = text ? JSON.parse(text) : null;
  if (!res.ok) {
    const message = data?.error || `Error HTTP ${res.status}`;
    throw new Error(message);
  }
  return data;
}

export async function ampLogin(email, password) {
  const data = await ampApiRequest("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
  ampSetSession(data);
  return data;
}

export async function ampRefreshSession() {
  const data = await ampApiRequest("/api/auth/refresh", { method: "POST" });
  ampSetSession(data);
  return data;
}

export async function ampLogout() {
  try {
    await ampApiRequest("/api/auth/logout", { method: "POST" });
  } finally {
    ampAccessToken = "";
  }
}

export async function ampMe() {
  return ampApiRequest("/api/me");
}

export async function ampListPrices() {
  return ampApiRequest("/api/prices");
}

export async function ampGetPrice(id) {
  return ampApiRequest(`/api/prices/${encodeURIComponent(id)}`);
}

export async function ampCreatePrice(item) {
  return ampApiRequest("/api/prices", {
    method: "POST",
    body: JSON.stringify(item)
  });
}

export async function ampUpdatePrice(id, item) {
  return ampApiRequest(`/api/prices/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: JSON.stringify(item)
  });
}

export async function ampDisablePrice(id) {
  return ampApiRequest(`/api/prices/${encodeURIComponent(id)}`, {
    method: "DELETE"
  });
}

export async function ampSyncDocument(doc) {
  return ampApiRequest("/api/documents", {
    method: "POST",
    body: JSON.stringify(ampDocumentPayload(doc))
  });
}

export async function ampListDocuments() {
  return ampApiRequest("/api/documents");
}

export async function ampGetDocument(id) {
  return ampApiRequest(`/api/documents/${encodeURIComponent(id)}`);
}

export async function ampQueueDocumentForBoss(id) {
  return ampApiRequest(`/api/documents/${encodeURIComponent(id)}/send-to-boss`, {
    method: "POST"
  });
}

export async function ampListUsers() {
  return ampApiRequest("/api/admin/users");
}

export async function ampCreateUser(user) {
  return ampApiRequest("/api/admin/users", {
    method: "POST",
    body: JSON.stringify(user)
  });
}

export async function ampRestoreSession() {
  return ampRefreshSession();
}

export function ampQueueLocalDocument(doc) {
  const pending = ampReadLocalArray(AMP_PENDING_DOCS_KEY);
  const now = new Date().toISOString();
  const queued = { ...doc, pendingSync: true, queuedAt: doc.queuedAt || now, lastSyncError: "" };
  const index = pending.findIndex((item) => item.ref === queued.ref);
  if (index >= 0) pending[index] = queued;
  else pending.unshift(queued);
  ampWriteLocalArray(AMP_PENDING_DOCS_KEY, pending);
  return queued;
}

export function ampListPendingLocalDocuments() {
  return ampReadLocalArray(AMP_PENDING_DOCS_KEY);
}

export async function ampSyncPendingLocalDocuments() {
  const pending = ampReadLocalArray(AMP_PENDING_DOCS_KEY);
  const remaining = [];
  const synced = ampReadLocalArray(AMP_SYNCED_DOCS_KEY);
  const results = [];

  for (const doc of pending) {
    try {
      const result = await ampSyncDocument(doc);
      const syncedDoc = { ...doc, pendingSync: false, syncedAt: new Date().toISOString(), backendId: result.id || doc.backendId };
      synced.unshift(syncedDoc);
      results.push({ ref: doc.ref, ok: true, result });
    } catch (error) {
      remaining.push({ ...doc, lastSyncError: error.message || String(error) });
      results.push({ ref: doc.ref, ok: false, error: error.message || String(error) });
    }
  }

  ampWriteLocalArray(AMP_PENDING_DOCS_KEY, remaining);
  ampWriteLocalArray(AMP_SYNCED_DOCS_KEY, ampDedupeByRef(synced));
  return results;
}

export function ampFindSyncedDocumentsOlderThan(days = 60) {
  const cutoff = Date.now() - days * 24 * 60 * 60 * 1000;
  return ampReadLocalArray(AMP_SYNCED_DOCS_KEY).filter((doc) => {
    const syncedAt = Date.parse(doc.syncedAt || doc.createdAt || 0);
    return Number.isFinite(syncedAt) && syncedAt < cutoff;
  });
}

export function ampDeleteSyncedDocumentsOlderThan(days = 60) {
  const oldRefs = new Set(ampFindSyncedDocumentsOlderThan(days).map((doc) => doc.ref));
  const kept = ampReadLocalArray(AMP_SYNCED_DOCS_KEY).filter((doc) => !oldRefs.has(doc.ref));
  ampWriteLocalArray(AMP_SYNCED_DOCS_KEY, kept);
  return oldRefs.size;
}

function ampDocumentPayload(doc) {
  return {
    ref: doc.ref,
    type: doc.type,
    clientName: doc.clientName,
    clientCif: doc.clientCif,
    clientPhone: doc.clientPhone,
    clientAddress: doc.clientAddress,
    workOrder: doc.workOrder,
    paymentMethod: doc.paymentMethod,
    base: doc.base,
    iva: doc.iva,
    total: doc.total,
    documentJson: doc
  };
}

function ampReadLocalArray(key) {
  try {
    const value = JSON.parse(window.localStorage.getItem(key) || "[]");
    return Array.isArray(value) ? value : [];
  } catch {
    return [];
  }
}

function ampWriteLocalArray(key, value) {
  window.localStorage.setItem(key, JSON.stringify(value));
}

function ampDedupeByRef(items) {
  const seen = new Set();
  return items.filter((item) => {
    if (!item.ref || seen.has(item.ref)) return false;
    seen.add(item.ref);
    return true;
  });
}

function ampSetSession(data) {
  if (data?.accessToken) {
    ampAccessToken = data.accessToken;
  }
}
