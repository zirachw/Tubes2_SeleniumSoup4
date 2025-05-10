"use client";
import React, { useState } from 'react';
import Image from 'next/image';

interface ElementData {
  tier: number;
  imageLink: string;
  recipes: string[][];
}

interface ElementsData {
  [elementName: string]: ElementData;
}

interface SidebarProps {
  isOpen: boolean;
  onToggle: () => void;
  elementsData: ElementsData;
  loading: boolean;
}

const Sidebar: React.FC<SidebarProps> = ({ isOpen, onToggle, elementsData, loading }) => {
  // Search and selection states
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedElement, setSelectedElement] = useState<string | null>(null);
  const [selectedAlgorithm, setSelectedAlgorithm] = useState<string | null>(null);
  
  // Mode states
  const [isLiveUpdateMode, setIsLiveUpdateMode] = useState(false);
  const [isMultipleRecipeMode, setIsMultipleRecipeMode] = useState(false);
  
  // Counter states
  const [recipeCount, setRecipeCount] = useState(1);
  const [isAllSelected, setIsAllSelected] = useState(false);

  // Filter elements based on search term
  const filteredElements = Object.keys(elementsData).filter(
    (element) => element.toLowerCase().includes(searchTerm.toLowerCase())
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
        setRecipeCount(1);
      }
    }
  };

  // Handle recipe count change
  const handleRecipeCountChange = (increment: boolean) => {
    if (isMultipleRecipeMode && !isAllSelected) {
      if (increment) {
        setRecipeCount(recipeCount + 1);
      } else if (recipeCount > 1) {
        setRecipeCount(recipeCount - 1);
      }
    }
  };

  // Handle "All" button click
  const handleAllClick = () => {
    if (isMultipleRecipeMode) {
      setIsAllSelected(!isAllSelected);
      if (!isAllSelected) {
        setRecipeCount(0); // Disable counter when "All" is selected
      } else {
        setRecipeCount(1); // Reset to 1 when deselecting "All"
      }
    }
  };

  // Handle find button click
  const handleFindClick = () => {
    // This function would trigger the search logic
    console.log('Finding recipes with:', {
      element: selectedElement,
      algorithm: selectedAlgorithm,
      multipleRecipes: isMultipleRecipeMode,
      liveUpdate: isLiveUpdateMode,
      count: isAllSelected ? 'all' : recipeCount
    });
  };

  return (
    <div className={`fixed top-0 left-0 h-full transition-all duration-300 z-10 ${
      isOpen ? 'w-64' : 'w-0'
    }`}>
      <div className="h-full w-64 overflow-y-auto bg-gradient-to-r from-[#303030] to-[#535353] text-white flex flex-col">
        {/* Header with close button */}
        <div className="p-4 flex justify-between items-center border-b border-gray-700">
          <h2 className="text-xl font-bold">Nani ga Suki?</h2>
          <button onClick={onToggle} className="text-white">
            <svg viewBox="0 0 24 24" fill="white" className="w-6 h-6">
              <path d="M6 18L18 6M6 6l12 12" stroke="white" strokeWidth="2"/>
            </svg>
          </button>
        </div>

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
          <h3 className="text-center mb-2">~ Elements List ~</h3>
          <div className="max-h-40 overflow-y-auto">
            {loading ? (
              <div className="text-center">Loading elements...</div>
            ) : (
              filteredElements.map((element) => (
                <div
                  key={element}
                  className={`p-2 mb-1 rounded flex items-center cursor-pointer ${
                    selectedElement === element ? 'bg-yellow-500 text-black' : 'bg-gray-200 text-gray-800'
                  }`}
                  onClick={() => handleElementSelect(element)}
                >
                  {elementsData[element]?.imageLink && (
                    <Image
                        src = {elementsData[element].imageLink}
                        alt = {element}
                        width = {30}
                        height = {30}
                        crossOrigin = "anonymous"
                    ></Image>
                  )}
                  <span>{element}</span>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Algorithm Selection */}
        <div className="p-4">
          <h3 className="text-center mb-2">~ Algorithm ~</h3>
          <div className="flex justify-between gap-2">
            <button
              className={`flex-1 p-2 rounded ${
                selectedAlgorithm === 'BFS' 
                  ? 'bg-yellow-500 text-black' 
                  : 'bg-gray-300 text-gray-800'
              } ${!selectedElement ? 'opacity-70 cursor-not-allowed' : ''}`}
              onClick={() => selectedElement && handleAlgorithmSelect('BFS')}
              disabled={!selectedElement}
            >
              BFS
            </button>
            <button
              className={`flex-1 p-2 rounded ${
                selectedAlgorithm === 'DFS' 
                  ? 'bg-yellow-500 text-black' 
                  : 'bg-gray-300 text-gray-800'
              } ${!selectedElement ? 'opacity-70 cursor-not-allowed' : ''}`}
              onClick={() => selectedElement && handleAlgorithmSelect('DFS')}
              disabled={!selectedElement}
            >
              DFS
            </button>
          </div>
        </div>

        {/* Mode Selection */}
        <div className="p-4">
          <h3 className="text-center mb-2">~ Mode ~</h3>
          <button
            className={`w-full p-2 mb-2 rounded ${
              isLiveUpdateMode 
                ? 'bg-yellow-500 text-black' 
                : 'bg-gray-300 text-gray-800'
            } ${!selectedAlgorithm ? 'opacity-70 cursor-not-allowed' : ''}`}
            onClick={handleLiveUpdateToggle}
            disabled={!selectedAlgorithm}
          >
            Live Update
          </button>

          {/* Multiple Recipe Section */}
          <div className="flex gap-2 items-center">
            <button
              className={`flex-grow p-2 rounded ${
                isMultipleRecipeMode 
                  ? 'bg-yellow-500 text-black' 
                  : 'bg-gray-300 text-gray-800'
              } ${!selectedAlgorithm ? 'opacity-70 cursor-not-allowed' : ''}`}
              onClick={handleMultipleRecipeToggle}
              disabled={!selectedAlgorithm}
            >
              Multiple Recipe
            </button>

            {/* Counter Controls */}
            <div className={`flex ${!isMultipleRecipeMode ? 'opacity-70' : ''}`}>
              <button
                className="px-3 py-1 rounded-l bg-gray-400 text-black"
                onClick={() => handleRecipeCountChange(false)}
                disabled={!isMultipleRecipeMode || isAllSelected}
              >
                -
              </button>
              <div className="px-3 py-1 bg-gray-200 text-black">
                {isAllSelected ? '' : recipeCount}
              </div>
              <button
                className="px-3 py-1 rounded-r bg-gray-400 text-black"
                onClick={() => handleRecipeCountChange(true)}
                disabled={!isMultipleRecipeMode || isAllSelected}
              >
                +
              </button>
            </div>

            {/* All Button */}
            <button
              className={`py-1 px-3 rounded ${
                isAllSelected 
                  ? 'bg-yellow-500 text-black' 
                  : 'bg-gray-300 text-gray-800'
              } ${!isMultipleRecipeMode ? 'opacity-70 cursor-not-allowed' : ''}`}
              onClick={handleAllClick}
              disabled={!isMultipleRecipeMode}
            >
              All
            </button>
          </div>
        </div>

        {/* Find Button */}
        <div className="mt-auto p-4">
          <button
            className={`w-full p-2 rounded bg-gray-300 text-gray-800 ${
              !selectedAlgorithm ? 'opacity-70 cursor-not-allowed' : ''
            }`}
            onClick={handleFindClick}
            disabled={!selectedAlgorithm}
          >
            Find !
          </button>
        </div>
      </div>
    </div>
  );
};

export default Sidebar;