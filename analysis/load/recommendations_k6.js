import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';
const ORG_ID = __ENV.ORG_ID || '00000000-0000-0000-0000-000000000001';
const NAMESPACE = __ENV.NAMESPACE || 'default';
const SURFACE = __ENV.SURFACE || 'home';
const USER_POOL = (__ENV.USER_POOL || 'load_user_0001,load_user_0002,load_user_0003,load_user_0004,load_user_0005').split(',');
const K = Number(__ENV.K || 20);
const INCLUDE_REASONS = (__ENV.INCLUDE_REASONS || 'true').toLowerCase() === 'true';
const SUMMARY_PATH = __ENV.SUMMARY_PATH || 'analysis/results/load_test_summary.json';
const STAGE_DURATION = __ENV.STAGE_DURATION || '30s';
const RPS_TARGETS = (__ENV.RPS_TARGETS || '10,100,1000').split(',').map((v) => Number(v.trim()));

const stages = RPS_TARGETS.map((target) => ({ target, duration: STAGE_DURATION }));

export const options = {
  scenarios: {
    ramped_rps: {
      executor: 'ramping-arrival-rate',
      startRate: 1,
      timeUnit: '1s',
      preAllocatedVUs: Number(__ENV.PREALLOCATED_VUS || 200),
      maxVUs: Number(__ENV.MAX_VUS || 2000),
      stages,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.02'],
  },
};

let userIndex = 0;
function nextUser() {
  const id = USER_POOL[userIndex % USER_POOL.length].trim();
  userIndex += 1;
  return id || `load_user_${userIndex}`;
}

export default function () {
  const payload = {
    namespace: NAMESPACE,
    user_id: nextUser(),
    k: K,
    include_reasons: INCLUDE_REASONS,
    context: { surface: SURFACE },
  };
  const res = http.post(`${BASE_URL}/v1/recommendations`, JSON.stringify(payload), {
    headers: {
      'Content-Type': 'application/json',
      'X-Org-ID': ORG_ID,
    },
    timeout: __ENV.HTTP_TIMEOUT || '30s',
  });
  check(res, {
    'status is 200': (r) => r.status === 200,
    'has items': (r) => r.json('items') && r.json('items').length > 0,
  });
}

export function handleSummary(data) {
  const summary = {
    created_at: new Date().toISOString(),
    base_url: BASE_URL,
    namespace: NAMESPACE,
    surface: SURFACE,
    stages: stages.map((stage) => ({ target_rps: stage.target, duration: stage.duration })),
    metrics: {
      http_req_duration_ms: {
        p50: data.metrics.http_req_duration.values['p(50)'],
        p95: data.metrics.http_req_duration.values['p(95)'],
        p99: data.metrics.http_req_duration.values['p(99)'],
      },
      http_req_failed_rate: data.metrics.http_req_failed.values.rate,
      iterations: data.metrics.iterations.values.count,
    },
  };
  const json = JSON.stringify(summary, null, 2);
  return {
    stdout: json + '\n',
    [SUMMARY_PATH]: json,
  };
}
