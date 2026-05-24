const baseUrl = process.env.RECSYS_BASE_URL || "http://localhost:8000";
const tenantId = process.env.RECSYS_TENANT_ID || "demo";
const apiKey = process.env.RECSYS_API_KEY || "";

async function recommend(userId, surface = "home") {
  const requestId = crypto.randomUUID();
  const headers = {
    "Content-Type": "application/json",
    "X-Org-Id": tenantId,
    "X-Request-Id": requestId
  };
  if (apiKey) {
    headers["X-API-Key"] = apiKey;
  }

  const response = await fetch(`${baseUrl}/v1/recommend`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      surface,
      k: 10,
      user: { anonymous_id: userId },
      options: { include_reasons: true }
    })
  });

  if (!response.ok) {
    return fallback(requestId, surface, `status_${response.status}`);
  }

  const payload = await response.json();
  if (!payload.items || payload.items.length === 0) {
    return fallback(requestId, surface, "empty");
  }

  logExposure({
    request_id: payload.meta?.request_id || requestId,
    tenant_id: payload.meta?.tenant_id || tenantId,
    surface,
    user_id: userId,
    items: payload.items.map((item) => ({ item_id: item.item_id, rank: item.rank }))
  });
  return payload.items;
}

function fallback(requestId, surface, reason) {
  console.warn(JSON.stringify({ event: "recsys_fallback", request_id: requestId, surface, reason }));
  return [
    { item_id: "fallback-1", rank: 1 },
    { item_id: "fallback-2", rank: 2 }
  ];
}

function logExposure(event) {
  console.log(JSON.stringify({ event: "recsys_exposure", ...event, ts: new Date().toISOString() }));
}

const userId = process.argv[2] || "anon-demo";
recommend(userId).then((items) => {
  console.log(JSON.stringify({ items }));
});
