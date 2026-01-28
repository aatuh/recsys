import { NextRequest, NextResponse } from "next/server";
import { syncAllItemsToRecsys } from "@/server/services/itemSync";
import { syncAllUsersToRecsys } from "@/server/services/userSync";
import {
  upsertEventTypeConfig,
  deleteItems,
  deleteUsers,
  deleteAllItemsInNamespace,
  deleteAllUsersInNamespace,
  deleteAllEventsInNamespace,
} from "@/server/services/recsys";
import { prisma } from "@/server/db/client";

// Helper function to nuke data directly (avoiding HTTP requests to self)
async function nukeData(tables: string[]) {
  const requested = new Set(tables);

  // Expand dependencies so FK constraints won't fail even if caller passes a subset
  if (requested.has("user")) {
    requested.add("events");
    requested.add("orderItem");
    requested.add("order");
    requested.add("cartItem");
    requested.add("cart");
  }
  if (requested.has("product")) {
    requested.add("orderItem");
    requested.add("cartItem");
  }
  if (requested.has("order")) {
    requested.add("orderItem");
  }
  if (requested.has("cart")) {
    requested.add("cartItem");
  }

  // Collect IDs for Recsys sync before deletion
  const userIds: string[] = [];
  const productIds: string[] = [];

  if (requested.has("user")) {
    const users = await prisma.user.findMany({ select: { id: true } });
    userIds.push(...users.map((u: { id: string }) => u.id));
  }

  if (requested.has("product")) {
    const products = await prisma.product.findMany({ select: { id: true } });
    productIds.push(...products.map((p: { id: string }) => p.id));
  }

  // Enforce safe deletion order (children first)
  const order = [
    "events",
    "orderItem",
    "order",
    "cartItem",
    "cart",
    "product",
    "user",
  ];

  for (const t of order) {
    if (!requested.has(t)) continue;
    if (t === "events") await prisma.event.deleteMany({});
    else if (t === "orderItem") await prisma.orderItem.deleteMany({});
    else if (t === "order") await prisma.order.deleteMany({});
    else if (t === "cartItem") await prisma.cartItem.deleteMany({});
    else if (t === "cart") await prisma.cart.deleteMany({});
    else if (t === "product") await prisma.product.deleteMany({});
    else if (t === "user") await prisma.user.deleteMany({});
  }

  // Sync deletions with Recsys
  // If the caller requested a full product wipe, delete all items in namespace
  if (requested.has("product")) {
    try {
      await deleteAllItemsInNamespace();
    } catch (error) {
      console.error("Failed to delete all items in recsys:", error);
    }
  } else if (productIds.length > 0) {
    try {
      await deleteItems(productIds);
    } catch (error) {
      console.error("Failed to sync product deletions to recsys:", error);
    }
  }

  if (requested.has("user")) {
    try {
      // Remove all user events and users from recsys namespace
      await deleteAllEventsInNamespace();
      await deleteAllUsersInNamespace();
    } catch (error) {
      console.error("Failed to delete all users/events in recsys:", error);
    }
  } else if (userIds.length > 0) {
    try {
      await deleteUsers(userIds);
    } catch (error) {
      console.error("Failed to sync user deletions to recsys:", error);
    }
  }

  return { status: "nuked", tables: Array.from(requested) };
}

