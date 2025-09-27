import React, { useState, useEffect } from "react";
import { getStorage } from "../../di";
import { Button } from "../primitives/UIComponents";

/**
 * Example component demonstrating the enhanced storage functionality.
 */
export function StorageExample() {
  const [logs, setLogs] = useState<string[]>([]);
  const [storageValue, setStorageValue] = useState<string>("");
  const [ttlValue, setTtlValue] = useState<number>(5000); // 5 seconds

  const storage = getStorage();

  const addLog = (message: string) => {
    setLogs((prev) => [...prev, `${new Date().toISOString()}: ${message}`]);
  };

  const handleSetValue = () => {
    const value = `stored-${Date.now()}`;
    storage.setItem("test-value", value, ttlValue);
    setStorageValue(value);
    addLog(`✅ Set value: "${value}" with TTL: ${ttlValue}ms`);
  };

  const handleGetValue = () => {
    const value = storage.getItem("test-value");
    if (value) {
      setStorageValue(value);
      addLog(`✅ Retrieved value: "${value}"`);
    } else {
      addLog(`❌ No value found (may have expired)`);
      setStorageValue("");
    }
  };

  const handleRemoveValue = () => {
    storage.removeItem("test-value");
    setStorageValue("");
    addLog(`🗑️ Removed value`);
  };

  const handleClearAll = () => {
    storage.clear();
    setStorageValue("");
    addLog(`🧹 Cleared all storage`);
  };

  const handleGetMetadata = () => {
    const metadata = storage.getItemWithMetadata("test-value");
    if (metadata) {
      addLog(`📊 Metadata: ${JSON.stringify(metadata, null, 2)}`);
    } else {
      addLog(`❌ No metadata found`);
    }
  };

  const handleListKeys = () => {
    const keys = storage.keys();
    addLog(`📋 Storage keys: ${keys.join(", ") || "none"}`);
  };

  const handleGetSize = () => {
    const size = storage.size();
    addLog(`📏 Storage size: ${size} items`);
  };

  // Auto-refresh value display
  useEffect(() => {
    const interval = setInterval(() => {
      const value = storage.getItem("test-value");
      if (value !== storageValue) {
        setStorageValue(value || "");
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [storageValue, storage]);

  return (
    <div style={{ padding: "20px", border: "1px solid #ccc", margin: "10px" }}>
      <h3>Enhanced Storage Example</h3>
      <p style={{ fontSize: "14px", color: "#666", marginBottom: "20px" }}>
        Demonstrates enhanced storage with TTL, metadata, and multiple backends.
      </p>

      <div style={{ marginBottom: "20px" }}>
        <h4>Storage Operations</h4>
        <div
          style={{
            display: "flex",
            gap: "10px",
            flexWrap: "wrap",
            marginBottom: "10px",
          }}
        >
          <input
            type="number"
            value={ttlValue}
            onChange={(e) => setTtlValue(parseInt(e.target.value) || 5000)}
            placeholder="TTL (ms)"
            style={{
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
              width: "120px",
            }}
          />
          <Button onClick={handleSetValue}>Set Value</Button>
          <Button onClick={handleGetValue}>Get Value</Button>
          <Button onClick={handleRemoveValue}>Remove Value</Button>
        </div>

        <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
          <Button
            onClick={handleGetMetadata}
            style={{ backgroundColor: "#6c757d", color: "white" }}
          >
            Get Metadata
          </Button>
          <Button
            onClick={handleListKeys}
            style={{ backgroundColor: "#6c757d", color: "white" }}
          >
            List Keys
          </Button>
          <Button
            onClick={handleGetSize}
            style={{ backgroundColor: "#6c757d", color: "white" }}
          >
            Get Size
          </Button>
          <Button
            onClick={handleClearAll}
            style={{ backgroundColor: "#dc3545", color: "white" }}
          >
            Clear All
          </Button>
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Current Value</h4>
        <div
          style={{
            padding: "10px",
            backgroundColor: "#f5f5f5",
            borderRadius: "4px",
            fontFamily: "monospace",
          }}
        >
          {storageValue || "(empty)"}
        </div>
      </div>

      <div>
        <h4>Operation Logs</h4>
        <div
          style={{
            height: "200px",
            overflow: "auto",
            border: "1px solid #ddd",
            padding: "10px",
            backgroundColor: "#f9f9f9",
            fontFamily: "monospace",
            fontSize: "12px",
          }}
        >
          {logs.map((log, index) => (
            <div key={index} style={{ marginBottom: "2px" }}>
              {log}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
