import React, { useState, useEffect, useRef } from "react";
import { Section, Row, Label, Button } from "./UIComponents";
import type { types_Overrides } from "../lib/api-client";
import {
  ALGORITHM_PROFILES,
  type AlgorithmProfile,
} from "../types/algorithmProfiles";
import { ProfileEditor } from "./ProfileEditor";

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
      <div style={{ marginBottom: 16 }}>
        <p style={{ color: "#666", fontSize: 14, marginBottom: 16 }}>
          Override algorithm parameters to test different recommendation
          strategies. Select a profile below or create custom profiles.
        </p>

        {/* Action Buttons */}
        <div style={{ marginBottom: 20 }}>
          <Row>
            <Button
              onClick={handleCreateNewProfile}
              style={{
                backgroundColor: "#28a745",
                color: "white",
                border: "none",
                padding: "8px 16px",
                borderRadius: 4,
                cursor: "pointer",
                marginRight: 8,
                fontSize: 14,
              }}
            >
              + New Custom Profile
            </Button>
            {hasOverrides && (
              <Button
                onClick={handleClearOverrides}
                style={{
                  backgroundColor: "#dc3545",
                  color: "white",
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
        <div style={{ marginBottom: 20 }}>
          <Label text="Built-in Algorithm Profiles">
            <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
              {ALGORITHM_PROFILES.map((profile) => (
                <div key={profile.id} style={{ position: "relative" }}>
                  <Button
                    onClick={() => handleProfileSelect(profile)}
                    style={{
                      backgroundColor:
                        selectedProfileId === profile.id
                          ? "#007acc"
                          : "#f5f5f5",
                      color:
                        selectedProfileId === profile.id ? "white" : "#333",
                      border: "1px solid #ddd",
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
                    <div style={{ fontSize: 11, opacity: 0.8 }}>
                      {profile.description}
                    </div>
                  </Button>
                  <button
                    onClick={() => handleCopyProfile(profile)}
                    style={{
                      position: "absolute",
                      top: 4,
                      right: 4,
                      backgroundColor: "#6c757d",
                      color: "white",
                      border: "none",
                      borderRadius: 2,
                      padding: "2px 4px",
                      fontSize: 10,
                      cursor: "pointer",
                    }}
                    title="Copy profile"
                  >
                    📋
                  </button>
                </div>
              ))}
            </div>
          </Label>
        </div>

        {/* Custom Profiles */}
        <div style={{ marginBottom: 20 }}>
          <Label text={`Custom Profiles (${customProfiles.length})`}>
            {customProfiles.length > 0 ? (
              <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
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
                            ? "#007acc"
                            : highlightedProfileId === profile.id
                            ? "#28a745"
                            : "#e9ecef",
                        color:
                          selectedProfileId === profile.id ||
                          highlightedProfileId === profile.id
                            ? "white"
                            : "#333",
                        border:
                          highlightedProfileId === profile.id
                            ? "2px solid #28a745"
                            : "1px solid #ddd",
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
                              color: "#28a745",
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
                      <div style={{ fontSize: 11, opacity: 0.8 }}>
                        {profile.description}
                      </div>
                    </Button>
                    <button
                      onClick={() => handleCopyProfile(profile)}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 56,
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                      }}
                      title="Copy profile"
                    >
                      📋
                    </button>
                    <button
                      onClick={() => handleEditProfile(profile)}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 30,
                        backgroundColor: "#007acc",
                        color: "white",
                        border: "none",
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                      }}
                      title="Edit profile"
                    >
                      ✏️
                    </button>
                    <button
                      onClick={() => handleDeleteProfile(profile.id)}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 4,
                        backgroundColor: "#dc3545",
                        color: "white",
                        border: "none",
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                      }}
                      title="Delete profile"
                    >
                      🗑️
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <div
                style={{
                  backgroundColor: "#f8f9fa",
                  border: "1px solid #e9ecef",
                  borderRadius: 4,
                  padding: 16,
                  textAlign: "center",
                  color: "#666",
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
          <div style={{ marginTop: 16 }}>
            <Label text="Active Overrides">
              <div
                style={{
                  backgroundColor: "#f8f9fa",
                  border: "1px solid #e9ecef",
                  borderRadius: 4,
                  padding: 12,
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
                      <span style={{ color: "#666" }}>{key}:</span>{" "}
                      <span style={{ color: "#333" }}>
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
              backgroundColor: "#f8f9fa",
              border: "1px solid #e9ecef",
              borderRadius: 4,
              padding: 16,
              textAlign: "center",
              color: "#666",
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
