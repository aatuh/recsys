import React from "react";
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { DataTable, type Column } from "./DataTable";

type Row = { id: string; name: string };

describe("DataTable", () => {
  it("renders empty message", () => {
    const columns: Column<Row>[] = [
      { key: "id", title: "ID" },
      { key: "name", title: "Name" },
    ];
    render(
      <DataTable<Row> data={[]} columns={columns} emptyMessage="No rows" />
    );

    expect(screen.getByText(/no rows/i)).toBeInTheDocument();
  });
});
