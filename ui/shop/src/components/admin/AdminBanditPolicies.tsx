"use client";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useToast } from "@/components/ToastProvider";
import { BanditPolicyManager } from "@/components/admin/BanditPolicyManager";

type BanditPolicy = {
  policy_id: string;
  name?: string;
  active?: boolean;
  blend_alpha?: number;
  blend_beta?: number;
  blend_gamma?: number;
  mmr_lambda?: number;
  brand_cap?: number;
  category_cap?: number;
  updated_at?: string;
};

export function AdminBanditPolicies() {
  const toast = useToast();
  const [policies, setPolicies] = useState<BanditPolicy[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const selectedPolicy = useMemo(
    () =>
      policies.find((policy) => policy.policy_id === selectedId) ?? undefined,
    [policies, selectedId]
  );

  const loadPolicies = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch("/api/admin/bandit-policies", {
        cache: "no-store",
      });
      if (!response.ok) {
        throw new Error(`Failed to load policies (${response.status})`);
      }
      const data = await response.json();
      setPolicies(Array.isArray(data?.policies) ? data.policies : []);
      if (data?.policies?.length && !selectedId) {
        setSelectedId(data.policies[0].policy_id);
      }
    } catch (err) {
      console.error("Failed to load bandit policies", err);
      setError(err instanceof Error ? err.message : "Failed to load policies");
    } finally {
      setLoading(false);
    }
  }, [selectedId]);

  useEffect(() => {
    loadPolicies();
  }, [loadPolicies]);

  const upsertPolicy = useCallback(
    async (policy: BanditPolicy) => {
      const response = await fetch("/api/admin/bandit-policies", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ policy }),
      });
      if (!response.ok) {
        const result = await response.json().catch(() => ({}));
        const message =
          (result && result.error) ||
          `Failed to save policy (status ${response.status})`;
        throw new Error(message);
      }
      await loadPolicies();
      setSelectedId(policy.policy_id);
    },
    [loadPolicies]
  );

  const deletePolicy = useCallback(
    async (policyId: string) => {
      const response = await fetch(`/api/admin/bandit-policies/${policyId}`, {
        method: "DELETE",
      });
      if (!response.ok) {
        const result = await response.json().catch(() => ({}));
        const message =
          (result && result.error) ||
          `Failed to deactivate policy (status ${response.status})`;
        throw new Error(message);
      }
      toast(`Deactivated policy ${policyId}`);
      await loadPolicies();
      setSelectedId(null);
    },
    [loadPolicies, toast]
  );

  const nukePolicies = useCallback(async () => {
    const response = await fetch(`/api/admin/bandit-policies/all`, {
      method: "DELETE",
    });
    if (!response.ok) {
      const result = await response.json().catch(() => ({}));
      const message =
        (result && result.error) ||
        `Failed to deactivate all policies (status ${response.status})`;
      throw new Error(message);
    }
    toast("Deactivated all bandit policies");
    await loadPolicies();
    setSelectedId(null);
  }, [loadPolicies, toast]);

  return (
    <section className="space-y-4">
      <header className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">Bandit Policies</h2>
          <p className="text-sm text-gray-600">
            Manage exploration policies stored in the RecSys service.
            Deactivating a policy marks it inactive, so the bandit will no
            longer pick it. You can reactivate policies later.
          </p>
        </div>
        <div className="flex gap-2">
          <button
            className="px-3 py-1 text-sm border rounded"
            onClick={() => loadPolicies()}
            disabled={loading}
          >
            Refresh
          </button>
          <button
            className="px-3 py-1 text-sm border rounded"
            onClick={() => {
              if (confirm("Deactivate all policies?")) {
                nukePolicies().catch((err) => toast((err as Error).message));
              }
            }}
            disabled={loading || policies.length === 0}
          >
            Deactivate All
          </button>
        </div>
      </header>

      {error ? <p className="text-sm text-red-600">{error}</p> : null}

      <div className="grid gap-4 md:grid-cols-3">
        <div className="md:col-span-1 space-y-2">
          <h3 className="text-sm font-medium text-gray-700">
            Existing policies
          </h3>
          <div className="border rounded divide-y">
            {policies.length === 0 ? (
              <p className="text-sm text-gray-500 p-3">No policies found.</p>
            ) : (
              policies.map((policy) => {
                const isSelected = policy.policy_id === selectedId;
                return (
                  <div
                    key={policy.policy_id}
                    className={`flex items-center justify-between px-3 py-2 text-sm cursor-pointer ${
                      isSelected ? "bg-blue-50" : "hover:bg-gray-50"
                    }`}
                    onClick={() => setSelectedId(policy.policy_id)}
                  >
                    <div>
                      <p className="font-medium">
                        {policy.name || policy.policy_id}
                      </p>
                      <p className="text-xs text-gray-500">
                        {policy.policy_id} ·{" "}
                        {policy.active ? "active" : "inactive"}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      {!policy.active && (
                        <button
                          className="text-xs text-blue-600 hover:underline"
                          onClick={(event) => {
                            event.stopPropagation();
                            upsertPolicy({ ...policy, active: true })
                              .then(() =>
                                toast(`Reactivated policy ${policy.policy_id}`)
                              )
                              .catch((err) => toast((err as Error).message));
                          }}
                        >
                          Reactivate
                        </button>
                      )}
                      <button
                        className="text-xs text-red-600 hover:underline"
                        onClick={(event) => {
                          event.stopPropagation();
                          if (
                            confirm(`Deactivate policy '${policy.policy_id}'?`)
                          ) {
                            deletePolicy(policy.policy_id).catch((err) =>
                              toast((err as Error).message)
                            );
                          }
                        }}
                      >
                        {policy.active ? "Deactivate" : "Deactivated"}
                      </button>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>

        <div className="md:col-span-2">
          <BanditPolicyManager
            onUpsert={async (policy) => {
              try {
                await upsertPolicy(policy);
                toast("Policy saved");
              } catch (error) {
                toast((error as Error).message);
                throw error;
              }
            }}
            initialPolicy={selectedPolicy}
          />
        </div>
      </div>
    </section>
  );
}
