"use client";
import React, { useState, useEffect } from "react";
import { Sidebar, TreeViewer, StatsPanel } from "@/components";
import { ElementsData } from "@/types";
import toast from "react-hot-toast";

// im kinda lazy to move these interface to /types
interface QueryParams {
    element: string | null;
    algorithm: string | null;
    multipleRecipes: boolean;
    liveUpdate: boolean;
    count: number | "all";
  }

interface StatsPanelProps {
  nodesExplored: number;
  timeTaken: string;
  /*  optional extra classes if i need to tweak positioning */
  className?: string;
}

const loadingMessages = [
  "Cooking recipes 🔥",
  "Mixing elements ⚗️",
  "Brewing formula 🧪",
  "Sifting powders",
  "Charting paths 🗺️",
  "Finding routes 🧭",
];

const Page: React.FC = () => {
  // State for sidebar
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [stats, setStats] = useState<StatsPanelProps>({
    nodesExplored:  -1,
    timeTaken: "-",
  });

  const [mainToast, setMainToast] = useState<string | undefined>();


  // State for elements data
  const [elementsData, setElementsData] = useState<ElementsData>({});
  const [loading, setLoading] = useState(true);
  const [queryParams, setQueryParams] = useState<QueryParams>({
    element: "",
    algorithm: "",
    multipleRecipes: false,
    liveUpdate: false,
    count: 0,
  });
  const [shouldSendRequest, setShouldSendRequest] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);

  // Fetch elements data on component mount
  useEffect(() => {
    const fetchData = async () => {
      const localData = localStorage.getItem("elements_data");
      if (localData) {
        try {
          const parsed = JSON.parse(localData) as ElementsData;
          setElementsData(parsed);
          setLoading(false);
          console.log("Data loaded from localStorage:", parsed);
          return;
        } catch (err) {
          console.error("Invalid JSON in localStorage:", err);
          localStorage.removeItem("elements_data");
        }
      }

      try {
        const response = await fetch("http://localhost:8080/api/data"); // replace with actual API later
        const result = await response.json();
        localStorage.setItem("elements_data", JSON.stringify(result));
        setElementsData(result);
      } catch (err) {
        console.error("Failed to fetch data:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  // Toggle sidebar visibility
  const toggleSidebar = () => {
    setSidebarOpen(!sidebarOpen);
  };

  const handleQueryParamsChange = (params: QueryParams) => {
    setQueryParams(params);
    setShouldSendRequest(true);
    console.log("Query Params Updated:", params);
    setIsProcessing(true);
    setStats({
      nodesExplored: -1,
      timeTaken: "-",
    });
    if(params.liveUpdate){
    const randomMsg =
      loadingMessages[Math.floor(Math.random() * loadingMessages.length)];
    setMainToast(toast.loading(randomMsg));
    }
    setSidebarOpen(false);
  };

  const handleFinishProcess = (stats : StatsPanelProps) => {
    setShouldSendRequest(false);
    setIsProcessing(false);
    setStats(stats);
    toast.success("Process Finished", { id: mainToast });
    console.log("Process Finished:");
  }

  const handleErrorProcess = (error: string) => {
    setShouldSendRequest(false);
    setIsProcessing(false);
    toast.error("Process Failed: " + error, { id: mainToast });
    console.log("Process Failed:", error);
  }

  return (
    <div className="flex min-h-screen bg-white relative">
      {/* Main content */}
      
      <TreeViewer elementsData={elementsData} loading={loading} queryParams={queryParams} trigger={shouldSendRequest} onFinish={handleFinishProcess} onError={handleErrorProcess}/>

      {/* Sidebar Toggle Button - visible when sidebar is closed */}
      {!sidebarOpen && (
        <button
          onClick={toggleSidebar}
          className="fixed top-4 left-4 z-20 bg-gray-700 p-2 rounded"
        >
          <svg viewBox="0 0 24 24" fill="white" className="w-6 h-6">
            <path d="M4 6h16M4 12h16M4 18h16" stroke="white" strokeWidth="2" />
          </svg>
        </button>
      )}

      {/* Sidebar Component */}
      <Sidebar
        isOpen={sidebarOpen}
        isProcessing={isProcessing}
        onToggle={toggleSidebar}
        onQueryParamsChange={handleQueryParamsChange}
        elementsData={elementsData}
        loading={loading}
      />

      {/* Stats Panel */}
      <StatsPanel
        nodesExplored={stats.nodesExplored}
        timeTaken={stats.timeTaken}
      />
        
    </div>
  );
};

export default Page;
