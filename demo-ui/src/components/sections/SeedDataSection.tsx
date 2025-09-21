import React, { useState } from "react";
import {
  Section,
  Row,
  Label,
  NumberInput,
  Button,
  Code,
} from "../primitives/UIComponents";
import { handleSeed } from "../../services/seedingService";
import { UserTraitsEditor, type TraitConfig } from "./UserTraitsEditor";
import {
  ItemConfigEditor,
  type ItemConfig,
  type PriceRange,
} from "./ItemConfigEditor";
import { EventTypeEditor } from "./EventTypeEditor";
import type { EventTypeConfig } from "../../types";

interface SeedDataSectionProps {
  userCount: number;
  setUserCount: (value: number) => void;
  userStartIndex: number;
  setUserStartIndex: (value: number) => void;
  itemCount: number;
  setItemCount: (value: number) => void;
  minEventsPerUser: number;
  setMinEventsPerUser: (value: number) => void;
  maxEventsPerUser: number;
  setMaxEventsPerUser: (value: number) => void;
  eventTypes: EventTypeConfig[];
  setEventTypes: (value: EventTypeConfig[]) => void;
  namespace: string;
  brands: string[];
  tags: string[];
  log: string;
  setLog: React.Dispatch<React.SetStateAction<string>>;
  setGeneratedUsers: (value: string[]) => void;
  setGeneratedItems: (value: string[]) => void;
  traitConfigs: TraitConfig[];
  setTraitConfigs: (value: TraitConfig[]) => void;
  itemConfigs: ItemConfig[];
  setItemConfigs: (value: ItemConfig[]) => void;
  priceRanges: PriceRange[];
  setPriceRanges: (value: PriceRange[]) => void;
  generatedUsers: string[];
  generatedItems: string[];
  onUpdateUser: (
    _userId: string,
    _traits: Record<string, any>
  ) => Promise<void>;
  onUpdateItem: (
    _itemId: string,
    _updates: Record<string, any>
  ) => Promise<void>;
}

