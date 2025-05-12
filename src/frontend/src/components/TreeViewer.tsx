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
} from "reactflow";
import "reactflow/dist/style.css";
import { CoupleNode } from "@/components";
import { ElementsData } from "@/types";
import { BOX_HEIGHT, BOX_WIDTH, GAP, PADDING } from "./CoupleNode";

// Initialize ELK
const elk = new ELK();

const getLayoutedElements = async (
  nodes: Node[],
  edges: Edge[],
  direction = "DOWN"
) => {
  const rightSideNodes = new Set<string>();
  edges.forEach((edge) => {
    if (edge.sourceHandle?.includes("right-child-")) {
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
};

interface TreeViewerProps {
  elementsData: ElementsData;
  loading: boolean;
}

const TreeViewer: React.FC<TreeViewerProps> = ({ elementsData, loading }) => {
  const [nodes, setNodes] = useState<Node[]>([
    {
      id: "root",
      data: {
        label: "Root",
        targetHandles: [],
        sourceHandles: [{ id: "parent-root" }],
      },
      position: { x: 0, y: 0 },
      width: 150,
      height: 50,
    },
  ]);
  const [edges, setEdges] = useState<Edge[]>([]);
  const nodeCountRef = useRef(1);

  // Keep React Flow instance
  const [rfInstance, setRfInstance] = useState<ReactFlowInstance | null>(null);

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

      animateLayout(baseNodes, layoutedNodes, setNodes, 400);
      setEdges(layoutedEdges);

      // smooth transition on fitView
      setTimeout(() => {
        if (rfInstance) {
          rfInstance.fitView({ padding: 0.2, duration: 400 });
        }
      }, 400);
    },
    [nodes, edges, rfInstance]
  );

  // Add a node under its parent, then relayout + fitView
  const addNode = useCallback(() => {
    // pick a random parent couple
    const keys = Object.keys(elementsData || {});
    if (keys.length === 0) {
      return undefined; // Return undefined for empty dictionaries
    }
    const randomIndex = Math.floor(Math.random() * keys.length);
    const randomKey = keys[randomIndex];
    const parentId = nodes[Math.floor(Math.random() * nodes.length)].id;
    const id = `couple_${nodeCountRef.current++}`;
    const leftLabel = elementsData![randomKey].recipes[0][0];
    const rightLabel = elementsData![randomKey].recipes[0][1];
    const leftImageLink =
      elementsData![elementsData![randomKey].recipes[0][0]].imageLink;
    const rightImageLink =
      elementsData![elementsData![randomKey].recipes[0][1]].imageLink;

    const parent = nodes.find((n) => n.id === parentId)!;
    const { x: px, y: py } = parent.position;
    const ph = parent.height ?? 50;

    const newNode: Node = {
      id,
      type: "couple",
      data: {
        leftLabel,
        rightLabel,
        leftImageLink,
        rightImageLink,
        id,
        targetHandles: [{ id: `parent-${id}` }],
        sourceHandles: [
          { id: `left-child-${id}` },
          { id: `right-child-${id}` },
        ],
      },
      position: { x: px, y: py + ph + 50 },
      // must match containerStyle dimensions:
      width: 2 * BOX_WIDTH + GAP + 5 * PADDING,
      height: BOX_HEIGHT + 2 * PADDING,
    };

    const isLeft = Math.random() < 0.5;
    const sourceHandle = isLeft
      ? "left-child-" + parentId
      : "right-child-" + parentId;
    // handle source if parent is root
    const sourceHandleRoot = isLeft ? "parent-root" : "parent-root";
    const sourceHandleParent =
      parentId === "root" ? sourceHandleRoot : sourceHandle;

    const newEdge: Edge = {
      id: `e_${parentId}_${id}`,
      source: parentId,
      sourceHandle: sourceHandleParent,
      target: id,
      targetHandle: `parent-${id}`,
      type: "smoothstep",
      markerStart: {
        type: MarkerType.ArrowClosed,
        width: 10,
        height: 10,
      },
      style: { strokeWidth: 2 },
    };

    const newNodes = [...nodes, newNode];
    const newEdges = [...edges, newEdge];

    setNodes(newNodes);
    setEdges(newEdges);

    setTimeout(() => layoutFlow("DOWN", newNodes, newEdges), 0);
  }, [nodes, edges, layoutFlow]);

  const deleteNode = useCallback(() => {
    // don’t delete the only node
    if (nodes.length <= 1) return;

    // pick a random non-root node
    const deletable = nodes.filter((n) => n.id !== "root");
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
          <button onClick={() => layoutFlow("DOWN")}>Top-Down</button>
          <button onClick={() => layoutFlow("UP")}>Bottom-Up</button>
          <button onClick={addNode}>Add Node</button>
          <button onClick={deleteNode}>Delete Node</button>
        </div>
        <ReactFlow
          nodeTypes={nodeTypes}
          nodes={nodes}
          edges={edges}
          onInit={onInit}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
        >
          <Controls />
          <Background />
        </ReactFlow>
      </div>
    </ReactFlowProvider>
  );
};

export default TreeViewer;
