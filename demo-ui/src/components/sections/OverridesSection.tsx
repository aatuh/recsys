import React, { useState, useEffect, useRef } from "react";
import { Section, Row, Label, Button } from "../primitives/UIComponents";
import type { types_Overrides } from "../../lib/api-client";
import {
  ALGORITHM_PROFILES,
  type AlgorithmProfile,
} from "../../types/algorithmProfiles";
import { ProfileEditor } from "./ProfileEditor";
import { useToast } from "../../ui/Toast";
import { color, spacing } from "../../ui/tokens";

interface CustomProfile {
  id: string;
  name: string;
  description: string;
  overrides: types_Overrides;
}

interface OverridesSectionProps {
  overrides: types_Overrides | null;
  setOverrides: (value: types_Overrides | null) => void;
  customProfiles: CustomProfile[];
  setCustomProfiles: (value: CustomProfile[]) => void;
  selectedProfileId: string | null;
  setSelectedProfileId: (value: string | null) => void;
  isEditingProfile: boolean;
  setIsEditingProfile: (value: boolean) => void;
}

export function OverridesSection({
  overrides,
  setOverrides,
  customProfiles,
  setCustomProfiles,
  selectedProfileId,
  setSelectedProfileId,
  isEditingProfile,
  setIsEditingProfile,
}: OverridesSectionProps) {
  const [editingProfile, setEditingProfile] = useState<CustomProfile | null>(
    null
  );
  const [highlightedProfileId, setHighlightedProfileId] = useState<
    string | null
  >(null);
  const prevCustomProfilesRef = useRef<string>("");
  const hasLoadedFromStorageRef = useRef<boolean>(false);
  const toast = useToast();

  // Load custom profiles from localStorage on mount
  useEffect(() => {
    if (hasLoadedFromStorageRef.current) {
      return; // Only load once
    }

    const saved = localStorage.getItem("recsys-custom-profiles");
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        setCustomProfiles(parsed);
        prevCustomProfilesRef.current = saved; // Initialize ref with loaded data
      } catch (e) {
        console.warn("Failed to load custom profiles from localStorage:", e);
        prevCustomProfilesRef.current = JSON.stringify([]);
        toast.error("Failed to load custom profiles");
      }
    } else {
      // Initialize ref with empty array
      prevCustomProfilesRef.current = JSON.stringify([]);
    }

    hasLoadedFromStorageRef.current = true;
  }, []); // Empty dependency array - only run once on mount

  // Save custom profiles to localStorage whenever they change
  useEffect(() => {
    const currentProfilesString = JSON.stringify(customProfiles);

    // Only save if the profiles have actually changed
    if (currentProfilesString !== prevCustomProfilesRef.current) {
      prevCustomProfilesRef.current = currentProfilesString;
      localStorage.setItem("recsys-custom-profiles", currentProfilesString);
    }
  }, [customProfiles]);

  const handleProfileSelect = (profile: AlgorithmProfile | CustomProfile) => {
    setSelectedProfileId(profile.id);
    setOverrides(profile.overrides);
  };

  const handleClearOverrides = () => {
    setSelectedProfileId(null);
    setOverrides(null);
  };

  const handleCreateNewProfile = () => {
    const newProfile: CustomProfile = {
      id: `custom-${Date.now()}`,
      name: "New Custom Profile",
      description: "A custom algorithm profile",
      overrides: {},
    };
    setEditingProfile(newProfile);
    setIsEditingProfile(true);
  };

  const handleCopyProfile = (profile: AlgorithmProfile | CustomProfile) => {
    const copiedProfile: CustomProfile = {
      id: `custom-${Date.now()}`,
      name: `${profile.name} (Copy)`,
      description: `Copy of ${profile.description}`,
      overrides: { ...profile.overrides },
    };
    setEditingProfile(copiedProfile);
    setIsEditingProfile(true);
  };

  const handleEditProfile = (profile: CustomProfile) => {
    setEditingProfile(profile);
    setIsEditingProfile(true);
  };

  const handleDeleteProfile = (profileId: string) => {
    setCustomProfiles(customProfiles.filter((p) => p.id !== profileId));
    if (selectedProfileId === profileId) {
      setSelectedProfileId(null);
      setOverrides(null);
    }
    toast.success("Profile deleted");
  };

  const handleSaveProfile = (
    name: string,
    description: string,
    overrides: types_Overrides
  ) => {
    if (editingProfile) {
      const updatedProfile: CustomProfile = {
        ...editingProfile,
        name,
        description,
        overrides,
      };

      if (
        editingProfile.id.startsWith("custom-") &&
        customProfiles.find((p) => p.id === editingProfile.id)
      ) {
        // Update existing custom profile
        setCustomProfiles(
          customProfiles.map((p) =>
            p.id === editingProfile.id ? updatedProfile : p
          )
        );
      } else {
        // Add new custom profile
        setCustomProfiles([...customProfiles, updatedProfile]);
      }

      setSelectedProfileId(updatedProfile.id);
      setOverrides(overrides);

      // Highlight and scroll to the newly created/updated profile
      setHighlightedProfileId(updatedProfile.id);
      setTimeout(() => {
        const profileElement = document.getElementById(
          `profile-${updatedProfile.id}`
        );
        if (profileElement) {
          profileElement.scrollIntoView({
            behavior: "smooth",
            block: "center",
          });
        }
      }, 100);

      // Remove highlight after animation
      setTimeout(() => {
        setHighlightedProfileId(null);
      }, 2000);
    }
    setIsEditingProfile(false);
    setEditingProfile(null);
  };

  const handleCancelEdit = () => {
    setIsEditingProfile(false);
    setEditingProfile(null);
  };

  const handleClearField = (_field: keyof types_Overrides) => {
    // This will be handled by the ProfileEditor component
  };

  const hasOverrides = overrides !== null;

  if (isEditingProfile && editingProfile) {
    return (
      <ProfileEditor
        profileName={editingProfile.name}
        profileDescription={editingProfile.description}
        overrides={editingProfile.overrides}
        onSave={handleSaveProfile}
        onCancel={handleCancelEdit}
        onClearField={handleClearField}
      />
    );
  }

  return (
    <Section title="Algorithm Overrides">
      <div style={{ marginBottom: spacing.lg }}>
        <p
          style={{
            color: color.textMuted,
            fontSize: 14,
            marginBottom: spacing.md,
          }}
        >
          Override algorithm parameters to test different recommendation
          strategies. Select a profile below or create custom profiles.
        </p>

        {/* Action Buttons */}
        <div style={{ marginBottom: spacing.lg }}>
          <Row>
            <Button
              onClick={handleCreateNewProfile}
              style={{
                backgroundColor: color.success,
                color: color.primaryTextOn,
                border: "none",
                padding: "8px 16px",
                borderRadius: 4,
                cursor: "pointer",
                marginRight: spacing.md,
                fontSize: 14,
              }}
            >
              + New Custom Profile
            </Button>
            {hasOverrides && (
              <Button
                onClick={handleClearOverrides}
                style={{
                  backgroundColor: color.danger,
                  color: color.primaryTextOn,
                  border: "none",
                  padding: "8px 16px",
                  borderRadius: 4,
                  cursor: "pointer",
                  fontSize: 14,
                }}
              >
                Clear Overrides
              </Button>
            )}
          </Row>
        </div>

        {/* Built-in Algorithm Profiles */}
        <div style={{ marginBottom: spacing.lg }}>
          <Label text="Built-in Algorithm Profiles">
            <div style={{ display: "flex", flexWrap: "wrap", gap: spacing.sm }}>
              {ALGORITHM_PROFILES.map((profile) => (
                <div key={profile.id} style={{ position: "relative" }}>
                  <Button
                    onClick={() => handleProfileSelect(profile)}
                    style={{
                      backgroundColor:
                        selectedProfileId === profile.id
                          ? color.primary
                          : color.buttonBg,
                      color:
                        selectedProfileId === profile.id
                          ? color.primaryTextOn
                          : color.text,
                      border: `1px solid ${color.buttonBorder}`,
                      padding: "8px 12px",
                      borderRadius: 4,
                      fontSize: 12,
                      cursor: "pointer",
                      textAlign: "left",
                      minWidth: 200,
                      paddingRight: 40,
                    }}
                  >
                    <div style={{ fontWeight: "bold", marginBottom: 2 }}>
                      {profile.name}
                    </div>
                    <div style={{ fontSize: 11, color: color.textMuted }}>
                      {profile.description}
                    </div>
                  </Button>
                  <button
                    onClick={() => handleCopyProfile(profile)}
                    style={{
                      position: "absolute",
                      top: 4,
                      right: 4,
                      backgroundColor: color.buttonBg,
                      color: color.text,
                      border: `1px solid ${color.buttonBorder}`,
                      borderRadius: 2,
                      padding: "2px 4px",
                      fontSize: 10,
                      cursor: "pointer",
                    }}
                    title="Copy profile"
                    aria-label={`Copy built-in profile ${profile.name}`}
                  >
                    📋
                  </button>
                </div>
              ))}
            </div>
          </Label>
        </div>

        {/* Custom Profiles */}
        <div style={{ marginBottom: spacing.lg }}>
          <Label text={`Custom Profiles (${customProfiles.length})`}>
            {customProfiles.length > 0 ? (
              <div
                style={{ display: "flex", flexWrap: "wrap", gap: spacing.sm }}
              >
                {customProfiles.map((profile) => (
                  <div
                    key={profile.id}
                    id={`profile-${profile.id}`}
                    style={{ position: "relative" }}
                  >
                    <Button
                      onClick={() => handleProfileSelect(profile)}
                      style={{
                        backgroundColor:
                          selectedProfileId === profile.id
                            ? color.primary
                            : highlightedProfileId === profile.id
                            ? color.success
                            : color.buttonBg,
                        color:
                          selectedProfileId === profile.id ||
                          highlightedProfileId === profile.id
                            ? color.primaryTextOn
                            : color.text,
                        border:
                          highlightedProfileId === profile.id
                            ? `2px solid ${color.success}`
                            : `1px solid ${color.buttonBorder}`,
                        padding: "8px 12px",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                        textAlign: "left",
                        minWidth: 200,
                        paddingRight: 86,
                        transition: "all 0.3s ease",
                        transform:
                          highlightedProfileId === profile.id
                            ? "scale(1.02)"
                            : "scale(1)",
                        boxShadow:
                          highlightedProfileId === profile.id
                            ? "0 4px 8px rgba(40, 167, 69, 0.3)"
                            : "none",
                      }}
                    >
                      <div
                        style={{
                          fontWeight: "bold",
                          marginBottom: 2,
                          display: "flex",
                          alignItems: "center",
                          gap: 4,
                        }}
                      >
                        {profile.name}
                        {highlightedProfileId === profile.id && (
                          <span
                            style={{
                              backgroundColor: "rgba(255, 255, 255, 0.9)",
                              color: color.success,
                              fontSize: 8,
                              fontWeight: "bold",
                              padding: "1px 4px",
                              borderRadius: 2,
                              textTransform: "uppercase",
                            }}
                          >
                            NEW
                          </span>
                        )}
                      </div>
                      <div style={{ fontSize: 11, color: color.textMuted }}>
                        {profile.description}
                      </div>
                    </Button>
                    <button
                      onClick={() => handleCopyProfile(profile)}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 56,
                        backgroundColor: color.buttonBg,
                        color: color.text,
                        border: `1px solid ${color.buttonBorder}`,
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                      }}
                      title="Copy profile"
                      aria-label={`Copy custom profile ${profile.name}`}
                    >
                      📋
                    </button>
                    <button
                      onClick={() => handleEditProfile(profile)}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 30,
                        backgroundColor: color.primary,
                        color: color.primaryTextOn,
                        border: "none",
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                      }}
                      title="Edit profile"
                      aria-label={`Edit custom profile ${profile.name}`}
                    >
                      ✏️
                    </button>
                    <button
                      onClick={() => handleDeleteProfile(profile.id)}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 4,
                        backgroundColor: color.danger,
                        color: color.primaryTextOn,
                        border: "none",
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                      }}
                      title="Delete profile"
                      aria-label={`Delete custom profile ${profile.name}`}
                    >
                      🗑️
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <div
                style={{
                  backgroundColor: color.panelSubtle,
                  border: `1px solid ${color.panelBorder}`,
                  borderRadius: 4,
                  padding: spacing.md,
                  textAlign: "center",
                  color: color.textMuted,
                  fontSize: 14,
                }}
              >
                No custom profiles yet. Create one using the "New Custom
                Profile" button above or copy an existing built-in profile.
              </div>
            )}
          </Label>
        </div>

        {/* Current Overrides Display */}
        {hasOverrides && (
          <div style={{ marginTop: spacing.md }}>
            <Label text="Active Overrides">
              <div
                style={{
                  backgroundColor: color.panelSubtle,
                  border: `1px solid ${color.panelBorder}`,
                  borderRadius: 4,
                  padding: spacing.md,
                  fontSize: 12,
                  fontFamily: "monospace",
                  maxHeight: 200,
                  overflowY: "auto",
                }}
              >
                {Object.entries(overrides)
                  .filter(([, value]) => value !== undefined && value !== null)
                  .map(([key, value]) => (
                    <div key={key} style={{ marginBottom: 4 }}>
                      <span style={{ color: color.textMuted }}>{key}:</span>{" "}
                      <span style={{ color: color.text }}>
                        {typeof value === "boolean" ? value.toString() : value}
                      </span>
                    </div>
                  ))}
              </div>
            </Label>
          </div>
        )}

        {!hasOverrides && (
          <div
            style={{
              backgroundColor: color.panelSubtle,
              border: `1px solid ${color.panelBorder}`,
              borderRadius: 4,
              padding: spacing.md,
              textAlign: "center",
              color: color.textMuted,
              fontSize: 14,
            }}
          >
            No overrides active. Using default environment variable values.
          </div>
        )}
      </div>
    </Section>
  );
}
