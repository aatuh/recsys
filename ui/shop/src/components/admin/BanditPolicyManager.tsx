"use client";
import { useEffect, useState } from "react";
import { useToast } from "@/components/ToastProvider";

type BanditPolicy = {
  policy_id: string;
  name: string;
  active: boolean;
  blend_alpha: number;
  blend_beta: number;
  blend_gamma: number;
  mmr_lambda: number;
  brand_cap: number;
  category_cap: number;
};

const DEFAULT_POLICY: BanditPolicy = {
  policy_id: "manual_explore_default",
  name: "Manual explore (default)",
  active: true,
  blend_alpha: 0.35,
  blend_beta: 0.4,
  blend_gamma: 0.25,
  mmr_lambda: 0.6,
  brand_cap: 2,
  category_cap: 3,
};

type Props = {
  onUpsert: (policy: BanditPolicy) => Promise<void>;
  initialPolicy?: BanditPolicy;
};

export function BanditPolicyManager({ onUpsert, initialPolicy }: Props) {
  const toast = useToast();
  const [policy, setPolicy] = useState<BanditPolicy>(
    initialPolicy ?? DEFAULT_POLICY
  );
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setPolicy(initialPolicy ?? DEFAULT_POLICY);
  }, [initialPolicy]);

  const updatePolicy = <K extends keyof BanditPolicy>(
    key: K,
    value: BanditPolicy[K]
  ) => {
    setPolicy((prev) => ({ ...prev, [key]: value }));
  };

  const handleNumberChange =
    (key: keyof BanditPolicy, fallback: number) =>
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const raw = event.target.value.trim();
      updatePolicy(key, raw === "" ? fallback : Number(raw));
    };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!policy.policy_id.trim()) {
      toast("Policy ID is required");
      return;
    }
    setLoading(true);
    try {
      await onUpsert({
        ...DEFAULT_POLICY,
        ...policy,
        policy_id: (policy.policy_id || DEFAULT_POLICY.policy_id).trim(),
        name:
          policy.name?.trim() ||
          (policy.policy_id || DEFAULT_POLICY.policy_id).trim(),
        blend_alpha: policy.blend_alpha ?? DEFAULT_POLICY.blend_alpha,
        blend_beta: policy.blend_beta ?? DEFAULT_POLICY.blend_beta,
        blend_gamma: policy.blend_gamma ?? DEFAULT_POLICY.blend_gamma,
        mmr_lambda: policy.mmr_lambda ?? DEFAULT_POLICY.mmr_lambda,
        brand_cap: policy.brand_cap ?? DEFAULT_POLICY.brand_cap,
        category_cap: policy.category_cap ?? DEFAULT_POLICY.category_cap,
      });
      toast("Bandit policy saved");
      setPolicy(DEFAULT_POLICY);
    } catch (error) {
      console.error("Failed to upsert bandit policy:", error);
      toast("Failed to save bandit policy");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-4 border rounded p-4">
      <div>
        <h3 className="text-lg font-medium">Bandit Policies</h3>
        <p className="text-sm text-gray-600">
          Create or update exploration policies. Policy IDs must match
          <code className="mx-1">SHOP_BANDIT_POLICY_IDS</code> for the shop to
          use them.
        </p>
      </div>

      <form
        onSubmit={handleSubmit}
        className="grid gap-3 md:grid-cols-2"
        autoComplete="off"
      >
        <label className="text-sm">
          <span className="text-gray-700">Policy ID</span>
          <input
            className="mt-1 w-full rounded border px-2 py-1 text-sm"
            value={policy.policy_id}
            onChange={(event) => updatePolicy("policy_id", event.target.value)}
            placeholder="manual_explore_default"
          />
        </label>

        <label className="text-sm">
          <span className="text-gray-700">Display name</span>
          <input
            className="mt-1 w-full rounded border px-2 py-1 text-sm"
            value={policy.name}
            onChange={(event) => updatePolicy("name", event.target.value)}
            placeholder="Manual explore (default)"
          />
        </label>

        <label className="text-sm">
          <span className="text-gray-700">Blend α (popularity)</span>
          <input
            type="number"
            step="0.05"
            className="mt-1 w-full rounded border px-2 py-1 text-sm"
            value={policy.blend_alpha}
            onChange={handleNumberChange(
              "blend_alpha",
              DEFAULT_POLICY.blend_alpha
            )}
          />
        </label>

        <label className="text-sm">
          <span className="text-gray-700">Blend β (co-visitation)</span>
          <input
            type="number"
            step="0.05"
            className="mt-1 w-full rounded border px-2 py-1 text-sm"
            value={policy.blend_beta}
            onChange={handleNumberChange(
              "blend_beta",
              DEFAULT_POLICY.blend_beta
            )}
          />
        </label>

        <label className="text-sm">
          <span className="text-gray-700">Blend γ (ALS)</span>
          <input
            type="number"
            step="0.05"
            className="mt-1 w-full rounded border px-2 py-1 text-sm"
            value={policy.blend_gamma}
            onChange={handleNumberChange(
              "blend_gamma",
              DEFAULT_POLICY.blend_gamma
            )}
          />
        </label>

        <label className="text-sm">
          <span className="text-gray-700">MMR λ</span>
          <input
            type="number"
            step="0.05"
            className="mt-1 w-full rounded border px-2 py-1 text-sm"
            value={policy.mmr_lambda}
            onChange={handleNumberChange(
              "mmr_lambda",
              DEFAULT_POLICY.mmr_lambda
            )}
          />
        </label>

        <label className="text-sm">
          <span className="text-gray-700">Brand cap</span>
          <input
            type="number"
            className="mt-1 w-full rounded border px-2 py-1 text-sm"
            value={policy.brand_cap}
            onChange={handleNumberChange("brand_cap", DEFAULT_POLICY.brand_cap)}
          />
        </label>

        <label className="text-sm">
          <span className="text-gray-700">Category cap</span>
          <input
            type="number"
            className="mt-1 w-full rounded border px-2 py-1 text-sm"
            value={policy.category_cap}
            onChange={handleNumberChange(
              "category_cap",
              DEFAULT_POLICY.category_cap
            )}
          />
        </label>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={policy.active}
            onChange={(event) => updatePolicy("active", event.target.checked)}
          />
          Active
        </label>

        <div className="md:col-span-2 flex gap-2">
          <button
            type="submit"
            className="px-4 py-2 bg-blue-600 text-white rounded disabled:opacity-60"
            disabled={loading}
          >
            {loading ? "Saving…" : "Save policy"}
          </button>
        </div>
      </form>
    </div>
  );
}
