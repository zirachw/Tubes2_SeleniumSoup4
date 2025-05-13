"use client";

import React, { useState, useCallback, useRef, useEffect } from "react";
import ELK from "elkjs/lib/elk.bundled.js";
import ReactFlow, {
  ReactFlowProvider,
  addEdge,
  applyEdgeChanges,
  applyNodeChanges,
  Connection,
  Edge,
  Node,
  Controls,
  Background,
  OnConnect,
  OnNodesChange,
  OnEdgesChange,
  ReactFlowInstance,
  MarkerType,
  MiniMap,
} from "reactflow";
import "reactflow/dist/style.css";
import { CoupleNode, SingleNode } from "@/components";
import { ElementsData } from "@/types";
import { BOX_HEIGHT, BOX_WIDTH, GAP, PADDING } from "./CoupleNode";

const edgeColorDefault = "#b1b1b7";
const edgeColorHighlight = "#4ade80";
const edgeWidthDefault = 2;
const edgeWidthHighlight = 4;

const elk = new ELK();

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

const getLayoutedElements = async (
  nodes: Node[],
  edges: Edge[],
  direction = "DOWN"
) => {
  const rightSideNodes = new Set<string>();
  edges.forEach((edge) => {
    if (parseInt(edge.sourceHandle ?? "") % 2 === 0) {
      // assume even handles are on the left lol
      rightSideNodes.add(edge.target);
    }
  });

  const graph = {
    id: "root",
    layoutOptions: {
      "elk.algorithm": "layered",
      "elk.direction": direction,
      "elk.spacing.nodeNode": "80", // more horizontal gap
      "elk.layered.spacing.nodeNodeBetweenLayers": "80", // more vertical gap
      "elk.layered.spacing.edgeNode": "40", // space edges around nodes
      "elk.layered.spacing.edgeEdge": "20", // space parallel edges
      "elk.layered.nodePlacement.strategy": "SIMPLE",
      // still bias toward straight edges
      // "elk.layered.nodePlacement.favorStraightEdges": "true",
    },
    children: nodes.map((n) => {
      const targetPorts = n.data.targetHandles.map((t: { id: string }) => ({
        id: t.id,

        // NOTES: it's important to let elk know on which side the port is
        properties: {
          side: "NORTH",
        },
      }));

      const sourcePorts = n.data.sourceHandles.map((s: { id: string }) => ({
        id: s.id,
        properties: {
          side: "SOUTH",
        },
      }));

      return {
        id: n.id,
        width: n.width ?? 150,
        height: n.height ?? 50,
        // ️NOTES: we need to tell elk that the ports are fixed, in order to reduce edge crossings
        properties: {
          "org.eclipse.elk.portConstraints": "FIXED_ORDER",
        },
        // we are also passing the id, so we can also handle edges without a sourceHandle or targetHandle option
        ports: [{ id: n.id }, ...targetPorts, ...sourcePorts],
      };
    }),
    edges: edges.map((e) => ({
      id: e.id,
      sources: [e.sourceHandle || e.source],
      targets: [e.targetHandle || e.target],
    })),
  };

  const { children: layoutedChildren = [] } = await elk.layout(graph);
  const layoutedNodes = layoutedChildren.map((n) => {
    const yOffset = rightSideNodes.has(n.id) ? 30 : 0; // add extra vertical offset for right-side nodes

    return {
      id: n.id,
      position: { x: n.x || 0, y: (n.y || 0) + yOffset },
      data: nodes.find((node) => node.id === n.id)?.data || { label: n.id },
      width: n.width,
      height: n.height,
    };
  });

  return { nodes: layoutedNodes, edges };
};

const animateLayout = (
  startNodes: Node[],
  endNodes: Node[],
  setNodes: React.Dispatch<React.SetStateAction<Node[]>>,
  duration = 300
) => {
  const startTime = performance.now();
  const startPos = Object.fromEntries(
    startNodes.map((n) => [n.id, { x: n.position.x, y: n.position.y }])
  );
  const endPos = Object.fromEntries(
    endNodes.map((n) => [n.id, { x: n.position.x, y: n.position.y }])
  );

  function frame(now: number) {
    const t = Math.min((now - startTime) / duration, 1);

    setNodes((nodes) =>
      nodes.map((node) => {
        const s = startPos[node.id];
        const e = endPos[node.id];
        if (!s || !e) return node;
        return {
          ...node,
          position: {
            x: s.x + (e.x - s.x) * t,
            y: s.y + (e.y - s.y) * t,
          },
        };
      })
    );

    if (t < 1) requestAnimationFrame(frame);
  }

  requestAnimationFrame(frame);
};

