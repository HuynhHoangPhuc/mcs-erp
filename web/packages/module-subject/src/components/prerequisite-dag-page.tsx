// Prerequisite DAG visualization page using React Flow + dagre auto-layout.
import { useState, useEffect, useCallback, useMemo } from "react";
import { useQueries } from "@tanstack/react-query";
import { ReactFlow, Controls, Background, useNodesState, useEdgesState, type Node, type Edge } from "@xyflow/react";
import Dagre from "@dagrejs/dagre";
import "@xyflow/react/dist/style.css";
import { Button, LoadingSpinner } from "@mcs-erp/ui";
import { useSubjects, type Prerequisite } from "@mcs-erp/api-client";
import { apiFetch } from "@mcs-erp/api-client";
import { SubjectNode, type SubjectNodeData } from "./subject-node";
import { AddPrerequisiteDialog } from "./add-prerequisite-dialog";

const NODE_TYPES = { subject: SubjectNode };

function applyDagreLayout(nodes: Node[], edges: Edge[]): Node[] {
  const g = new Dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({ rankdir: "TB", nodesep: 60, ranksep: 80 });
  for (const node of nodes) g.setNode(node.id, { width: 160, height: 80 });
  for (const edge of edges) g.setEdge(edge.source, edge.target);
  Dagre.layout(g);
  return nodes.map((node) => {
    const pos = g.node(node.id);
    return { ...node, position: { x: pos.x - 80, y: pos.y - 40 } };
  });
}

export function PrerequisiteDagPage() {
  const [addOpen, setAddOpen] = useState(false);
  const [highlightedChain, setHighlightedChain] = useState<Set<string>>(new Set());

  const { data: subjectsData, isLoading: subjectsLoading } = useSubjects({ limit: 1000 });
  const subjects = useMemo(() => subjectsData?.items ?? [], [subjectsData]);

  // Parallel queries — useQueries is designed for dynamic arrays, no hooks-in-loop.
  const prerequisiteQueries = useQueries({
    queries: subjects.map((s) => ({
      queryKey: ["subjects", s.id, "prerequisites"],
      queryFn: () => apiFetch<Prerequisite[]>(`/subjects/${s.id}/prerequisites`),
      enabled: subjects.length > 0,
    })),
  });

  const edgesLoading = prerequisiteQueries.some((q) => q.isLoading);

  const rawEdges: Edge[] = useMemo(() => {
    const result: Edge[] = [];
    for (const q of prerequisiteQueries) {
      for (const p of q.data ?? []) {
        result.push({
          id: `${p.prerequisite_id}->${p.subject_id}`,
          source: p.prerequisite_id,
          target: p.subject_id,
          style: { stroke: "#94a3b8" },
        });
      }
    }
    return result;
    // prerequisiteQueries array ref changes every render; stringify for stable dep
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [prerequisiteQueries.map((q) => q.dataUpdatedAt).join(",")]);

  // Build nodes without highlight first (for stable dagre input).
  const baseNodes: Node[] = useMemo(
    () =>
      subjects.map((s) => ({
        id: s.id,
        type: "subject",
        position: { x: 0, y: 0 },
        data: { label: s.name, code: s.code, credits: s.credits, highlighted: false } satisfies SubjectNodeData,
      })),
    [subjects]
  );

  // Re-run dagre only when subjects or edges change, not on highlight.
  const layoutNodes = useMemo(
    () => (baseNodes.length > 0 ? applyDagreLayout(baseNodes, rawEdges) : []),
    [baseNodes, rawEdges]
  );

  // Overlay highlight without re-running dagre layout.
  const displayNodes = useMemo(
    () =>
      layoutNodes.map((n) => ({
        ...n,
        data: { ...(n.data as unknown as SubjectNodeData), highlighted: highlightedChain.has(n.id) },
      })),
    [layoutNodes, highlightedChain]
  );

  const [nodes, setNodes, onNodesChange] = useNodesState(displayNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(rawEdges);

  useEffect(() => { setNodes(displayNodes); }, [displayNodes, setNodes]);
  useEffect(() => { setEdges(rawEdges); }, [rawEdges, setEdges]);

  // BFS upward through edges to find full ancestor chain on click.
  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      if (highlightedChain.has(node.id) && highlightedChain.size === 1) {
        setHighlightedChain(new Set());
        return;
      }
      const chain = new Set<string>([node.id]);
      const queue = [node.id];
      while (queue.length > 0) {
        const current = queue.shift()!;
        for (const e of rawEdges) {
          if (e.target === current && !chain.has(e.source)) {
            chain.add(e.source);
            queue.push(e.source);
          }
        }
      }
      setHighlightedChain(chain);
    },
    [highlightedChain, rawEdges]
  );

  const isLoading = subjectsLoading || edgesLoading;

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)]">
      <div className="flex items-center justify-between p-4 border-b bg-background">
        <div>
          <h1 className="text-xl font-bold tracking-tight">Prerequisite DAG</h1>
          <p className="text-muted-foreground text-sm">Click a node to highlight its prerequisite chain.</p>
        </div>
        <Button onClick={() => setAddOpen(true)}>Add Prerequisite</Button>
      </div>

      <div className="flex-1 relative">
        {isLoading ? (
          <div className="flex items-center justify-center h-full">
            <LoadingSpinner />
          </div>
        ) : (
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onNodeClick={onNodeClick}
            nodeTypes={NODE_TYPES}
            fitView
            fitViewOptions={{ padding: 0.2 }}
            minZoom={0.2}
            maxZoom={2}
          >
            <Controls />
            <Background />
          </ReactFlow>
        )}
      </div>

      <AddPrerequisiteDialog open={addOpen} onOpenChange={setAddOpen} />
    </div>
  );
}
