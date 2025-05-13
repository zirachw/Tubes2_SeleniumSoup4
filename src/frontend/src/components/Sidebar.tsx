"use client";
import React, { useState, useEffect } from "react";
import Image from "next/image";
import { ElementsData } from "@/types";

interface SidebarProps {
  isOpen: boolean;
  isProcessing: boolean;
  onToggle: () => void;
  onQueryParamsChange: (params: {
    element: string | null;
    algorithm: string | null;
    multipleRecipes: boolean;
    liveUpdate: boolean;
    count: number | "all";
  }) => void;
  elementsData: ElementsData;
  loading: boolean;
}

const Sidebar: React.FC<SidebarProps> = ({
  isOpen,
  isProcessing,
  onToggle,
  onQueryParamsChange,
  elementsData,
  loading,
}) => {
  // Search and selection states
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedElement, setSelectedElement] = useState<string | null>(null);
  const [selectedAlgorithm, setSelectedAlgorithm] = useState<string | null>(
    null
  );

  // Mode states
  const [isLiveUpdateMode, setIsLiveUpdateMode] = useState(false);
  const [isMultipleRecipeMode, setIsMultipleRecipeMode] = useState(false);

  // Counter states
  const [recipeCount, setRecipeCount] = useState(1);
  const [isAllSelected, setIsAllSelected] = useState(false);

  // Update recipe count when Multiple Recipe mode is toggled
  useEffect(() => {
    if (isMultipleRecipeMode && recipeCount < 2) {
      setRecipeCount(2);
    }
  }, [isMultipleRecipeMode]);

  // Filter elements based on search term
  const filteredElements = Object.keys(elementsData).filter((element) =>
    element.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // Handle element selection
  const handleElementSelect = (element: string) => {
    setSelectedElement(element);
    // Reset subsequent selections
    setSelectedAlgorithm(null);
    setIsMultipleRecipeMode(false);
    setIsLiveUpdateMode(false);
    setRecipeCount(1);
    setIsAllSelected(false);
  };

  // Handle algorithm selection
  const handleAlgorithmSelect = (algorithm: string) => {
    setSelectedAlgorithm(algorithm);
    // Reset subsequent selections
    setIsMultipleRecipeMode(false);
    setIsLiveUpdateMode(false);
    setRecipeCount(1);
    setIsAllSelected(false);
  };

  // Handle Live Update mode selection
  const handleLiveUpdateToggle = () => {
    if (selectedAlgorithm) {
      setIsLiveUpdateMode(!isLiveUpdateMode);
    }
  };

  // Handle Multiple Recipe mode selection
  const handleMultipleRecipeToggle = () => {
    if (selectedAlgorithm) {
      setIsMultipleRecipeMode(!isMultipleRecipeMode);
      if (isAllSelected) {
        setIsAllSelected(false);
      }
      // recipeCount will be updated in useEffect
    }
  };

  // Handle recipe count input change
  const handleRecipeCountChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    // Only allow numeric input
    if (!/^\d*$/.test(value)) {
      return;
    }

    if (value === "") {
      setRecipeCount(2);
      return;
    }

    const numValue = parseInt(value, 10);
    if (!isNaN(numValue)) {
      if (numValue < 2) setRecipeCount(2);
      else if (numValue > 50) setRecipeCount(50);
      else setRecipeCount(numValue);
    }
  };

  // Handle "All" button click
  const handleAllClick = () => {
    if (isMultipleRecipeMode) {
      setIsAllSelected(!isAllSelected);
    }
  };

  // Handle find button click
  const handleFindClick = () => {
    onQueryParamsChange({
      element: selectedElement,
      algorithm: selectedAlgorithm,
      multipleRecipes: isMultipleRecipeMode,
      liveUpdate: isLiveUpdateMode,
      count: isAllSelected ? "all" : recipeCount,
    });
  };

  return (
    <>
      {/* Hamburger menu button - visible when sidebar is closed */}
      {!isOpen && (
        <button
          onClick={onToggle}
          className="fixed top-4 left-4 z-20 bg-[#404040] p-2 rounded"
        >
          <svg viewBox="0 0 24 24" fill="white" className="w-6 h-6">
            <path d="M4 6h16M4 12h16M4 18h16" stroke="white" strokeWidth="2" />
          </svg>
        </button>
      )}

      {/* Sidebar */}
      <div
        className={`fixed top-0 left-0 h-full transition-all duration-300 ease-in-out z-10 ${
          isOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="h-full w-64 overflow-hidden bg-[#404040] text-[#DBDBDB] flex flex-col font-[Georgia]">
          {/* Header with close button */}
          <div className="p-4 flex justify-between items-center">
            {/* Group image and heading together */}
            <div className="flex items-center">
              {/* Wrap the image in a link that opens YouTube in a new tab */}
              <a
                href={"https://www.youtube.com/watch?v=W-0lSiV-H7k&t=91s"}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-block"
              >
                <img
                  src="/frieren.png"
                  alt="Beatiful n Cute Frieren"
                  className="w-6 h-6 mr-2"
                  draggable="false"
                  onDragStart={(e) => e.preventDefault()}
                />
              </a>
              <h2 className="text-xl font-[Georgia]">Nani ga Suki?</h2>
            </div>

            <button
              onClick={onToggle}
              className="text-[#DBDBDB] hover:text-white"
            >
              <svg viewBox="0 0 24 24" fill="currentColor" className="w-6 h-6">
                <path
                  d="M6 18L18 6M6 6l12 12"
                  stroke="currentColor"
                  strokeWidth="2"
                />
              </svg>
            </button>
          </div>

          {/* Separator line */}
          <div className="mx-4 h-px bg-[#DBDBDB]/20"></div>

          {/* Search Bar */}
          <div className="p-4">
            <input
              type="text"
              placeholder="Search by name"
              className="w-full p-2 rounded bg-white text-black"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>

          {/* Elements List */}
          <div className="p-4">
            <h3 className="text-center mb-2 text-[#DBDBDB] font-light">
              ~ Elements List ~
            </h3>
            <div className="h-40 overflow-y-auto">
              {loading ? (
                <div className="text-center">Loading elements...</div>
              ) : filteredElements.length === 0 ? (
                <div className="text-center">No elements found</div>
              ) : (
                filteredElements.map((element) => (
                  <div
                    key={element}
                    className={`p-2 mb-1 rounded flex items-center cursor-pointer ${
                      selectedElement === element
                        ? "bg-[#DBDBDB] text-black"
                        : "bg-[#505050] text-[#DBDBDB]"
                    }`}
                    onClick={() => handleElementSelect(element)}
                  >
                    {elementsData[element]?.imageLink && (
                      <Image
                        src={elementsData[element].imageLink}
                        alt={element}
                        width={30}
                        height={30}
                        className="mr-2"
                        crossOrigin="anonymous"
                      />
                    )}
                    <span className="truncate">{element}</span>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* Algorithm Selection */}
          <div className="p-4">
            <h3 className="text-center mb-2 text-[#DBDBDB] font-light">
              ~ Algorithm ~
            </h3>
            <div className="flex justify-between gap-2">
              <button
                className={`flex-1 p-2 rounded ${
                  selectedAlgorithm === "BFS"
                    ? "bg-[#DBDBDB] hover:bg-gray-300 text-black"
                    : "bg-[#505050] text-[#DBDBDB]"
                } ${
                  !selectedElement ? "opacity-70" : "hover:bg-[#606060]"
                } transition-colors`}
                onClick={() => selectedElement && handleAlgorithmSelect("BFS")}
                disabled={!selectedElement}
              >
                BFS
              </button>
              <button
                className={`flex-1 p-2 rounded ${
                  selectedAlgorithm === "DFS"
                    ? "bg-[#DBDBDB] hover:bg-gray-300 text-black"
                    : "bg-[#505050] text-[#DBDBDB]"
                } ${
                  !selectedElement ? "opacity-70" : "hover:bg-[#606060]"
                } transition-colors`}
                onClick={() => selectedElement && handleAlgorithmSelect("DFS")}
                disabled={!selectedElement}
              >
                DFS
              </button>
            </div>
          </div>

          {/* Mode Selection */}
          <div className="p-4">
            <h3 className="text-center mb-2 text-[#DBDBDB] font-light">
              ~ Mode ~
            </h3>
            <button
              className={`w-full p-2 mb-2 rounded ${
                isLiveUpdateMode
                  ? "bg-[#DBDBDB] hover:bg-gray-300 text-black"
                  : "bg-[#505050] text-[#DBDBDB]"
              } ${
                !selectedAlgorithm ? "opacity-70" : "hover:bg-[#606060]"
              } transition-colors`}
              onClick={handleLiveUpdateToggle}
              disabled={!selectedAlgorithm}
            >
              Live Update
            </button>

            {/* Multiple Recipe Section */}
            <div className="flex gap-2 items-center h-10">
              <button
                className={`h-full text-sm flex-grow p-1 rounded ${
                  isMultipleRecipeMode
                    ? "bg-[#DBDBDB] hover:bg-gray-300 text-black"
                    : "bg-[#505050] text-[#DBDBDB]"
                } ${
                  !selectedAlgorithm ? "opacity-70" : "hover:bg-[#606060]"
                } transition-colors`}
                onClick={handleMultipleRecipeToggle}
                disabled={!selectedAlgorithm}
              >
                <span>Multiple Recipe</span>
              </button>

              {/* Recipe Count Input */}
              <input
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                value={isAllSelected ? "?" : recipeCount}
                onChange={handleRecipeCountChange}
                placeholder="1"
                disabled={!isMultipleRecipeMode || isAllSelected}
                className={`w-12 h-full text-center rounded bg-[#555555] text-[#DBDBDB] ${
                  !isMultipleRecipeMode || isAllSelected ? "opacity-70" : ""
                } [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none`}
              />

              {/* All Button */}
              <button
                className={`h-full px-2 rounded ${
                  isAllSelected
                    ? "bg-[#DBDBDB] hover:bg-gray-300 text-black"
                    : "bg-[#505050] text-[#DBDBDB]"
                } ${
                  !isMultipleRecipeMode ? "opacity-70" : "hover:bg-[#606060]"
                } transition-colors`}
                onClick={handleAllClick}
                disabled={!isMultipleRecipeMode}
              >
                All
              </button>
            </div>
          </div>

          {/* Separator line */}
          <div className="mt-auto mx-4 h-px bg-[#DBDBDB]/20"></div>

          {/* Find Button */}
          <div className="p-4">
            <button
              className={`w-full p-2 rounded flex items-center justify-center transition-colors ${
                selectedAlgorithm && !isProcessing
                  ? "bg-[#DBDBDB] hover:bg-gray-300 text-black"
                  : "bg-[#505050] text-[#DBDBDB] opacity-70"
              }`}
              onClick={handleFindClick}
              disabled={!selectedAlgorithm || isProcessing}
            >
              <span>{isProcessing ? "Finding..." : "Find!"}</span>
              {isProcessing && <div className="loading ml-2"></div>}
            </button>
          </div>
        </div>
      </div>
    </>
  );
};

export default Sidebar;