// Helper function to seed products directly
async function seedProducts(count: number) {
  const brands = ["Acme", "Globex", "Umbrella", "Stark", "Wayne"];
  const categories = ["Shoes", "Electronics", "Home", "Books", "Toys"];
  const adjectives = [
    "Advanced",
    "Bold",
    "Brilliant",
    "Chic",
    "Classic",
    "Classy",
    "Compact",
    "Crystal",
    "Deluxe",
    "Dynamic",
    "Eco-Friendly",
    "Eco",
    "Efficient",
    "Elegant",
    "Energy-Saving",
    "Exclusive",
    "Exquisite",
    "Fashionable",
    "Glow",
    "Hyper",
    "Innovative",
    "Intelligent",
    "Luxury",
    "Minimalist",
    "Modern",
    "Modular",
    "Multi-Functional",
    "NextGen",
    "Portable",
    "Powerful",
    "Premium",
    "Pro",
    "Radiant",
    "Refined",
    "Retro",
    "Robust",
    "Shiny",
    "Sleek",
    "Smart",
    "Sophisticated",
    "Sparkling",
    "State-of-the-Art",
    "Sturdy",
    "Stylish",
    "Supreme",
    "Trendy",
    "Ultimate",
    "Ultra",
    "Velvet",
    "Versatile",
    "Vivid",
    "Wireless",
  ];
  const nouns = [
    "Adapter",
    "Air Purifier",
    "Amplifier",
    "Backpack",
    "Battery Pack",
    "Board",
    "Bracelet",
    "Cable",
    "Camera Lens",
    "Camera",
    "Charger",
    "Clock",
    "Console",
    "Controller",
    "Converter",
    "Cooker",
    "Desktop",
    "Disk",
    "Dongle",
    "Drive",
    "Drone",
    "DSP",
    "Earbud",
    "Equalizer",
    "Fitness Tracker",
    "Flashlight",
    "Gaming Chair",
    "Graphics Card",
    "Headphone",
    "Holster",
    "Hub",
    "Jacket",
    "Keyboard Case",
    "Keyboard",
    "Lamp",
    "Laptop",
    "Lens",
    "Microphone",
    "MicroSD",
    "Modem",
    "Monitor",
    "Mouse",
    "Mug",
    "Network Switch",
    "Notebook",
    "Pen Tablet",
    "Pen",
    "Phone",
    "Preamp",
    "Printer",
    "Processor",
    "Projector Screen",
    "Projector",
    "Receiver",
    "Router",
    "Routerboard",
    "Scanner",
    "Server",
    "Smart Ring",
    "Smartwatch",
    "Sneaker",
    "Speaker",
    "Stabilizer",
    "Steamer",
    "Subwoofer",
    "Switch",
    "Tablet",
    "Thermostat",
    "Tripod",
    "VR Headset",
    "Watch",
    "Webcam",
  ];

  const data = Array.from({ length: count }).map((_, i) => ({
    sku: `RND-${Date.now()}-${i}`,
    name: `${adjectives[i % adjectives.length]} ${nouns[i % nouns.length]}`,
    description: "Random product",
    price: Math.round((10 + Math.random() * 190) * 100) / 100,
    currency: "USD",
    brand: brands[i % brands.length],
    category: categories[i % categories.length],
    imageUrl: "", // use local placeholder for reliability
    stockCount: Math.floor(Math.random() * 50) + 1,
    tagsCsv: categories[i % categories.length].toLowerCase(),
  }));

  await prisma.product.createMany({ data });

  // Get the created products to sync to recsys
  const createdProducts = await prisma.product.findMany({
    where: {
      sku: {
        in: data.map((d) => d.sku),
      },
    },
    orderBy: { createdAt: "desc" },
    take: count,
  });

  // Upsert to recsys with actual database IDs
  const { upsertItems } = await import("@/server/services/recsys");
  const { buildItemContract } = await import("@/lib/contracts/item");
  void upsertItems(
    createdProducts.map(
      (product: {
        id: string;
        name: string;
        sku: string;
        price: number;
        currency: string;
        brand?: string | null;
        category?: string | null;
        description?: string | null;
        imageUrl?: string | null;
        stockCount: number;
        tagsCsv?: string | null;
      }) => buildItemContract(product)
    )
  ).catch((error) => {
    console.error("Failed to sync products to recsys:", error);
  });

  return { inserted: data.length };
}

// Helper function to seed users directly
async function seedUsers(count: number) {
  const firstNames = [
    "Alex",
    "Jordan",
    "Taylor",
    "Casey",
    "Morgan",
    "Riley",
    "Avery",
    "Quinn",
    "Blake",
    "Cameron",
    "Drew",
    "Emery",
    "Finley",
    "Hayden",
    "Jamie",
    "Kendall",
    "Logan",
    "Parker",
    "Reese",
    "Sage",
    "Skyler",
    "Tatum",
    "River",
    "Phoenix",
    "Rowan",
    "Sage",
    "Aspen",
    "Cedar",
    "Oakley",
    "Indigo",
  ];
  const lastNames = [
    "Smith",
    "Johnson",
    "Williams",
    "Brown",
    "Jones",
    "Garcia",
    "Miller",
    "Davis",
    "Rodriguez",
    "Martinez",
    "Hernandez",
    "Lopez",
    "Gonzalez",
    "Wilson",
    "Anderson",
    "Thomas",
    "Taylor",
    "Moore",
    "Jackson",
    "Martin",
    "Lee",
    "Perez",
    "Thompson",
    "White",
    "Harris",
    "Sanchez",
    "Clark",
    "Ramirez",
    "Lewis",
    "Robinson",
  ];

  const data = Array.from({ length: count }).map((_, i) => ({
    displayName: `${firstNames[i % firstNames.length]} ${
      lastNames[i % lastNames.length]
    }`,
    traitsText: JSON.stringify({
      age: Math.floor(Math.random() * 50) + 18,
      location: ["New York", "Los Angeles", "Chicago", "Houston", "Phoenix"][
        i % 5
      ],
      preferences: ["electronics", "books", "clothing", "home", "sports"][
        i % 5
      ],
    }),
  }));

  await prisma.user.createMany({ data });

  // Get the created users to sync to recsys
  const createdUsers = await prisma.user.findMany({
    where: {
      displayName: {
        in: data.map((d) => d.displayName),
      },
    },
    orderBy: { createdAt: "desc" },
    take: count,
  });

  // Upsert to recsys with actual database IDs
  const { upsertUsers } = await import("@/server/services/recsys");
  const { buildUserContract } = await import("@/lib/contracts/user");
  void upsertUsers(
    createdUsers.map(
      (user: { id: string; displayName: string; traitsText?: string | null }) =>
        buildUserContract(user)
    )
  ).catch((error) => {
    console.error("Failed to sync users to recsys:", error);
  });

  return { inserted: data.length };
}

