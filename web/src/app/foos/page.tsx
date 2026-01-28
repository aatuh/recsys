import { SectionHeader } from "@api-boilerplate-core/ui";
import { FoosClient } from "@/components/foos-client";

export default function FoosPage() {
  return (
    <div className="space-y-6">
      <SectionHeader
        title="Foo API playground"
        description="Create, search, and delete Foo records using the generated client and domain adapters."
      />
      <FoosClient />
    </div>
  );
}
