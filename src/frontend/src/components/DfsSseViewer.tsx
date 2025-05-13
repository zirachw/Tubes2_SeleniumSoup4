"use client";

// components/DfsSseViewer.tsx
import { useEffect, useRef, useState } from "react";

interface Update {
  Stage: string;
  ElementName: string;
  Tier: number;
  RecipeIndex: number;
  Info: string;
  ParentID: number;
  LeftID: number;
  RightID: number;
  LeftLabel: string;
  RightLabel: string;
}

export function DfsSseViewer() {
  const [displayed, setDisplayed] = useState<Update[]>([]);
  const queueRef = useRef<Update[]>([]);
  const evtSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    // 1) open SSE
    const es = new EventSource("http://localhost:8080/stream");
    evtSourceRef.current = es;

    es.onmessage = (e) => {
      try {
        console.log("SSE data received:", e.data);
        const updates: Update[] = JSON.parse(e.data);
        queueRef.current.push(...updates);
      } catch (err) {
        console.error("Failed to parse SSE data", err);
      }
    };

    es.onerror = () => {
      console.warn("SSE connection error/closed");
      es.close();
    };

    return () => {
      es.close();
    };
  }, []);

  useEffect(() => {
    // 2) every 500ms, pop one update and add to displayed
    const interval = setInterval(() => {
      const q = queueRef.current;
      if (q.length === 0) {
        return;
      }
      const next = q.shift()!; // non-null because length>0
      setDisplayed((prev) => [...prev, next]);
    }, 500);

    return () => clearInterval(interval);
  }, []);

  return (
    <div>
      <h2>DFS Updates</h2>
      <ul>
        {displayed.map((u, i) => (
          <li key={i}>
            <strong>{u.Stage}</strong> — {u.ElementName} (tier {u.Tier}) —
            {u.Stage == "startRecipe"
              ? u.ParentID +
                " " +
                u.LeftID +
                " " +
                u.RightID +
                " " +
                u.LeftLabel +
                " " +
                u.RightLabel
              : ""}
            {u.Info && <em>: {u.Info}</em>}
          </li>
        ))}
      </ul>
    </div>
  );
}
