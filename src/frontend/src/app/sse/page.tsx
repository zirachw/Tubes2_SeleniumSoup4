'use client';

import React from 'react';
import { DfsSseViewer } from '@/components/DfsSseViewer';

export default function Page() {
  return (
    <main className="min-h-screen p-8">
      <h1 className="text-3xl font-bold mb-6">DFS Live Stream</h1>
      <DfsSseViewer />
    </main>
  );
}
