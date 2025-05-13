"use client";

import React from "react";

interface StatsPanelProps {
  nodesExplored: number;
  nodesInTree: number;
  uniquePaths: number;
  timeTaken: string;
  /*  optional extra classes if i need to tweak positioning */
  className?: string;
}

const StatsPanel: React.FC<StatsPanelProps> = ({
  nodesExplored,
  nodesInTree,
  uniquePaths,
  timeTaken,
  className = "",
}) => {
  return (
    <div
      className={
        `absolute top-4 right-4 z-20 bg-white bg-opacity-90 shadow-lg rounded-lg p-4 flex flex-col space-y-2 w-52 ` +
        className
      }
    >
       <div className="flex justify-between">
        <span className="text-sm font-medium text-gray-600">Unique Paths</span>
        <span className="text-sm font-semibold text-gray-900">
          {uniquePaths != -1 ? uniquePaths: "-"}
        </span>
      </div>
      <div className="flex justify-between">
        <span className="text-sm font-medium text-gray-600">Nodes Explored</span>
        <span className="text-sm font-semibold text-gray-900">
          {nodesExplored != -1 ? nodesExplored : "-"}
        </span>
      </div>
       <div className="flex justify-between">
        <span className="text-sm font-medium text-gray-600">Nodes in Tree</span>
        <span className="text-sm font-semibold text-gray-900">
          {nodesInTree != -1 ? nodesInTree: "-"}
        </span>
      </div>
      <div className="flex justify-between">
        <span className="text-sm font-medium text-gray-600">Time Taken</span>
        <span className="text-sm font-semibold text-gray-900">
          {timeTaken}
        </span>
      </div>
    </div>
  );
};

export default StatsPanel;