const nodeTypes = {
  couple: CoupleNode,
  single: SingleNode,
};

interface QueryParams {
  element: string | null;
  algorithm: string | null;
  multipleRecipes: boolean;
  liveUpdate: boolean;
  count: number | "all";
}

interface TreeViewerProps {
  elementsData: ElementsData;
  loading: boolean;
  queryParams: QueryParams;
  trigger: boolean;
  onFinish: () => void;
}

const TreeViewer: React.FC<TreeViewerProps> = ({
  elementsData,
  loading,
  queryParams,
  trigger,
  onFinish,
}) => {
  const [nodes, setNodes] = useState<Node[]>([]);
  const queueRef = useRef<Update[]>([]);
  const evtSourceRef = useRef<EventSource | null>(null);
  const [edges, setEdges] = useState<Edge[]>([]);
  const hasStartedRef = useRef(false);
  const hasRoot = useRef(false);

  const rootNode = useRef<Node | null>(null);

  const nodeCountRef = useRef(1);
  const makeId = () => `n${nodeCountRef.current++}`;
  // Keep React Flow instance
  const [rfInstance, setRfInstance] = useState<ReactFlowInstance | null>(null);

  // Layout with tween + smooth fitView
  const layoutFlow = useCallback(
    async (
      direction: "DOWN" | "UP" = "DOWN",
      sourceNodes?: Node[],
      sourceEdges?: Edge[]
    ) => {
      const baseNodes = sourceNodes || nodes;
      const baseEdges = sourceEdges || edges;
      const { nodes: layoutedNodes, edges: layoutedEdges } =
        await getLayoutedElements(baseNodes, baseEdges, direction);

      if (!document.hidden) {
        animateLayout(baseNodes, layoutedNodes, setNodes, 400);
      }
      setEdges(layoutedEdges);

      // smooth transition on fitView
      if (!document.hidden) {
        setTimeout(() => {
          if (rfInstance) {
            rfInstance.fitView({ padding: 0.2, duration: 400 });
          }
        }, 400);
      }
    },
    [nodes, edges, rfInstance]
  );
  // Add a node under its parent, then relayout + fitView
  const addNode = useCallback(
    (
      ParentID: number,
      LeftID: number,
      RightID: number,
      LeftLabel: string,
      RightLabel: string
    ) => {
      setNodes((oldNodes) => {
        const id = `node-${LeftID}`;
        const leftImageLink = elementsData![LeftLabel].imageLink;
        const rightImageLink = elementsData![RightLabel].imageLink;

        const parent = oldNodes.find(
          (n) =>
            n.id === "node-" + ParentID.toString() ||
            (n.id != "node-1" && n.id === "node-" + (ParentID - 1).toString())
        );
        if (!parent) {
          console.error("Parent node not found");
          return oldNodes; // Return the current state if no parent is found
        }

        const parentNodeId = parent?.id;
        const { x: px, y: py } = parent.position;
        const ph = parent.height ?? 50;

        const newNode: Node = {
          id,
          type: "couple",
          data: {
            LeftLabel,
            RightLabel,
            leftImageLink,
            rightImageLink,
            id,
            LeftID,
            RightID,
            targetHandles: [{ id: `${id}` }],
            sourceHandles: [{ id: `${LeftID}` }, { id: `${RightID}` }],
          },
          position: { x: px, y: py + ph + 50 },
          // must match containerStyle dimensions:
          width: 2 * BOX_WIDTH + GAP + 5 * PADDING,
          height: BOX_HEIGHT + 2 * PADDING,
        };

        setEdges((oldEdges) => {
          const sourceHandle = ParentID.toString();
          // handle source if parent is root
          const sourceHandleRoot = "parent-root";
          const sourceHandleParent =
            ParentID === 1 ? sourceHandleRoot : sourceHandle;

          const newEdge: Edge = {
            id: `e_${parentNodeId}_${makeId()}`,
            source: parentNodeId,
            sourceHandle: sourceHandleParent,
            target: id,
            targetHandle: `${id}`,
            type: "smoothstep",
            markerStart: {
              type: MarkerType.ArrowClosed,
              width: 10,
              height: 10,
            },
            style: { strokeWidth: 2 },
          };

          const updatedEdges = [...oldEdges, newEdge];
          // trigger layout with the brand-new arrays
          layoutFlow("DOWN", [...oldNodes, newNode], updatedEdges);
          return updatedEdges;
        });

        return [...oldNodes, newNode];
      });
    },
    [elementsData, layoutFlow]
  );
  const liveUpdate = useCallback(() => {
    if (loading || hasStartedRef.current) {
      console.log("Already started live update or loading");
      return;
    }

    hasStartedRef.current = true;

    // 1) open SSE
    const query = new URLSearchParams(
      Object.entries(queryParams).reduce((acc, [key, value]) => {
        if (value !== null) {
          acc[key] = value.toString();
        }
        return acc;
      }, {} as Record<string, string>)
    ).toString();


    
    const es = new EventSource(`http://localhost:8080/stream?${query}`);
    evtSourceRef.current = es;
    es.onmessage = (e) => {
      try {
        const updates: Update[] = JSON.parse(e.data);
        queueRef.current.push(...updates);
      } catch (err) {
        console.error("Failed to parse SSE data", err);
      }
    };
    es.onerror = () => {
      console.warn("SSE error/closed");
      es.close();
    };

    let isPaused = false;
    const timeoutRef = { current: 0 as number };

    function scheduleNext() {
      const start = performance.now();
      if (queueRef.current.length) {
        let next = queueRef.current.shift()!;
        if ((next.Stage === "startDFS" || next.Stage === "startBFS") && !hasRoot.current) {
          hasRoot.current = true;
          const newNodes = [
            {
              id: "node-1",
              type: "single",
              data: {
                label: next.ElementName,
                imageLink: elementsData[next.ElementName]?.imageLink || "",
                targetHandles: [],
                sourceHandles: [{ id: `parent-root` }],
              },
              // set position to center
              position: { x: 0, y: 0 },
              width: BOX_WIDTH,
              height: BOX_HEIGHT,
            },
          ];
          setNodes(newNodes);
          setEdges([]);
          rootNode.current = newNodes[0];
          layoutFlow("DOWN", newNodes, []);
        } else {
          while (next.Stage !== "startRecipe") {
            if (
              next.Stage == "doneRecipe" &&
              next.ElementName == rootNode.current?.data.label
            ) {
              // we are done
              console.log("Done with recipe");
              hasStartedRef.current = false;
              onFinish();
              if (timeoutRef.current) {
                clearTimeout(timeoutRef.current);
                timeoutRef.current = 0;
              }
              return;
            }
            // pop until we find a startRecipe
            if (queueRef.current.length === 0) {
              break;
            }
            next = queueRef.current.shift()!;
          }
          if (next.Stage === "startRecipe") {
            addNode(
              next.ParentID,
              next.LeftID,
              next.RightID,
              next.LeftLabel,
              next.RightLabel
            );
          }
        }
      }
      const elapsed = performance.now() - start;
      timeoutRef.current = window.setTimeout(
        scheduleNext,
        // ms
        Math.max(0, 800 - elapsed)
      );
    }

    document.addEventListener("visibilitychange", () => {
      if (document.hidden) {
        isPaused = true;
        if (timeoutRef.current) {
          clearTimeout(timeoutRef.current);
          timeoutRef.current = 0;
        }
      } else {
        isPaused = false;
        scheduleNext(); // resume ticking
      }
    });

    scheduleNext();

    // 3) cleanup
    return () => {
      clearTimeout(timeoutRef.current);
      es.close();
    };
  }, [loading, addNode, queryParams]);

  useEffect(() => {
    if (!trigger || hasStartedRef.current) return;
    if (evtSourceRef.current) {
      evtSourceRef.current.close();
      evtSourceRef.current = null;
    }

    setNodes([]);
    setEdges([]);
    queueRef.current = [];
    hasStartedRef.current = false;
    hasRoot.current = false;
    rootNode.current = null;
    nodeCountRef.current = 1;

    if (rfInstance) {
      rfInstance.setViewport({ x: 0, y: 0, zoom: 1 });
    }

    const cleanup = liveUpdate();
    return cleanup;
  }, [trigger, queryParams]);

  const findAncestorEdges = useCallback((nodeId: string, allEdges: Edge[]) => {
    const visited = new Set<string>();
    const highlight = new Set<string>();
    function traverse(id: string) {
      allEdges.forEach((e) => {
        if (e.target === id && !visited.has(e.id)) {
          visited.add(e.id);
          highlight.add(e.id);
          traverse(e.source);
        }
      });
    }
    traverse(nodeId);
    return Array.from(highlight);
  }, []);

  const findDescendantEdges = useCallback(
    (nodeId: string, allEdges: Edge[]) => {
      const visited = new Set<string>();
      const highlight = new Set<string>();
      function traverse(id: string) {
        allEdges.forEach((e) => {
          if (e.source === id && !visited.has(e.id)) {
            visited.add(e.id);
            highlight.add(e.id);
            traverse(e.target);
          }
        });
      }
      traverse(nodeId);
      return Array.from(highlight);
    },
    []
  );

  const onNodeMouseEnter = useCallback(
    (_: React.MouseEvent, node: Node) => {
      const downIds = findDescendantEdges(node.id, edges);
      const upIds = findAncestorEdges(node.id, edges);
      const ids = Array.from(new Set([...downIds, ...upIds]));

      setEdges((es) =>
        es.map((e) => ({
          ...e,
          style: {
            ...e.style,
            stroke: ids.includes(e.id) ? edgeColorHighlight : edgeColorDefault,
            strokeWidth: ids.includes(e.id)
              ? edgeWidthHighlight
              : edgeWidthDefault,
            strokeDasharray: ids.includes(e.id) ? "6 4" : "0",
            animation: ids.includes(e.id) // cool shyt
              ? "dash 1s linear infinite"
              : "none",
          },
        }))
      );
    },
    [edges, findDescendantEdges, findAncestorEdges]
  );

  const onNodeMouseLeave = useCallback(() => {
    setEdges((es) =>
      es.map((e) => ({
        ...e,
        style: {
          stroke: edgeColorDefault,
          strokeWidth: edgeWidthDefault,
          strokeDasharray: "0",
          animation: "none",
        },
      }))
    );
  }, []);

  // Capture instance on init
  const onInit = useCallback((instance: ReactFlowInstance) => {
    setRfInstance(instance);
  }, []);

  const onNodesChange: OnNodesChange = useCallback(
    (changes) => setNodes((nds) => applyNodeChanges(changes, nds)),
    []
  );
  const onEdgesChange: OnEdgesChange = useCallback(
    (changes) => setEdges((eds) => applyEdgeChanges(changes, eds)),
    []
  );
  const onConnect: OnConnect = useCallback(
    (connection: Connection) => setEdges((eds) => addEdge(connection, eds)),
    []
  );

  const deleteNode = useCallback(() => {
    // don’t delete the only node
    if (nodes.length <= 1) return;

    // pick a random non-root node
    const deletable = nodes.filter((n) => n.id !== "1");
    const nodeToDelete =
      deletable[Math.floor(Math.random() * deletable.length)];

    // find its incoming edge (so we know its parent)
    const parentEdge = edges.find((e) => e.target === nodeToDelete.id);
    const parentId = parentEdge?.source;

    // find all outgoing edges (its children)
    const childEdges = edges.filter((e) => e.source === nodeToDelete.id);

    // remove the node & all its incident edges
    let newNodes = nodes.filter((n) => n.id !== nodeToDelete.id);
    let newEdges = edges.filter(
      (e) => e.source !== nodeToDelete.id && e.target !== nodeToDelete.id
    );

    // re‑attach each of its children to its parent
    if (parentId) {
      const reconnected = childEdges.map((e) => ({
        id: `e_${parentId}_${e.target}`,
        source: parentId,
        target: e.target,
        type: "smoothstep",
        markerStart: {
          type: MarkerType.ArrowClosed,
          width: 10,
          height: 10,
        },
        style: {
          strokeWidth: 2,
        },
      }));
      newEdges = [...newEdges, ...reconnected];
    }

    setNodes(newNodes);
    setEdges(newEdges);

    // re‑layout so things stay neat
    setTimeout(() => layoutFlow("DOWN", newNodes, newEdges), 0);
  }, [nodes, edges, layoutFlow]);

  if (loading) return <p>Loading...</p>;

  return (
    <ReactFlowProvider>
      <div style={{ width: "100%", height: "100vh" }}>
        <div style={{ position: "absolute", zIndex: 10, right: 10, top: 10 }}>
          <button onClick={() => layoutFlow("DOWN")}>Fix Layout</button>
          <button onClick={liveUpdate}>Live Update</button>
          {/*
          <button onClick={deleteNode}>Delete Node</button>
          */}
        </div>
        <ReactFlow
          nodeTypes={nodeTypes}
          nodes={nodes}
          edges={edges}
          onInit={onInit}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeMouseEnter={onNodeMouseEnter}
          onNodeMouseLeave={onNodeMouseLeave}
        >
          <MiniMap />
          <Controls />
          <Background />
        </ReactFlow>
      </div>
    </ReactFlowProvider>
  );
};

export default TreeViewer;
