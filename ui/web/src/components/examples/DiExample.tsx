import React, { useState, useEffect } from "react";
import { getLogger, getStorage, getEmbeddings } from "../../di";
import { Button } from "../primitives/UIComponents";

/**
 * Example component demonstrating the new dependency injection patterns.
 * This shows how to use the DI container in React components.
 */
export function DiExample() {
  const [logs, setLogs] = useState<string[]>([]);
  const [embedding, setEmbedding] = useState<number[] | null>(null);
  const [storageValue, setStorageValue] = useState<string>("");

  // Get services from DI container
  const logger = getLogger();
  const storage = getStorage();
  const embeddings = getEmbeddings();

  useEffect(() => {
    // Example of using child logger with component context
    const componentLogger = logger.child({ component: "DiExample" });

    componentLogger.info("Component mounted");

    // Load from storage
    const stored = storage.getItem("di-example-value");
    setStorageValue(stored || "");

    return () => {
      componentLogger.info("Component unmounted");
    };
  }, [logger, storage]);

  const handleLogExample = () => {
    const componentLogger = logger.child({ action: "log_example" });

    componentLogger.debug("Debug message", { timestamp: Date.now() });
    componentLogger.info("Info message", { user: "demo" });
    componentLogger.warn("Warning message", { level: "medium" });
    componentLogger.error("Error message", { code: "DEMO_ERROR" });

    setLogs((prev) => [...prev, "Logs sent to console/analytics"]);
  };

  const handleStorageExample = () => {
    const value = `stored-${Date.now()}`;
    storage.setItem("di-example-value", value);
    setStorageValue(value);

    logger.info("Storage updated", { key: "di-example-value", value });
  };

  const handleEmbeddingExample = async () => {
    try {
      logger.info("Computing embedding", { text: "Hello, world!" });

      const result = await embeddings.embedText("Hello, world!");
      setEmbedding(result);

      logger.info("Embedding computed", {
        dimension: result.length,
        firstFew: result.slice(0, 5),
      });
    } catch (error) {
      logger.error("Embedding failed", {
        error: error instanceof Error ? error.message : String(error),
      });
    }
  };

  const handleItemToTextExample = () => {
    const item = {
      item_id: "demo-item-123",
      tags: ["electronics", "gadget"],
      price: 99.99,
      props: { brand: "DemoBrand", category: "tech" },
    };

    const text = embeddings.itemToText(item);
    logger.info("Item converted to text", { item, text });
    setLogs((prev) => [...prev, `Item text: "${text}"`]);
  };

  return (
    <div style={{ padding: "20px", border: "1px solid #ccc", margin: "10px" }}>
      <h3>DI Container Example</h3>

      <div style={{ marginBottom: "20px" }}>
        <h4>Logger Examples</h4>
        <Button onClick={handleLogExample}>Send Test Logs</Button>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Storage Examples</h4>
        <Button onClick={handleStorageExample}>Store Value</Button>
        <p>Current value: {storageValue}</p>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Embeddings Examples</h4>
        <Button onClick={handleEmbeddingExample}>Compute Embedding</Button>
        <Button onClick={handleItemToTextExample}>Convert Item to Text</Button>
        {embedding && <p>Embedding dimension: {embedding.length}</p>}
      </div>

      <div>
        <h4>Logs</h4>
        {logs.map((log, index) => (
          <div key={index} style={{ fontSize: "12px", color: "#666" }}>
            {log}
          </div>
        ))}
      </div>
    </div>
  );
}
