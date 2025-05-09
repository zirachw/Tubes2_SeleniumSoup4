"use client";
import React, { useState, useEffect } from 'react';
import Sidebar from './components/Sidebar';

interface ElementData {
  tier: number;
  imageLink: string;
  recipes: string[][];
}

interface ElementsData {
  [elementName: string]: ElementData;
}

const Page: React.FC = () => {
  // State for sidebar
  const [sidebarOpen, setSidebarOpen] = useState(true);
  
  // State for elements data
  const [elementsData, setElementsData] = useState<ElementsData>({});
  const [loading, setLoading] = useState(true);

  // Fetch elements data on component mount
  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/data');
        if (!response.ok) {
          throw new Error('Failed to fetch elements data');
        }
        const data = await response.json();
        setElementsData(data);
        setLoading(false);
      } catch (err) {
        console.error("Error fetching data:", err);
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  // Toggle sidebar visibility
  const toggleSidebar = () => {
    setSidebarOpen(!sidebarOpen);
  };

  return (
    <div className="flex min-h-screen bg-white">
      {/* Main content */}
      <div className="flex-1 flex flex-col items-center justify-center">
        {/* Logo */}
        <div className="mb-8">
          <svg viewBox="0 0 200 200" className="w-32 h-32">
            <path
              d="M100,20 C146.5,20 185,58.5 185,105 C185,151.5 146.5,190 100,190 C53.5,190 15,151.5 15,105 C15,58.5 53.5,20 100,20 Z"
              fill="none"
              stroke="black"
              strokeWidth="5"
            />
            <path
              d="M70,65 C70,55 80,45 90,45 C100,45 110,55 110,65 L110,120 C110,130 120,140 130,140 C140,140 150,130 150,120 L150,65"
              fill="none"
              stroke="black"
              strokeWidth="5"
            />
            <circle cx="70" cy="120" r="5" fill="black" />
            <circle cx="110" cy="140" r="5" fill="black" />
            <circle cx="70" cy="140" r="5" fill="black" />
            <circle cx="150" cy="140" r="5" fill="black" />
          </svg>
        </div>
        <h1 className="text-3xl font-bold mb-4">Welcome back to Master Alchemy V2</h1>
        <div className="text-sm absolute bottom-4 right-4">Copyright: SeleniumSoup4</div>
      </div>

      {/* Sidebar Toggle Button - visible when sidebar is closed */}
      {!sidebarOpen && (
        <button 
          onClick={toggleSidebar} 
          className="fixed top-4 left-4 z-20 bg-gray-700 p-2 rounded"
        >
          <svg viewBox="0 0 24 24" fill="white" className="w-6 h-6">
            <path d="M4 6h16M4 12h16M4 18h16" stroke="white" strokeWidth="2"/>
          </svg>
        </button>
      )}

      {/* Sidebar Component */}
      <Sidebar 
        isOpen={sidebarOpen} 
        onToggle={toggleSidebar}
        elementsData={elementsData}
        loading={loading}
      />
    </div>
  );
};

export default Page;