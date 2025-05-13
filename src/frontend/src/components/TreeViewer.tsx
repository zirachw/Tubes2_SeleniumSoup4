"use client";

import React, { useState, useCallback, useRef, useEffect } from "react";
import { flushSync } from "react-dom";
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
import { on } from "events";

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

interface StatsPanelProps {
  nodesExplored: number;
  nodesInTree: number;
  uniquePaths: number;
  timeTaken: string;
  /*  optional extra classes if i need to tweak positioning */
  className?: string;
}

// hhihi
const minimapNodeColor = (node: Node) => {
  switch (node.type) {
    case "single":
      return "#ff0072";
    case "couple":
      return "#d1d1d1";
    default:
      // greyish more white than coupe
      return "#f0f0f0";
  }
};
const getLayoutedElements = async (
  nodes: Node[],
  edges: Edge[],
  direction = "DOWN"
) => {
  const rightSideNodes = new Set<string>();
  edges.forEach((edge) => {
    if (
      parseInt(edge.sourceHandle ?? "") % 2 === 0 ||
      edge.sourceHandle?.startsWith("right-")
    ) {
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
  onFinish: (stats: StatsPanelProps) => void;
  onError: (error: string) => void;
}

const TreeViewer: React.FC<TreeViewerProps> = ({
  elementsData,
  loading,
  queryParams,
  trigger,
  onFinish,
  onError,
}) => {
  const [nodes, setNodes] = useState<Node[]>([]);
  const queueRef = useRef<Update[]>([]);
  const resultRef = useRef<any | null>(null);
  const evtSourceRef = useRef<EventSource | null>(null);
  const [edges, setEdges] = useState<Edge[]>([]);
  const hasStartedRef = useRef(false);
  const timeoutRef = useRef<number | null>(null);
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
        const temp: any = JSON.parse(e.data);
        if (temp.hasOwnProperty("recipeTree")) {
          resultRef.current = temp;
        } else {
          const updates: Update[] = temp;
          queueRef.current.push(...updates);
        }
      } catch (err) {
        console.error("Failed to parse SSE data", err);
        onError("Failed to parse SSE data");
        hasStartedRef.current = false;
        if (evtSourceRef.current) {
          evtSourceRef.current.close();
          evtSourceRef.current = null;
        }
        return;
      }
    };
    es.onerror = () => {
      console.warn("SSE error/closed");
      es.close();
    };

    let isPaused = false;

    function scheduleNext() {
      if (queueRef.current.length) {
        let next = queueRef.current.shift()!;
        if (
          (next.Stage === "startDFS" || next.Stage === "startBFS") &&
          !hasRoot.current
        ) {
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
              if (
                resultRef.current &&
                resultRef.current.hasOwnProperty("recipeTree")
              ) {
                onFinish({
                  nodesExplored: resultRef.current.nodesExplored,
                  nodesInTree: resultRef.current.nodesInTree,
                  uniquePaths: resultRef.current.uniquePaths,
                  timeTaken: resultRef.current.timeTaken,
                });
              } else {
                onFinish({ nodesExplored: -1, nodesInTree: -1, uniquePaths: -1, timeTaken: "Result not found" });
              }
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
      timeoutRef.current = window.setTimeout(
        scheduleNext,
        // ms
        800
      );
    }

    scheduleNext();

    // 3) cleanup
    return () => {
      es.close();
    };
  }, [loading, addNode, queryParams]);

  const buildGraph = useCallback(
    (treeData: any) => {
      console.log("Building graph from treeData", treeData);
      hasStartedRef.current = true;
      const newNodes: Node[] = [];
      const newEdges: Edge[] = [];

      function walkPair(
        left: any,
        right: any,
        parentId: string | null,
        isLeftBranch: boolean
      ) {
        const id = makeId();
        const LeftLabel = left.name;
        const RightLabel = right.name;
        const leftImageLink = elementsData[left.name]?.imageLink || "";
        const rightImageLink = elementsData[right.name]?.imageLink || "";

        // 1) add the couple‐node itself
        newNodes.push({
          id,
          type: "couple",
          data: {
            LeftLabel,
            RightLabel,
            LeftID: `left-child-${id}`,
            RightID: `right-child-${id}`,
            leftImageLink,
            rightImageLink,
            id,
            targetHandles: [{ id: `parent-${id}` }],
            sourceHandles: [
              { id: `left-child-${id}` },
              { id: `right-child-${id}` },
            ],
          },
          position: { x: 0, y: 0 },
          width: BOX_WIDTH * 2 + GAP + PADDING * 5,
          height: BOX_HEIGHT + PADDING * 2,
        });

        // 2) if we have a parent, wire up an edge into this node
        if (parentId) {
          const sourceHandle = isLeftBranch
            ? `left-child-${parentId}`
            : `right-child-${parentId}`;

          // handle source if parent is root
          const sourceHandleRoot = isLeftBranch ? "parent-root" : "parent-root";
          const sourceHandleParent =
            parentId === "node-1" ? sourceHandleRoot : sourceHandle;
          newEdges.push({
            id: `e-${parentId}-${id}`,
            source: parentId,
            sourceHandle: sourceHandleParent,
            target: id,
            targetHandle: `parent-${id}`,
            type: "smoothstep",
            markerStart: {
              type: MarkerType.ArrowClosed,
              width: 8,
              height: 8,
            },
            style: { strokeWidth: 2 },
          });
        }

        // 3) if this JSON pair itself has deeper children, recurse
        if (Array.isArray(left.recipes) && left.recipes.length) {
          // our shape is { children: [ { left: {...}, right: {...} }, ... ] }
          left.recipes.forEach((pair: any) =>
            walkPair(pair.left, pair.right, id, true)
          );
        }
        if (Array.isArray(right.recipes) && right.recipes.length) {
          right.recipes.forEach((pair: any) =>
            walkPair(pair.left, pair.right, id, false)
          );
        }
      }

      nodeCountRef.current = 0;

      // 1️⃣ first handle the very top: JSON root has .value + .children[0]

      if (treeData?.recipeTree) {
        const rootId = `node-1`;
        // create a dummy “root” node so we can wire into the first couple
        newNodes.push({
          id: "node-1",
          type: "single",
          data: {
            label: treeData.element,
            imageLink: elementsData[treeData.element]?.imageLink || "",
            targetHandles: [],
            sourceHandles: [{ id: `parent-root` }],
          },
          // set position to center
          position: { x: 0, y: 0 },
          width: 128 + 2 * PADDING, // hell yeah @ref SingleNode.tsx
          height: BOX_HEIGHT,
        });

        // now hook up the first real couple(s)
        if (treeData.recipeTree.recipes) {
          treeData.recipeTree.recipes.forEach((pair: any) => {
            walkPair(pair.left, pair.right, rootId, true);
          });
        }
      }

      console.log("Done with recipe");
      hasStartedRef.current = false;
      // malas validate
      onFinish({
        nodesExplored: treeData.nodesExplored,
        nodesInTree: treeData.nodesInTree,
        uniquePaths: treeData.uniquePaths,
        timeTaken: treeData.timeTaken,
      });

      flushSync(() => {
        setNodes(newNodes);
        setEdges(newEdges);
      });
      layoutFlow("DOWN", newNodes, newEdges);
      if (rfInstance) {
        rfInstance.fitView({ padding: 0.2, duration: 400 });
      }
    },
    [elementsData, rfInstance]
  );

  const directTree = useCallback(() => {
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
        const temp: any = JSON.parse(e.data);
        // check if a key is exist
        if (temp.hasOwnProperty("recipeTree")) {
          const treeData = temp;
          buildGraph(treeData);
        }
      } catch (err) {
        console.error("Failed to parse SSE data", err);
        onError("Failed to parse SSE data");
        hasStartedRef.current = false;
        if (evtSourceRef.current) {
          evtSourceRef.current.close();
          evtSourceRef.current = null;
        }
        return;
      }
    };
    es.onerror = () => {
      console.warn("SSE error/closed");
      es.close();
    };

    // 3) cleanup
    return () => {
      es.close();
    };
  }, [loading, buildGraph, queryParams]);

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

    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }

    if (rfInstance) {
      rfInstance.setViewport({ x: 0, y: 0, zoom: 1 });
    }
    if (queryParams.liveUpdate == true) {
      const cleanup = liveUpdate();
      return cleanup;
    } else {
      const cleanup = directTree();
      return cleanup;
    }
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

  if (loading) return <p>Loading...</p>;

  return (
    <ReactFlowProvider>
      <div style={{ width: "100%", height: "100vh" }}>
        <div
          style={{ position: "absolute", zIndex: 10, right: 10, top: 10 }}
        ></div>
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
          <MiniMap nodeColor={minimapNodeColor} zoomable pannable />
          <Controls onFitView={() => layoutFlow("DOWN")} />
          <Background />
        </ReactFlow>
      </div>
    </ReactFlowProvider>
  );
};

export default TreeViewer;
