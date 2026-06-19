// Cliente API inicial para conectar AntenaManager PRO con el backend Go.
// De momento no se importa desde index.html para no romper la app actual.

const AMP_API_BASE_URL = window.AMP_API_BASE_URL || "http://localhost:8080";
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
    body: JSON.stringify({
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
    })
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

function ampSetSession(data) {
  if (data?.accessToken) {
    ampAccessToken = data.accessToken;
  }
}
