"use client";

// components/CoupleNode.tsx
import React from "react";
import { Handle, Position } from "reactflow";
import Image from "next/image";

interface CoupleData {
  leftLabel: string;
  rightLabel: string;
  leftImageLink: string;
  rightImageLink: string;
}

const BOX_WIDTH = 80;
const BOX_HEIGHT = 40;
const GAP = 12; // space between boxes (plus sign sits here)
const PADDING = 8;

const containerStyle: React.CSSProperties = {
  display: "flex",
  alignItems: "center",
  padding: PADDING,
  // total width = 2*BOX_WIDTH + GAP + 2*PADDING
  width: 2 * BOX_WIDTH + GAP + 2 * PADDING,
  // height = BOX_HEIGHT + 2*PADDING
  height: BOX_HEIGHT + 2 * PADDING,
  position: "relative",
};

const boxStyle: React.CSSProperties = {
  width: BOX_WIDTH,
  height: BOX_HEIGHT,
  border: "1px solid #777",
  borderRadius: 4,
  display: "flex",
  alignItems: "center",
  justifyContent: "left",
  color: "#FFA500",
  background: "#fff",
  overflow: "hidden",
  textOverflow: "ellipsis",
};

export default function CoupleNode({ data }: { data: CoupleData }) {
  return (
    <div style={containerStyle}>
      {/* single parent handle */}
      <Handle
        type="target"
        id="parent"
        position={Position.Top}
        style={{
          left: BOX_WIDTH + 2 * PADDING + GAP / 2,
          top: -BOX_HEIGHT / 8,
        }}
      />

      {/* mom box + its child handle */}
      <div style={{ position: "relative", marginRight: GAP / 2 }}>
        <div style={boxStyle}>
          <Image
            src={data.leftImageLink}
            alt={data.leftLabel}
            width={30}
            height={30}
            crossOrigin="anonymous"
          ></Image>
          {data.leftLabel}
        </div>
        <Handle
          type="source"
          id="left-child"
          position={Position.Bottom}
          style={{ left: BOX_WIDTH / 2 }}
        />
      </div>

      {/* plus sign */}
      <div style={{ userSelect: "none", fontSize: 16, margin: "0 4px" }}>+</div>

      {/* dad box + its child handle */}
      <div style={{ position: "relative", marginLeft: GAP / 2 }}>
        <div style={boxStyle}>
          <Image
            src={data.rightImageLink}
            alt={data.rightLabel}
            width={30}
            height={30}
            crossOrigin="anonymous"
          ></Image>
          {data.rightLabel}
        </div>
        <Handle
          type="source"
          id="right-child"
          position={Position.Bottom}
          style={{ left: BOX_WIDTH / 2 }}
        />
      </div>
    </div>
  );
}
