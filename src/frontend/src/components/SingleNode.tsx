"use client";

// components/SingleNode.tsx
import React from "react";
import { Handle, Position } from "reactflow";
import Image from "next/image";
import { getElementBackgroundColor } from "@/utils/getElementBackgroundColor";

interface SingleData {
  label: string;
  imageLink: string;
  id: string;
}

export const BOX_WIDTH = 128;
export const BOX_HEIGHT = 48;
export const PADDING = 8;

const containerStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  padding: PADDING,
  width:  BOX_WIDTH + 2 * PADDING,
  height: BOX_HEIGHT + 2 * PADDING,
  position: "relative",
};

const getBoxStyle = (label: string): React.CSSProperties => ({
  width: BOX_WIDTH,
  height: BOX_HEIGHT,
  padding: PADDING,
  border: "1px solid #777",
  borderRadius: 4,
  display: "flex",
  alignItems: "center",
  justifyContent: "left",
  color: "#FFA500",
  background: getElementBackgroundColor(label),
  overflow: "hidden",
  textOverflow: "ellipsis",
});

const labelStyle: React.CSSProperties = {
  marginLeft: 6,
  whiteSpace: "nowrap",
  overflow: "hidden",
  textOverflow: "ellipsis",
  fontSize: 14,
  color: "#333",
};

const imageStyle: React.CSSProperties = {
  width: 30,
  height: 30,
  filter: "drop-shadow(0px 0px 4px rgba(0, 0, 0, 0.3))",
};

export default function SingleNode({ data }: { data: SingleData }) {
  return (
    <div style={containerStyle}>
      {/* Single box + its child handle */}
      <div style={{ position: "relative"}}>
        <div style={getBoxStyle(data.label)} title={data.label}>
          <Image
            src={data.imageLink}
            alt={data.label}
            width={30}
            height={30}
            crossOrigin="anonymous"
            style={imageStyle}
          ></Image>
          <span style={labelStyle}>{data.label}</span>
        </div>
        <Handle
          type="source"
          id={`left-child-${data.id}`}
          position={Position.Bottom}
          style={{ left: BOX_WIDTH / 2 }}
        />
      </div>
    </div>
  );
}