export function SeedDataSection({
  userCount,
  setUserCount,
  userStartIndex,
  setUserStartIndex,
  itemCount,
  setItemCount,
  minEventsPerUser,
  setMinEventsPerUser,
  maxEventsPerUser,
  setMaxEventsPerUser,
  eventTypes,
  setEventTypes,
  namespace,
  brands,
  tags,
  log,
  setLog,
  setGeneratedUsers,
  setGeneratedItems,
  traitConfigs,
  setTraitConfigs,
  itemConfigs,
  setItemConfigs,
  priceRanges,
  setPriceRanges,
  generatedUsers,
  generatedItems,
  onUpdateUser,
  onUpdateItem,
}: SeedDataSectionProps) {
  const [isTraitsEditorOpen, setIsTraitsEditorOpen] = useState(false);
  const [isItemConfigEditorOpen, setIsItemConfigEditorOpen] = useState(false);

  const append = (s: string) => {
    setLog((prev) => `${prev}${prev ? "\n" : ""}${s}`);
  };

  // Generate preview of trait configuration
  const getTraitPreview = () => {
    if (traitConfigs.length === 0) {
      return "No traits configured (will use default 'plan' trait)";
    }

    return traitConfigs
      .map((config) => {
        const values = config.values
          .map((v) => `${v.value} (${Math.round(v.probability * 100)}%)`)
          .join(", ");
        return `${config.key} (${Math.round(
          config.probability * 100
        )}% chance): ${values}`;
      })
      .join("; ");
  };

  // Generate preview of item configuration
  const getItemConfigPreview = () => {
    const parts: string[] = [];

    if (priceRanges.length > 0) {
      const priceInfo = priceRanges
        .map(
          (range) =>
            `$${range.min}-$${range.max} (${Math.round(
              range.probability * 100
            )}%)`
        )
        .join(", ");
      parts.push(`Price ranges: ${priceInfo}`);
    }

    if (itemConfigs.length > 0) {
      const configInfo = itemConfigs
        .map((config) => {
          const values = config.values
            .map((v) => `${v.value} (${Math.round(v.probability * 100)}%)`)
            .join(", ");
          return `${config.key} (${Math.round(
            config.probability * 100
          )}% chance): ${values}`;
        })
        .join("; ");
      parts.push(`Properties: ${configInfo}`);
    }

    return parts.length > 0
      ? parts.join("; ")
      : "No item configuration (will use defaults)";
  };

  const onSeed = () => {
    setLog("");
    handleSeed(
      namespace,
      userCount,
      userStartIndex,
      itemCount,
      minEventsPerUser,
      maxEventsPerUser,
      brands,
      tags,
      eventTypes,
      append,
      setGeneratedUsers,
      setGeneratedItems,
      traitConfigs,
      itemConfigs,
      priceRanges
    );
  };

  return (
    <Section title="Seed Data">
      {/* Data Counts Sub-section */}
      <div
        style={{
          border: "1px solid #e0e0e0",
          borderRadius: 6,
          padding: 12,
          marginBottom: 16,
          backgroundColor: "#fafafa",
        }}
      >
        <h3
          style={{ marginTop: 0, marginBottom: 8, fontSize: 14, color: "#333" }}
        >
          Data Counts
        </h3>
        <p style={{ color: "#666", fontSize: 12, marginBottom: 12 }}>
          Configure how many users, items, and events per user to generate.
        </p>
        <Row>
          <Label text="Users">
            <NumberInput
              min={1}
              value={userCount}
              onChange={(e) => setUserCount(Number(e.target.value))}
            />
          </Label>
          <Label text="User Start Index">
            <NumberInput
              min={1}
              value={userStartIndex}
              onChange={(e) => setUserStartIndex(Number(e.target.value))}
            />
          </Label>
        </Row>
        <div style={{ marginTop: 8 }}>
          <p
            style={{
              color: "#888",
              fontSize: 11,
              marginTop: 4,
              marginBottom: 0,
            }}
          >
            User IDs will range from user-
            {String(userStartIndex).padStart(4, "0")} to user-
            {String(userStartIndex + userCount - 1).padStart(4, "0")} (
            {userCount} users total)
          </p>
        </div>
        <div style={{ marginTop: 8 }}>
          <Row>
            <Label text="Min Events per User">
              <NumberInput
                min={0}
                value={minEventsPerUser}
                onChange={(e) => setMinEventsPerUser(Number(e.target.value))}
              />
            </Label>
            <Label text="Max Events per User">
              <NumberInput
                min={minEventsPerUser}
                value={maxEventsPerUser}
                onChange={(e) => setMaxEventsPerUser(Number(e.target.value))}
              />
            </Label>
          </Row>
          <p
            style={{
              color: "#888",
              fontSize: 11,
              marginTop: 4,
              marginBottom: 0,
            }}
          >
            Each user will get a random number of events between min and max.
            Set min=max for consistent events per user.
          </p>
        </div>
        <Row>
          <Label text="Items">
            <NumberInput
              min={1}
              value={itemCount}
              onChange={(e) => setItemCount(Number(e.target.value))}
            />
          </Label>
        </Row>
      </div>

      {/* Event Type Configuration */}
      <EventTypeEditor eventTypes={eventTypes} setEventTypes={setEventTypes} />

      {/* User Traits Configuration Sub-section */}
      <div
        style={{
          border: "1px solid #e0e0e0",
          borderRadius: 6,
          padding: 12,
          marginBottom: 16,
          backgroundColor: "#fafafa",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            marginBottom: 8,
          }}
        >
          <h3
            style={{
              marginTop: 0,
              marginBottom: 0,
              fontSize: 14,
              color: "#333",
            }}
          >
            User Traits Configuration
          </h3>
          <Button
            onClick={() => setIsTraitsEditorOpen(!isTraitsEditorOpen)}
            style={{
              padding: "4px 8px",
              fontSize: 12,
              backgroundColor: isTraitsEditorOpen ? "#e3f2fd" : "#f5f5f5",
              color: isTraitsEditorOpen ? "#1565c0" : "#666",
            }}
          >
            {isTraitsEditorOpen ? "▼ Hide" : "▶ Configure"}
          </Button>
        </div>

        <p style={{ color: "#666", fontSize: 12, marginBottom: 8 }}>
          Configure dynamic user traits that will be randomly assigned during
          seeding.
        </p>

        {/* Preview of current configuration */}
        <div
          style={{
            backgroundColor: "#f0f0f0",
            border: "1px solid #ddd",
            borderRadius: 4,
            padding: 8,
            fontSize: 11,
            color: "#555",
            fontFamily: "monospace",
            marginBottom: 8,
          }}
        >
          <strong>Preview:</strong> {getTraitPreview()}
        </div>

        {/* Accordion content */}
        {isTraitsEditorOpen && (
          <div style={{ marginTop: 12 }}>
            <UserTraitsEditor
              traitConfigs={traitConfigs}
              setTraitConfigs={setTraitConfigs}
              generatedUsers={generatedUsers}
              onUpdateUser={onUpdateUser}
            />
          </div>
        )}
      </div>

      {/* Item Configuration Sub-section */}
      <div
        style={{
          border: "1px solid #e0e0e0",
          borderRadius: 6,
          padding: 12,
          marginBottom: 16,
          backgroundColor: "#fafafa",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            marginBottom: 8,
          }}
        >
          <h3
            style={{
              marginTop: 0,
              marginBottom: 0,
              fontSize: 14,
              color: "#333",
            }}
          >
            Item Configuration
          </h3>
          <Button
            onClick={() => setIsItemConfigEditorOpen(!isItemConfigEditorOpen)}
            style={{
              padding: "4px 8px",
              fontSize: 12,
              backgroundColor: isItemConfigEditorOpen ? "#e3f2fd" : "#f5f5f5",
              color: isItemConfigEditorOpen ? "#1565c0" : "#666",
            }}
          >
            {isItemConfigEditorOpen ? "▼ Hide" : "▶ Configure"}
          </Button>
        </div>

        <p style={{ color: "#666", fontSize: 12, marginBottom: 8 }}>
          Configure dynamic item properties and price ranges that will be
          randomly assigned during seeding.
        </p>

        {/* Preview of current configuration */}
        <div
          style={{
            backgroundColor: "#f0f0f0",
            border: "1px solid #ddd",
            borderRadius: 4,
            padding: 8,
            fontSize: 11,
            color: "#555",
            fontFamily: "monospace",
            marginBottom: 8,
          }}
        >
          <strong>Preview:</strong> {getItemConfigPreview()}
        </div>

        {/* Accordion content */}
        {isItemConfigEditorOpen && (
          <div style={{ marginTop: 12 }}>
            <ItemConfigEditor
              itemConfigs={itemConfigs}
              setItemConfigs={setItemConfigs}
              priceRanges={priceRanges}
              setPriceRanges={setPriceRanges}
              generatedItems={generatedItems}
              onUpdateItem={onUpdateItem}
            />
          </div>
        )}
      </div>

      {/* Seed Action */}
      <div style={{ textAlign: "center", marginBottom: 12 }}>
        <Button onClick={onSeed} style={{ padding: "12px 24px", fontSize: 16 }}>
          Seed namespace
        </Button>
      </div>

      {/* Log Output */}
      <Code>{log || "Ready."}</Code>
    </Section>
  );
}