export async function POST(req: NextRequest) {
  try {
    const { action, params } = await req.json();

    switch (action) {
      case "sync-all-items":
        const itemResult = await syncAllItemsToRecsys();
        return NextResponse.json({ status: "success", ...itemResult });

      case "sync-all-users":
        const userResult = await syncAllUsersToRecsys();
        return NextResponse.json({ status: "success", ...userResult });

      case "flush-events":
        // Flush all events to recsys
        const events = await prisma.event.findMany({
          where: { recsysStatus: "pending" },
          take: 1000, // Process in batches
        });

        if (events.length > 0) {
          // Import the event sync service
          const { forwardEventsBatch } = await import(
            "@/server/services/recsys"
          );
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          await forwardEventsBatch(
            events.map(
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              (e: any) => ({
                user_id: e.userId || "anonymous",
                item_id: e.productId || "",
                type: e.type,
                value: e.value,
                ts: e.createdAt.toISOString(),
                meta: e.meta ? JSON.parse(e.meta) : {},
              })
            )
          );

          // Mark events as synced
          await prisma.event.updateMany({
            where: { id: { in: events.map((e: { id: string }) => e.id) } },
            data: { recsysStatus: "sent" },
          });
        }

        return NextResponse.json({
          status: "success",
          flushed: events.length,
          message: `Flushed ${events.length} events to recsys`,
        });

      case "retry-failed-events":
        // Retry failed events by finding events that failed to sync
        const failedEvents = await prisma.event.findMany({
          where: {
            recsysStatus: "failed",
            ts: { lt: new Date(Date.now() - 5 * 60 * 1000) }, // Older than 5 minutes
          },
          take: 1000,
        });

        if (failedEvents.length > 0) {
          const { forwardEventsBatch } = await import(
            "@/server/services/recsys"
          );
          try {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            await forwardEventsBatch(
              failedEvents.map(
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                (e: any) => ({
                  user_id: e.userId || "anonymous",
                  item_id: e.productId || "",
                  type: e.type,
                  value: e.value,
                  ts: e.createdAt.toISOString(),
                  meta: e.meta ? JSON.parse(e.meta) : {},
                })
              )
            );

            // Mark as synced on success
            await prisma.event.updateMany({
              where: {
                id: { in: failedEvents.map((e: { id: string }) => e.id) },
              },
              data: { recsysStatus: "sent" },
            });
          } catch (error) {
            console.error("Failed to retry events:", error);
          }
        }

        return NextResponse.json({
          status: "success",
          retried: failedEvents.length,
          message: `Retried ${failedEvents.length} failed events`,
        });

      case "seed-products":
        const productCount = params?.count || 50;
        const seedProductsResult = await seedProducts(productCount);
        return NextResponse.json({ status: "success", ...seedProductsResult });

      case "seed-users":
        const userCount = params?.count || 20;
        const seedUsersResult = await seedUsers(userCount);
        return NextResponse.json({ status: "success", ...seedUsersResult });

      case "init-event-types":
        const eventTypeResult = await upsertEventTypeConfig();
        return NextResponse.json({
          status: "success",
          result: eventTypeResult,
        });

      case "nuke-all":
        const nukeResult = await nukeData([
          "events",
          "orderItem",
          "order",
          "cartItem",
          "cart",
          "product",
          "user",
        ]);
        return NextResponse.json({ ...nukeResult });

      case "check-recsys-health":
        // Simple health check by trying to get recommendations
        try {
          const healthRes = await fetch(
            `${req.nextUrl.origin}/api/recommendations?userId=health-check&k=1`
          );
          const isHealthy = healthRes.ok;
          return NextResponse.json({
            status: "success",
            healthy: isHealthy,
            message: isHealthy
              ? "Recsys is responding"
              : "Recsys is not responding",
          });
        } catch {
          return NextResponse.json({
            status: "success",
            healthy: false,
            message: "Recsys health check failed",
          });
        }

      case "validate-data-integrity":
        // Check for orphaned records and data consistency
        const orphanedEvents = await prisma.event.count({
          where: {
            OR: [{ userId: null }, { productId: null }],
          },
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } as any);

        const orphanedCartItems = await prisma.cartItem.count({
          where: {
            OR: [{ cartId: null }, { productId: null }],
          },
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } as any);

        const orphanedOrderItems = await prisma.orderItem.count({
          where: {
            OR: [{ orderId: null }, { productId: null }],
          },
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } as any);

        return NextResponse.json({
          status: "success",
          integrity: {
            orphanedEvents,
            orphanedCartItems,
            orphanedOrderItems,
            isHealthy:
              orphanedEvents === 0 &&
              orphanedCartItems === 0 &&
              orphanedOrderItems === 0,
          },
        });

      default:
        return NextResponse.json({ error: "Invalid action" }, { status: 400 });
    }
  } catch (error) {
    console.error("Admin tools error:", error);
    return NextResponse.json({ error: "Admin action failed" }, { status: 500 });
  }
}
