import React from "react";
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { RecommendationsSection } from "./RecommendationsSection";
import { ToastProvider } from "../../ui/Toast";

describe("RecommendationsSection", () => {
  it("renders button and empty state", () => {
    render(
      <ToastProvider>
        <RecommendationsSection
          recUserId=""
          setRecUserId={() => {}}
          k={10}
          setK={() => {}}
          blend={{ pop: 1, cooc: 0.5, als: 0 }}
          setBlend={() => {}}
          namespace="ns"
          exampleUser="user-1"
          recOut={null}
          setRecOut={() => {}}
          recLoading={false}
          setRecLoading={() => {}}
          overrides={null}
          recResponse={null}
        />
      </ToastProvider>
    );

    expect(
      screen.getByRole("button", { name: /get recommendations/i })
    ).toBeInTheDocument();
  });
});
