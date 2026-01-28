export interface BanditMeta {
  policyId?: string;
  requestId?: string;
  algorithm?: string;
  bucket?: string;
  explore?: boolean;
  experiment?: string;
  variant?: string;
}

export function hasBanditMeta(meta?: BanditMeta | null): meta is BanditMeta {
  return Boolean(meta && (meta.policyId || meta.requestId));
}
