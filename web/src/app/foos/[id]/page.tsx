// Commented: Kept for reference
// import Link from "next/link";
// import { notFound } from "next/navigation";
// import { createFooService } from "@foo/domain";
// import { createFooRepository } from "@foo/domain-adapters";
// import { Button, Card, SectionHeader } from "@api-boilerplate-core/ui";

// export default async function FooDetailPage({
//   params,
// }: {
//   params: { id: string };
// }) {
//   const service = createFooService(createFooRepository());
//   let foo;
//   try {
//     foo = await service.get(params.id);
//   } catch {
//     notFound();
//   }

//   return (
//     <div className="space-y-6">
//       <SectionHeader
//         title={foo.name}
//         description="Generated detail page using the domain service and adapters."
//         actions={
//           <Button href="/foos" variant="secondary" size="md">
//             Back to list
//           </Button>
//         }
//       />
//       <Card>
//         <dl className="grid gap-4 text-sm text-muted sm:grid-cols-2">
//           <div>
//             <dt className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-strong">
//               Foo ID
//             </dt>
//             <dd>{foo.id}</dd>
//           </div>
//           <div>
//             <dt className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-strong">
//               Org ID
//             </dt>
//             <dd>{foo.orgId}</dd>
//           </div>
//           <div>
//             <dt className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-strong">
//               Namespace
//             </dt>
//             <dd>{foo.namespace}</dd>
//           </div>
//           <div>
//             <dt className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-strong">
//               Updated
//             </dt>
//             <dd>{foo.updatedAt}</dd>
//           </div>
//         </dl>
//         <Link
//           className="mt-6 inline-flex text-sm font-semibold text-primary"
//           href="/docs/architecture"
//         >
//           See the mapping flow in docs
//         </Link>
//       </Card>
//     </div>
//   );
// }
