"use client";
// components/CoupleNode.tsx
import React from "react";
import { Handle, Position } from "reactflow";
import Image from "next/image";
import { getElementBackgroundColor } from "@/utils/getElementBackgroundColor";
interface CoupleData {
  LeftLabel: string;
  RightLabel: string;
  leftImageLink: string;
  rightImageLink: string;
  id: string;
  RightID: number;
  LeftID: number;
}
export const BOX_WIDTH = 96;
export const BOX_HEIGHT = 48;
export const GAP = 12; // space between boxes (plus sign sits here)
export const PADDING = 8;

const containerStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  padding: PADDING,
  width: 2 * BOX_WIDTH + GAP + 5 * PADDING,
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

export default function CoupleNode({ data }: { data: CoupleData }) {
  return (
    <div style={containerStyle}>
      {/* single parent handle */}
      <Handle
        type="target"
        id={`parent-${data.id}`}
        position={Position.Top}
        style={{
          left: BOX_WIDTH + 2.5 * PADDING + GAP / 2,
          top: -BOX_HEIGHT / 8,
        }}
      />

      {/* mom box + its child handle */}
      <div style={{ position: "relative", marginRight: GAP / 2 }}>
        <div style={getBoxStyle(data.LeftLabel)} title={data.LeftLabel}>
          <Image
            src={data.leftImageLink}
            alt={data.LeftLabel}
            width={30}
            height={30}
            crossOrigin="anonymous"
            style={imageStyle}
          ></Image>
          <span style={labelStyle}>{data.LeftLabel}</span>
        </div>
        <Handle
          type="source"
          id={`${data.LeftID}`}
          position={Position.Bottom}
          style={{ left: BOX_WIDTH / 2 }}
        />
      </div>

      {/* plus sign */}
      <div
        style={{
          userSelect: "none",
          fontSize: 18,
          margin: "0 6px",
          color: "#555", // ← explicit color so it shows on white
        }}
      >
        +
      </div>

      {/* dad box + its child handle */}
      <div style={{ position: "relative", marginLeft: GAP / 2 }}>
        <div style={getBoxStyle(data.RightLabel)} title={data.RightLabel}>
          <Image
            src={data.rightImageLink}
            alt={data.RightLabel}
            width={30}
            height={30}
            crossOrigin="anonymous"
            style={imageStyle}
          ></Image>
          <span style={labelStyle}>{data.RightLabel}</span>
        </div>
        <Handle
          type="source"
          id={`${data.RightID}`}
          position={Position.Bottom}
          style={{ left: BOX_WIDTH / 2 }}
        />
      </div>
    </div>
  );
}