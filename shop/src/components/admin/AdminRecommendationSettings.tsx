"use client";
import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type ChangeEvent,
  type FormEvent,
} from "react";
import { useToast } from "@/components/ToastProvider";

const KNOWN_EVENT_TYPES = ["view", "click", "add", "purchase"] as const;

type AlgorithmProfile = {
  profileId: string;
  name: string;
  description?: string | null;
  surface?: string | null;
  isDefault: boolean;
  blendAlpha: number;
  blendBeta: number;
  blendGamma: number;
  popularityHalflifeDays: number;
  covisWindowDays: number;
  popularityFanout: number;
  mmrLambda: number;
  brandCap: number;
  categoryCap: number;
  ruleExcludeEvents: boolean;
  purchasedWindowDays: number;
  profileWindowDays: number;
  profileTopN: number;
  profileBoost: number;
  excludeEventTypes: string[];
  createdAt: string;
  updatedAt: string;
};

type BanditStatus = {
  configured: boolean;
  enabled: boolean;
  reason?: string;
  checkedAt: string;
  policyIds: string[];
  missingPolicies?: string[];
};

type SettingsResponse = {
  profile: AlgorithmProfile;
  profile_source?: string;
  profiles: AlgorithmProfile[];
  bandit?: BanditStatus;
  configuredPolicies?: string[];
};

type RecommendationSettingsForm = {
  profileId: string;
  name: string;
  description: string;
  surface: string;
  isDefault: boolean;
  blendAlpha: string;
  blendBeta: string;
  blendGamma: string;
  popularityHalflifeDays: string;
  covisWindowDays: string;
  popularityFanout: string;
  mmrLambda: string;
  brandCap: string;
  categoryCap: string;
  ruleExcludeEvents: boolean;
  purchasedWindowDays: string;
  profileWindowDays: string;
  profileTopN: string;
  profileBoost: string;
  excludeEventTypes: string[];
  updatedAt?: string;
};

const DEFAULT_FORM_VALUES: RecommendationSettingsForm = {
  profileId: "default",
  name: "Default profile",
  description: "Seeded defaults",
  surface: "",
  isDefault: false,
  blendAlpha: "0.25",
  blendBeta: "0.35",
  blendGamma: "0.4",
  popularityHalflifeDays: "4",
  covisWindowDays: "28",
  popularityFanout: "500",
  mmrLambda: "0.3",
  brandCap: "2",
  categoryCap: "3",
  ruleExcludeEvents: true,
  purchasedWindowDays: "180",
  profileWindowDays: "30",
  profileTopN: "64",
  profileBoost: "0.7",
  excludeEventTypes: ["view", "click", "add"],
  updatedAt: undefined,
};

function toFormState(profile: AlgorithmProfile): RecommendationSettingsForm {
  return {
    profileId: profile.profileId,
    name: profile.name,
    description: profile.description ?? "",
    surface: profile.surface ?? "",
    isDefault: profile.isDefault,
    blendAlpha: profile.blendAlpha.toString(),
    blendBeta: profile.blendBeta.toString(),
    blendGamma: profile.blendGamma.toString(),
    popularityHalflifeDays: profile.popularityHalflifeDays.toString(),
    covisWindowDays: profile.covisWindowDays.toString(),
    popularityFanout: profile.popularityFanout.toString(),
    mmrLambda: profile.mmrLambda.toString(),
    brandCap: profile.brandCap.toString(),
    categoryCap: profile.categoryCap.toString(),
    ruleExcludeEvents: profile.ruleExcludeEvents,
    purchasedWindowDays: profile.purchasedWindowDays.toString(),
    profileWindowDays: profile.profileWindowDays.toString(),
    profileTopN: profile.profileTopN.toString(),
    profileBoost: profile.profileBoost.toString(),
    excludeEventTypes: [...profile.excludeEventTypes],
    updatedAt: profile.updatedAt,
  };
}

function parseFloatField(label: string, value: string): number {
  const parsed = Number.parseFloat(value);
  if (!Number.isFinite(parsed)) {
    throw new Error(`${label} must be a number`);
  }
  return parsed;
}

function parseIntField(label: string, value: string): number {
  const parsed = Number.parseInt(value, 10);
  if (!Number.isFinite(parsed)) {
    throw new Error(`${label} must be an integer`);
  }
  return parsed;
}

