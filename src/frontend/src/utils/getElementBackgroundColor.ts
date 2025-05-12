export const getElementBackgroundColor = (label: string): string => {
  
  if (label == "Earth") {
    return "#8B45133F"; // Brown with some transparency
  } else if (label == "Air") {
    return "#87CEFA3F"; // Light blue with some transparency
  } else if (label == "Water") {
    return "#1E90FF3F"; // Blue with some transparency
  } else if (label == "Fire") {
    return "#FF45003F"; // Orange-red with some transparency
  }
  
  return "#fff"; // Default white background
};

