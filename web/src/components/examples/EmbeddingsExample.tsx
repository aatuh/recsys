/**
 * Example component demonstrating embeddings functionality.
 * Shows local/remote provider selection, chunking, and batch processing.
 */

import React, { useState, useEffect } from "react";
import { useEmbeddings, useBatchEmbeddings } from "../../hooks/useEmbeddings";
import { useFeatureFlags } from "../../contexts/FeatureFlagsContext";
import { ChunkingUtils } from "../../embeddings";
import { Button } from "../primitives/UIComponents";

export function EmbeddingsExample() {
  const [logs, setLogs] = useState<string[]>([]);
  const [inputText, setInputText] = useState(
    "Hello, world! This is a test of the embeddings system."
  );
  const [embedding, setEmbedding] = useState<number[] | null>(null);
  const [similarity, setSimilarity] = useState<number | null>(null);
  const [processingTime, setProcessingTime] = useState<number | null>(null);
  const [chunkedTexts, setChunkedTexts] = useState<string[]>([]);

  const { isEnabled } = useFeatureFlags();
  const useRemoteEmbeddings = isEnabled("useRemoteEmbeddings");

  const { provider, isReady, isLoading, error, embed, initialize, dispose } =
    useEmbeddings({
      config: {
        model: "Xenova/all-MiniLM-L6-v2",
        dimension: 384,
        maxBatchSize: 16,
        maxTextLength: 512,
      },
    });

  const { embedBatch } = useBatchEmbeddings({
    maxBatchSize: 8,
    chunking: true,
  });

  const addLog = (message: string) => {
    setLogs((prev) => [...prev, `${new Date().toISOString()}: ${message}`]);
  };

  useEffect(() => {
    if (provider) {
      addLog(`✅ Provider initialized: ${provider.constructor.name}`);
      addLog(`📊 Model: ${provider.getModelName()}`);
      addLog(`📏 Dimension: ${provider.getDimension()}`);
    }
  }, [provider]);

  useEffect(() => {
    if (error) {
      addLog(`❌ Error: ${error.message}`);
    }
  }, [error]);

  const handleInitialize = async () => {
    try {
      addLog("🔄 Initializing embeddings provider...");
      await initialize();
    } catch (error) {
      addLog(
        `❌ Initialization failed: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  };

  const handleDispose = async () => {
    try {
      addLog("🗑️ Disposing embeddings provider...");
      await dispose();
      setEmbedding(null);
      setSimilarity(null);
      setProcessingTime(null);
    } catch (error) {
      addLog(
        `❌ Disposal failed: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  };

  const handleEmbedText = async () => {
    if (!isReady) {
      addLog("❌ Provider not ready");
      return;
    }

    try {
      addLog(`🔄 Embedding text: "${inputText.substring(0, 50)}..."`);
      const startTime = Date.now();

      const embeddings = await embed([inputText]);
      const processingTime = Date.now() - startTime;

      if (embeddings[0]) {
        setEmbedding(embeddings[0]);
        setProcessingTime(processingTime);

        addLog(`✅ Embedding generated (${processingTime}ms)`);
        addLog(`📊 Dimension: ${embeddings[0].length}`);
        addLog(
          `📈 First 5 values: [${embeddings[0]
            .slice(0, 5)
            .map((v) => v.toFixed(4))
            .join(", ")}...]`
        );
      }
    } catch (error) {
      addLog(
        `❌ Embedding failed: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  };

  const handleBatchEmbed = async () => {
    if (!isReady) {
      addLog("❌ Provider not ready");
      return;
    }

    try {
      const texts = [
        "The quick brown fox jumps over the lazy dog.",
        "Machine learning is a subset of artificial intelligence.",
        "Natural language processing helps computers understand human language.",
        "Embeddings represent text as dense vectors in high-dimensional space.",
      ];

      addLog(`🔄 Batch embedding ${texts.length} texts...`);
      const startTime = Date.now();

      const embeddings = await embedBatch(texts);
      const processingTime = Date.now() - startTime;

      addLog(`✅ Batch embeddings generated (${processingTime}ms)`);
      addLog(`📊 Generated ${embeddings.length} embeddings`);

      // Calculate similarity between first two embeddings
      if (embeddings.length >= 2 && embeddings[0] && embeddings[1]) {
        const similarity = cosineSimilarity(embeddings[0], embeddings[1]);
        setSimilarity(similarity);
        addLog(`📈 Similarity between texts 1 & 2: ${similarity.toFixed(4)}`);
      }
    } catch (error) {
      addLog(
        `❌ Batch embedding failed: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  };

  const handleChunkText = () => {
    const chunkingOptions = ChunkingUtils.getRecommendedChunkingOptions(
      useRemoteEmbeddings ? "remote" : "local"
    );

    const chunks = ChunkingUtils.chunkText(inputText, chunkingOptions);
    setChunkedTexts(chunks);

    addLog(`✂️ Text chunked into ${chunks.length} pieces`);
    addLog(`📏 Chunk sizes: ${chunks.map((c) => c.length).join(", ")}`);
  };

  const handleEmbedChunks = async () => {
    if (!isReady || chunkedTexts.length === 0) {
      addLog("❌ Provider not ready or no chunks available");
      return;
    }

    try {
      addLog(`🔄 Embedding ${chunkedTexts.length} chunks...`);
      const startTime = Date.now();

      const embeddings = await embedBatch(chunkedTexts);
      const processingTime = Date.now() - startTime;

      addLog(`✅ Chunk embeddings generated (${processingTime}ms)`);
      addLog(`📊 Generated ${embeddings.length} chunk embeddings`);
    } catch (error) {
      addLog(
        `❌ Chunk embedding failed: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  };

  const handleClearLogs = () => {
    setLogs([]);
  };

  const cosineSimilarity = (a: number[], b: number[]): number => {
    if (!a || !b || a.length !== b.length) return 0;

    let dotProduct = 0;
    let normA = 0;
    let normB = 0;

    for (let i = 0; i < a.length; i++) {
      const aVal = a[i] || 0;
      const bVal = b[i] || 0;
      dotProduct += aVal * bVal;
      normA += aVal * aVal;
      normB += bVal * bVal;
    }

    return dotProduct / (Math.sqrt(normA) * Math.sqrt(normB));
  };

  return (
    <div style={{ padding: "20px", border: "1px solid #ccc", margin: "10px" }}>
      <h3>Embeddings Example</h3>
      <p style={{ fontSize: "14px", color: "#666", marginBottom: "20px" }}>
        Demonstrates local/remote embeddings providers, chunking, and batch
        processing. Current mode:{" "}
        <strong>{useRemoteEmbeddings ? "Remote" : "Local"}</strong>
      </p>

      <div style={{ marginBottom: "20px" }}>
        <h4>Provider Management</h4>
        <div
          style={{
            display: "flex",
            gap: "10px",
            flexWrap: "wrap",
            marginBottom: "10px",
          }}
        >
          <Button onClick={handleInitialize} disabled={isLoading || isReady}>
            {isLoading ? "Initializing..." : "Initialize Provider"}
          </Button>
          <Button onClick={handleDispose} disabled={!isReady}>
            Dispose Provider
          </Button>
        </div>

        <div style={{ fontSize: "14px", color: "#666" }}>
          Status:{" "}
          {isLoading ? "Loading..." : isReady ? "Ready" : "Not initialized"}
          {provider && (
            <>
              <br />
              Provider: {provider.constructor.name}
              <br />
              Model: {provider.getModelName()}
              <br />
              Dimension: {provider.getDimension()}
            </>
          )}
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Text Embedding</h4>
        <div style={{ marginBottom: "10px" }}>
          <textarea
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            placeholder="Enter text to embed..."
            style={{
              width: "100%",
              height: "80px",
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
              resize: "vertical",
            }}
          />
        </div>

        <div
          style={{
            display: "flex",
            gap: "10px",
            flexWrap: "wrap",
            marginBottom: "10px",
          }}
        >
          <Button onClick={handleEmbedText} disabled={!isReady}>
            Embed Text
          </Button>
          <Button onClick={handleChunkText} disabled={!inputText.trim()}>
            Chunk Text
          </Button>
          <Button
            onClick={handleEmbedChunks}
            disabled={!isReady || chunkedTexts.length === 0}
          >
            Embed Chunks
          </Button>
        </div>

        {embedding && (
          <div style={{ marginTop: "10px", fontSize: "12px", color: "#666" }}>
            <strong>Embedding:</strong> [
            {embedding
              .slice(0, 5)
              .map((v) => v.toFixed(4))
              .join(", ")}
            ...]
            {processingTime && (
              <>
                <br />
                <strong>Processing Time:</strong> {processingTime}ms
              </>
            )}
          </div>
        )}

        {similarity !== null && (
          <div style={{ marginTop: "10px", fontSize: "12px", color: "#666" }}>
            <strong>Similarity:</strong> {similarity.toFixed(4)}
          </div>
        )}

        {chunkedTexts.length > 0 && (
          <div style={{ marginTop: "10px" }}>
            <strong>Chunks ({chunkedTexts.length}):</strong>
            {chunkedTexts.map((chunk, index) => (
              <div
                key={index}
                style={{
                  margin: "5px 0",
                  padding: "5px",
                  backgroundColor: "#f5f5f5",
                  borderRadius: "4px",
                  fontSize: "12px",
                }}
              >
                {index + 1}. {chunk.substring(0, 100)}
                {chunk.length > 100 ? "..." : ""}
              </div>
            ))}
          </div>
        )}
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Batch Processing</h4>
        <Button onClick={handleBatchEmbed} disabled={!isReady}>
          Batch Embed Sample Texts
        </Button>
      </div>

      <div>
        <h4>Activity Logs</h4>
        <div style={{ display: "flex", gap: "10px", marginBottom: "10px" }}>
          <Button
            onClick={handleClearLogs}
            style={{ backgroundColor: "#6c757d", color: "white" }}
          >
            Clear Logs
          </Button>
        </div>

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