export default function AdminRecommendationSettings() {
  const toast = useToast();
  const [profiles, setProfiles] = useState<AlgorithmProfile[]>([]);
  const [profileSource, setProfileSource] = useState<string>("default");
  const [selectedProfileId, setSelectedProfileId] = useState<string | null>(
    null
  );
  const [isCreating, setIsCreating] = useState(false);
  const [form, setForm] = useState<RecommendationSettingsForm | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [customEventInput, setCustomEventInput] = useState("");
  const [banditStatus, setBanditStatus] = useState<BanditStatus | null>(null);
  const [configuredPolicies, setConfiguredPolicies] = useState<string[]>([]);

  const loadSettings = useCallback(async (profileId?: string) => {
    setLoading(true);
    setError(null);
    try {
      const url = new URL(
        "/api/admin/recommendation-settings",
        window.location.origin
      );
      if (profileId) {
        url.searchParams.set("profileId", profileId);
      }
      const response = await fetch(url.toString(), { cache: "no-store" });
      if (!response.ok) {
        throw new Error(`Failed to load profiles: ${response.status}`);
      }
      const data: SettingsResponse = await response.json();
      setProfiles(data.profiles ?? []);
      setBanditStatus(data.bandit ?? null);
      setConfiguredPolicies(data.configuredPolicies ?? []);
      setProfileSource(data.profile_source ?? "default");
      setForm(toFormState(data.profile));
      setSelectedProfileId(data.profile.profileId);
      setIsCreating(false);
    } catch (err) {
      console.error("Failed to load recommendation profiles", err);
      setError(err instanceof Error ? err.message : "Failed to load profiles");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadSettings();
  }, [loadSettings]);

  const eventTypesSelection = useMemo(
    () => new Set(form?.excludeEventTypes ?? []),
    [form?.excludeEventTypes]
  );

  const handleNumberChange = useCallback(
    (field: keyof RecommendationSettingsForm) =>
      (event: ChangeEvent<HTMLInputElement>) => {
        const value = event.target.value;
        setForm((prev) => (prev ? { ...prev, [field]: value } : prev));
      },
    []
  );

  const handleToggle = useCallback(
    (field: keyof RecommendationSettingsForm) =>
      (event: ChangeEvent<HTMLInputElement>) => {
        const checked = event.target.checked;
        setForm((prev) => (prev ? { ...prev, [field]: checked } : prev));
      },
    []
  );

  const toggleEventType = useCallback((eventType: string) => {
    setForm((prev) => {
      if (!prev) return prev;
      const set = new Set(prev.excludeEventTypes);
      if (set.has(eventType)) {
        set.delete(eventType);
      } else {
        set.add(eventType);
      }
      return { ...prev, excludeEventTypes: Array.from(set) };
    });
  }, []);

  const addCustomEventType = useCallback(() => {
    const value = customEventInput.trim();
    if (!value) {
      return;
    }
    setForm((prev) => {
      if (!prev) return prev;
      const normalized = value.toLowerCase();
      if (prev.excludeEventTypes.includes(normalized)) {
        return prev;
      }
      return {
        ...prev,
        excludeEventTypes: [...prev.excludeEventTypes, normalized],
      };
    });
    setCustomEventInput("");
  }, [customEventInput]);

  const removeEventType = useCallback((eventType: string) => {
    setForm((prev) => {
      if (!prev) return prev;
      const next = prev.excludeEventTypes.filter((item) => item !== eventType);
      return { ...prev, excludeEventTypes: next };
    });
  }, []);

  const handleSelectProfile = useCallback(
    (event: ChangeEvent<HTMLSelectElement>) => {
      const value = event.target.value;
      if (value) {
        loadSettings(value);
      }
    },
    [loadSettings]
  );

  const handleNewProfile = useCallback(() => {
    setIsCreating(true);
    setSelectedProfileId(null);
    setProfileSource("explicit");
    setForm({
      ...DEFAULT_FORM_VALUES,
      excludeEventTypes: [...DEFAULT_FORM_VALUES.excludeEventTypes],
    });
  }, []);

  const handleDeleteProfile = useCallback(async () => {
    if (!selectedProfileId) return;
    if (
      !confirm(
        `Delete profile '${selectedProfileId}'? A default profile will be restored automatically.`
      )
    ) {
      return;
    }
    try {
      setSaving(true);
      const response = await fetch(
        `/api/admin/recommendation-settings?profileId=${encodeURIComponent(
          selectedProfileId
        )}`,
        {
          method: "DELETE",
        }
      );
      if (!response.ok) {
        const result = await response.json().catch(() => ({}));
        const message =
          (result && result.error) ||
          `Failed to delete profile (status ${response.status})`;
        throw new Error(message);
      }
      const data: SettingsResponse = await response.json();
      setProfiles(data.profiles ?? []);
      setBanditStatus(data.bandit ?? null);
      setConfiguredPolicies(data.configuredPolicies ?? []);
      const fallbackProfile = data.profiles?.[0];
      if (fallbackProfile) {
        setSelectedProfileId(fallbackProfile.profileId);
        setForm(toFormState(fallbackProfile));
      } else {
        setSelectedProfileId(null);
        setForm({ ...DEFAULT_FORM_VALUES });
      }
      toast("Profile deleted");
    } catch (err) {
      console.error("Failed to delete profile", err);
      toast(err instanceof Error ? err.message : "Failed to delete profile");
    } finally {
      setSaving(false);
    }
  }, [selectedProfileId, toast]);

  const handleDeleteAllProfiles = useCallback(async () => {
    if (
      !confirm(
        "Delete all profiles (defaults will be re-created)? This removes any custom overrides."
      )
    ) {
      return;
    }
    try {
      setSaving(true);
      const response = await fetch(
        `/api/admin/recommendation-settings?profileId=all`,
        {
          method: "DELETE",
        }
      );
      if (!response.ok) {
        const result = await response.json().catch(() => ({}));
        const message =
          (result && result.error) ||
          `Failed to delete profiles (status ${response.status})`;
        throw new Error(message);
      }
      const data: SettingsResponse = await response.json();
      setProfiles(data.profiles ?? []);
      setBanditStatus(data.bandit ?? null);
      setConfiguredPolicies(data.configuredPolicies ?? []);
      const fallbackProfile = data.profiles?.[0];
      if (fallbackProfile) {
        setSelectedProfileId(fallbackProfile.profileId);
        setForm(toFormState(fallbackProfile));
      } else {
        setSelectedProfileId(null);
        setForm({ ...DEFAULT_FORM_VALUES });
      }
      toast("All profiles deleted");
    } catch (err) {
      console.error("Failed to delete profiles", err);
      toast(err instanceof Error ? err.message : "Failed to delete profiles");
    } finally {
      setSaving(false);
    }
  }, [toast]);

  const onSubmit = useCallback(
    async (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      if (!form) return;

      try {
        setSaving(true);
        setError(null);

        const fallbackNumber = (
          value: string,
          fallback: string,
          parser: (label: string, value: string) => number,
          label: string
        ) => parser(label, value.trim() === "" ? fallback : value);

        const profileId =
          form.profileId.trim() || DEFAULT_FORM_VALUES.profileId;
        const name = form.name.trim() || profileId;
        const payload = {
          profileId,
          name,
          description: form.description.trim() || undefined,
          surface: form.surface.trim() || undefined,
          isDefault: form.isDefault,
          blendAlpha: fallbackNumber(
            form.blendAlpha,
            DEFAULT_FORM_VALUES.blendAlpha,
            parseFloatField,
            "Blend alpha"
          ),
          blendBeta: fallbackNumber(
            form.blendBeta,
            DEFAULT_FORM_VALUES.blendBeta,
            parseFloatField,
            "Blend beta"
          ),
          blendGamma: fallbackNumber(
            form.blendGamma,
            DEFAULT_FORM_VALUES.blendGamma,
            parseFloatField,
            "Blend gamma"
          ),
          popularityHalflifeDays: fallbackNumber(
            form.popularityHalflifeDays,
            DEFAULT_FORM_VALUES.popularityHalflifeDays,
            parseFloatField,
            "Popularity halflife days"
          ),
          covisWindowDays: fallbackNumber(
            form.covisWindowDays,
            DEFAULT_FORM_VALUES.covisWindowDays,
            parseFloatField,
            "Co-visitation window days"
          ),
          popularityFanout: fallbackNumber(
            form.popularityFanout,
            DEFAULT_FORM_VALUES.popularityFanout,
            parseIntField,
            "Popularity fanout"
          ),
          mmrLambda: fallbackNumber(
            form.mmrLambda,
            DEFAULT_FORM_VALUES.mmrLambda,
            parseFloatField,
            "MMR lambda"
          ),
          brandCap: fallbackNumber(
            form.brandCap,
            DEFAULT_FORM_VALUES.brandCap,
            parseIntField,
            "Brand cap"
          ),
          categoryCap: fallbackNumber(
            form.categoryCap,
            DEFAULT_FORM_VALUES.categoryCap,
            parseIntField,
            "Category cap"
          ),
          ruleExcludeEvents: Boolean(form.ruleExcludeEvents),
          purchasedWindowDays: fallbackNumber(
            form.purchasedWindowDays,
            DEFAULT_FORM_VALUES.purchasedWindowDays,
            parseFloatField,
            "Purchased window days"
          ),
          profileWindowDays: fallbackNumber(
            form.profileWindowDays,
            DEFAULT_FORM_VALUES.profileWindowDays,
            parseFloatField,
            "Profile window days"
          ),
          profileTopN: fallbackNumber(
            form.profileTopN,
            DEFAULT_FORM_VALUES.profileTopN,
            parseIntField,
            "Profile top N"
          ),
          profileBoost: fallbackNumber(
            form.profileBoost,
            DEFAULT_FORM_VALUES.profileBoost,
            parseFloatField,
            "Profile boost"
          ),
          excludeEventTypes:
            form.excludeEventTypes.length > 0
              ? form.excludeEventTypes
              : DEFAULT_FORM_VALUES.excludeEventTypes,
        };

        if (!payload.profileId) {
          throw new Error("Profile ID is required");
        }

        const response = await fetch("/api/admin/recommendation-settings", {
          method: isCreating ? "POST" : "PUT",
          headers: { "content-type": "application/json" },
          body: JSON.stringify(payload),
        });

        if (!response.ok) {
          const result = await response.json().catch(() => ({}));
          const message =
            (result && result.error) ||
            `Failed to save profile (status ${response.status})`;
          throw new Error(message);
        }

        const data: SettingsResponse = await response.json();
        setProfiles(data.profiles ?? []);
        setBanditStatus(data.bandit ?? null);
        setConfiguredPolicies(data.configuredPolicies ?? []);
        setProfileSource(data.profile_source ?? "explicit");
        setForm(toFormState(data.profile));
        setSelectedProfileId(data.profile.profileId);
        setIsCreating(false);
        toast("Profile saved");
      } catch (err) {
        console.error("Failed to save recommendation profile", err);
        setError(err instanceof Error ? err.message : "Failed to save profile");
        toast(err instanceof Error ? err.message : "Failed to save profile");
      } finally {
        setSaving(false);
      }
    },
    [form, isCreating, toast]
  );

  if (loading) {
    return <p className="text-sm text-gray-600">Loading settings…</p>;
  }

  if (error && !form) {
    return (
      <div className="space-y-3">
        <p className="text-sm text-red-600">{error}</p>
        <button
          type="button"
          className="px-3 py-2 text-sm border rounded"
          onClick={() => loadSettings(selectedProfileId ?? undefined)}
        >
          Retry
        </button>
      </div>
    );
  }

  if (!form) {
    return null;
  }

  return (
    <section className="space-y-4">
      <header className="space-y-2">
        <div className="flex items-center justify-between gap-2">
          <div>
            <h2 className="text-lg font-semibold">Recommendation Profiles</h2>
            <p className="text-sm text-gray-600">
              Manage blend weights, diversity caps, and recency windows applied
              to recommendation requests. Profiles can be selected per-request
              or marked as defaults for a surface.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-700" htmlFor="profile-select">
              Active profile
            </label>
            <select
              id="profile-select"
              className="rounded border px-2 py-1 text-sm"
              value={selectedProfileId ?? ""}
              onChange={handleSelectProfile}
            >
              {profiles.map((profile) => (
                <option key={profile.profileId} value={profile.profileId}>
                  {profile.name} ({profile.profileId})
                </option>
              ))}
            </select>
            <button
              type="button"
              className="px-3 py-1 text-sm border rounded"
              onClick={handleNewProfile}
            >
              New profile
            </button>
            <button
              type="button"
              className="px-3 py-1 text-sm border rounded disabled:opacity-60"
              onClick={handleDeleteProfile}
              disabled={saving || !selectedProfileId}
            >
              Delete
            </button>
            <button
              type="button"
              className="px-3 py-1 text-sm border rounded disabled:opacity-60"
              onClick={handleDeleteAllProfiles}
              disabled={saving || profiles.length === 0}
            >
              Nuke all
            </button>
          </div>
        </div>
        <p className="text-xs text-gray-500">
          Loaded via: {profileSource}. Last updated: {""}
          {form.updatedAt ? new Date(form.updatedAt).toLocaleString() : "—"}
        </p>
      </header>

      {banditStatus ? (
        <div
          className={`rounded border p-3 text-sm ${
            banditStatus.enabled
              ? "border-green-500 bg-green-50 text-green-800"
              : "border-amber-500 bg-amber-50 text-amber-800"
          }`}
        >
          <p className="font-medium">
            Bandit exploration {banditStatus.enabled ? "enabled" : "disabled"}
          </p>
          <p>
            {banditStatus.enabled
              ? "Requests will call /v1/bandit/recommendations before falling back to ranking."
              : banditStatus.reason ?? "Exploration disabled by configuration."}
          </p>
          <p className="mt-1 text-xs opacity-80">
            Checked {new Date(banditStatus.checkedAt).toLocaleTimeString()} ·
            Configured policies:{" "}
            {configuredPolicies.length > 0
              ? configuredPolicies.join(", ")
              : "—"}
          </p>
          {!banditStatus.enabled && banditStatus.missingPolicies?.length ? (
            <p className="mt-1 text-xs opacity-80">
              Missing policies in recsys:{" "}
              {banditStatus.missingPolicies.join(", ")}
            </p>
          ) : null}
        </div>
      ) : null}

      {error ? <p className="text-sm text-red-600">{error}</p> : null}

      <form onSubmit={onSubmit} className="space-y-6">
        <fieldset className="space-y-3 border rounded p-3">
          <legend className="px-1 text-sm font-medium text-gray-700">
            Profile metadata
          </legend>
          <label className="block text-sm">
            <span className="text-gray-700">Profile ID</span>
            <input
              type="text"
              value={form.profileId}
              onChange={(event) =>
                setForm((prev) =>
                  prev ? { ...prev, profileId: event.target.value } : prev
                )
              }
              readOnly={!isCreating}
              className={`mt-1 w-full rounded border px-2 py-1 text-sm ${
                isCreating ? "bg-white" : "bg-gray-100"
              }`}
              placeholder="default"
            />
          </label>
          <label className="block text-sm">
            <span className="text-gray-700">Name</span>
            <input
              type="text"
              value={form.name}
              onChange={(event) =>
                setForm((prev) =>
                  prev ? { ...prev, name: event.target.value } : prev
                )
              }
              className="mt-1 w-full rounded border px-2 py-1 text-sm"
              placeholder="Homepage default"
            />
          </label>
          <label className="block text-sm">
            <span className="text-gray-700">Description</span>
            <textarea
              value={form.description}
              onChange={(event) =>
                setForm((prev) =>
                  prev ? { ...prev, description: event.target.value } : prev
                )
              }
              className="mt-1 w-full rounded border px-2 py-1 text-sm"
              rows={2}
            />
          </label>
          <label className="block text-sm">
            <span className="text-gray-700">Surface (optional)</span>
            <input
              type="text"
              value={form.surface}
              onChange={(event) =>
                setForm((prev) =>
                  prev ? { ...prev, surface: event.target.value } : prev
                )
              }
              className="mt-1 w-full rounded border px-2 py-1 text-sm"
              placeholder="home"
            />
          </label>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={form.isDefault}
              onChange={handleToggle("isDefault")}
            />
            Mark as default for this surface
          </label>
        </fieldset>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <fieldset className="space-y-3 border rounded p-3">
            <legend className="px-1 text-sm font-medium text-gray-700">
              Blend Weights
            </legend>
            <NumberField
              label="Blend alpha (popularity)"
              value={form.blendAlpha}
              step="0.05"
              min="0"
              onChange={handleNumberChange("blendAlpha")}
            />
            <NumberField
              label="Blend beta (co-visitation)"
              value={form.blendBeta}
              step="0.05"
              min="0"
              onChange={handleNumberChange("blendBeta")}
            />
            <NumberField
              label="Blend gamma (ALS)"
              value={form.blendGamma}
              step="0.05"
              min="0"
              onChange={handleNumberChange("blendGamma")}
            />
            <NumberField
              label="MMR lambda"
              value={form.mmrLambda}
              step="0.05"
              min="0"
              onChange={handleNumberChange("mmrLambda")}
            />
          </fieldset>

          <fieldset className="space-y-3 border rounded p-3">
            <legend className="px-1 text-sm font-medium text-gray-700">
              Recency Windows
            </legend>
            <NumberField
              label="Popularity halflife (days)"
              value={form.popularityHalflifeDays}
              step="0.5"
              min="0"
              onChange={handleNumberChange("popularityHalflifeDays")}
            />
            <NumberField
              label="Co-visitation window (days)"
              value={form.covisWindowDays}
              step="1"
              min="0"
              onChange={handleNumberChange("covisWindowDays")}
            />
            <NumberField
              label="Purchased window (days)"
              value={form.purchasedWindowDays}
              step="1"
              min="0"
              onChange={handleNumberChange("purchasedWindowDays")}
            />
            <NumberField
              label="Profile window (days)"
              value={form.profileWindowDays}
              step="1"
              min="0"
              onChange={handleNumberChange("profileWindowDays")}
            />
          </fieldset>

          <fieldset className="space-y-3 border rounded p-3">
            <legend className="px-1 text-sm font-medium text-gray-700">
              Diversity & Coverage
            </legend>
            <NumberField
              label="Brand cap"
              value={form.brandCap}
              step="1"
              min="0"
              onChange={handleNumberChange("brandCap")}
            />
            <NumberField
              label="Category cap"
              value={form.categoryCap}
              step="1"
              min="0"
              onChange={handleNumberChange("categoryCap")}
            />
            <NumberField
              label="Popularity fanout"
              value={form.popularityFanout}
              step="1"
              min="1"
              onChange={handleNumberChange("popularityFanout")}
            />
            <NumberField
              label="Profile top N tags"
              value={form.profileTopN}
              step="1"
              min="0"
              onChange={handleNumberChange("profileTopN")}
            />
            <NumberField
              label="Profile boost"
              value={form.profileBoost}
              step="0.05"
              min="0"
              onChange={handleNumberChange("profileBoost")}
            />
          </fieldset>

          <fieldset className="space-y-3 border rounded p-3">
            <legend className="px-1 text-sm font-medium text-gray-700">
              Event Controls
            </legend>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={form.ruleExcludeEvents}
                onChange={handleToggle("ruleExcludeEvents")}
              />
              Exclude events in rule engine
            </label>
            <div className="space-y-2">
              <p className="text-sm text-gray-700">Exclude event types</p>
              <div className="flex flex-wrap gap-2">
                {KNOWN_EVENT_TYPES.map((eventType) => (
                  <label
                    key={eventType}
                    className="flex items-center gap-1 text-sm"
                  >
                    <input
                      type="checkbox"
                      checked={eventTypesSelection.has(eventType)}
                      onChange={() => toggleEventType(eventType)}
                    />
                    {eventType}
                  </label>
                ))}
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  className="flex-1 rounded border px-2 py-1 text-sm"
                  placeholder="custom event type"
                  value={customEventInput}
                  onChange={(event) => setCustomEventInput(event.target.value)}
                />
                <button
                  type="button"
                  className="px-3 py-1 text-sm border rounded"
                  onClick={addCustomEventType}
                >
                  Add
                </button>
              </div>
              {form.excludeEventTypes.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {form.excludeEventTypes.map((eventType) => (
                    <span
                      key={eventType}
                      className="inline-flex items-center gap-1 rounded bg-gray-200 px-2 py-1 text-xs"
                    >
                      {eventType}
                      <button
                        type="button"
                        className="text-gray-600 hover:text-gray-900"
                        onClick={() => removeEventType(eventType)}
                        aria-label={`Remove ${eventType}`}
                      >
                        ×
                      </button>
                    </span>
                  ))}
                </div>
              ) : (
                <p className="text-xs text-gray-500">
                  No event types excluded.
                </p>
              )}
            </div>
          </fieldset>
        </div>

        <div className="flex items-center gap-2">
          <button
            type="submit"
            className="px-4 py-2 text-sm border rounded bg-blue-600 text-white disabled:opacity-60"
            disabled={saving}
          >
            {saving
              ? "Saving…"
              : isCreating
              ? "Create profile"
              : "Save profile"}
          </button>
        </div>
      </form>
    </section>
  );
}

type NumberFieldProps = {
  label: string;
  value: string;
  onChange: (event: ChangeEvent<HTMLInputElement>) => void;
  step?: string;
  min?: string;
};

function NumberField({ label, value, onChange, step, min }: NumberFieldProps) {
  return (
    <label className="block text-sm">
      <span className="text-gray-700">{label}</span>
      <input
        type="number"
        value={value}
        onChange={onChange}
        step={step}
        min={min}
        className="mt-1 w-full rounded border px-2 py-1 text-sm"
      />
    </label>
  );
}
