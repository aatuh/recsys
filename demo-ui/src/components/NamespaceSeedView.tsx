import React, { useRef, useState } from "react";
import { NamespaceSection, SeedDataSection } from "./";
import { useViewState } from "../contexts/ViewStateContext";

interface NamespaceSeedViewProps {
  namespace: string;
  setNamespace: (value: string) => void;
  apiBase: string;
  setGeneratedUsers: (users: string[]) => void;
  setGeneratedItems: (items: string[]) => void;
}

export function NamespaceSeedView({
  namespace,
  setNamespace,
  apiBase,
  setGeneratedUsers: setGlobalGeneratedUsers,
  setGeneratedItems: setGlobalGeneratedItems,
}: NamespaceSeedViewProps) {
  const { namespaceSeed, setNamespaceSeed } = useViewState();

  /* Local cache for generated ids to ease testing */
  const generatedUsersRef = useRef<string[]>([]);
  const generatedItemsRef = useRef<string[]>([]);

  /* State for generated users and items to trigger re-renders */
  const [generatedUsers, setGeneratedUsers] = useState<string[]>([]);
  const [generatedItems, setGeneratedItems] = useState<string[]>([]);

  /* Static data that doesn't need to be preserved */
  const [brands] = useState([
    "alfa",
    "bravo",
    "charlie",
    "delta",
    "echo",
    "foxtrot",
  ]);
  const [tags] = useState([
    "action",
    "indie",
    "rpg",
    "strategy",
    "sim",
    "puzzle",
  ]);

  /* User traits update handler */
  const handleUpdateUser = async (
    userId: string,
    traits: Record<string, any>
  ) => {
    // In a real application, this would call the API to update user traits
    // For now, we'll just log the update
    console.log(`Updating user ${userId} with traits:`, traits);
    // TODO: Implement actual API call to update user traits
    return Promise.resolve();
  };

  /* Item update handler */
  const handleUpdateItem = async (
    itemId: string,
    updates: Record<string, any>
  ) => {
    // In a real application, this would call the API to update item data
    // For now, we'll just log the update
    console.log(`Updating item ${itemId} with data:`, updates);
    // TODO: Implement actual API call to update item data
    return Promise.resolve();
  };

  return (
    <div style={{ padding: 16, fontFamily: "system-ui, sans-serif" }}>
      <p style={{ color: "#444", marginBottom: 24 }}>
        Configure namespace and generate synthetic data for testing. Set up
        event types, user traits, and item configurations to create realistic
        test data.
      </p>

      <NamespaceSection
        namespace={namespace}
        setNamespace={setNamespace}
        apiBase={apiBase}
      />

      <SeedDataSection
        userCount={namespaceSeed.userCount}
        setUserCount={(value) =>
          setNamespaceSeed((prev) => ({ ...prev, userCount: value }))
        }
        userStartIndex={namespaceSeed.userStartIndex}
        setUserStartIndex={(value) =>
          setNamespaceSeed((prev) => ({ ...prev, userStartIndex: value }))
        }
        itemCount={namespaceSeed.itemCount}
        setItemCount={(value) =>
          setNamespaceSeed((prev) => ({ ...prev, itemCount: value }))
        }
        minEventsPerUser={namespaceSeed.minEventsPerUser}
        setMinEventsPerUser={(value) =>
          setNamespaceSeed((prev) => ({ ...prev, minEventsPerUser: value }))
        }
        maxEventsPerUser={namespaceSeed.maxEventsPerUser}
        setMaxEventsPerUser={(value) =>
          setNamespaceSeed((prev) => ({ ...prev, maxEventsPerUser: value }))
        }
        eventTypes={namespaceSeed.eventTypes}
        setEventTypes={(value) =>
          setNamespaceSeed((prev) => ({ ...prev, eventTypes: value }))
        }
        namespace={namespace}
        brands={brands}
        tags={tags}
        log={namespaceSeed.log}
        setLog={(value) =>
          setNamespaceSeed((prev) => ({
            ...prev,
            log: typeof value === "function" ? value(prev.log) : value,
          }))
        }
        setGeneratedUsers={(users) => {
          generatedUsersRef.current = users;
          setGeneratedUsers(users);
          setGlobalGeneratedUsers(users);
        }}
        setGeneratedItems={(items) => {
          generatedItemsRef.current = items;
          setGeneratedItems(items);
          setGlobalGeneratedItems(items);
        }}
        traitConfigs={namespaceSeed.traitConfigs}
        setTraitConfigs={(value) =>
          setNamespaceSeed((prev) => ({ ...prev, traitConfigs: value }))
        }
        itemConfigs={namespaceSeed.itemConfigs}
        setItemConfigs={(value) =>
          setNamespaceSeed((prev) => ({ ...prev, itemConfigs: value }))
        }
        priceRanges={namespaceSeed.priceRanges}
        setPriceRanges={(value) =>
          setNamespaceSeed((prev) => ({ ...prev, priceRanges: value }))
        }
        generatedUsers={generatedUsers}
        generatedItems={generatedItems}
        onUpdateUser={handleUpdateUser}
        onUpdateItem={handleUpdateItem}
      />
    </div>
  );
}
